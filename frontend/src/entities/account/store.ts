import type { AuthResponse, User } from './types.ts'
import type { Character, CharacterCatalog, CharacterRoster, CreateCharacterInput } from '../character/types.ts'
import { apiError, isUnauthenticated } from '../../shared/api/errors.ts'

interface AccountAPI {
  currentUser(signal?: AbortSignal): Promise<AuthResponse>
  listCharacters(accountId: number, signal?: AbortSignal): Promise<CharacterRoster>
  login(email: string, password: string, rememberMe: boolean, signal?: AbortSignal): Promise<AuthResponse>
  logout(accountId: number, signal?: AbortSignal): Promise<unknown>
  resetPassword(token: string, password: string, signal?: AbortSignal): Promise<unknown>
  setActiveCharacter(accountId: number, id: number | null, signal?: AbortSignal): Promise<{ character: Character | null }>
  createCharacter(accountId: number, input: CreateCharacterInput, signal?: AbortSignal): Promise<{ character: Character }>
  characterCatalog(accountId: number, signal?: AbortSignal): Promise<CharacterCatalog>
}
interface Snapshot {
  status: 'loading' | 'ready' | 'anonymous' | 'error'
  user: User | null
  roster: CharacterRoster | null
  pending: boolean
  error: string
}
const sessionChanged = () => new Error('Сеанс изменился. Повторите действие.')
const accountChanged = (error: unknown) => apiError(error).code === 'account_changed'

export class AccountStore {
  private api: AccountAPI
  private generation = 0
  // Catalog reads survive ordinary refreshes, but never account changes or mutations.
  private contextGeneration = 0
  private controllers = new Set<AbortController>()
  private inFlightRefresh: { generation: number; promise: Promise<void> } | null = null
  private listeners = new Set<() => void>()
  private authListeners = new Set<() => void>()
  private state: Snapshot = { status: 'loading', user: null, roster: null, pending: false, error: '' }
  constructor(api: AccountAPI) { this.api = api }
  getSnapshot = () => this.state
  subscribe = (listener: () => void) => { this.listeners.add(listener); return () => { this.listeners.delete(listener) } }
  subscribeAuthChanges = (listener: () => void) => { this.authListeners.add(listener); return () => { this.authListeners.delete(listener) } }
  private broadcast() { this.authListeners.forEach(listener => listener()) }
  private update(patch: Partial<Snapshot>) {
    this.state = { ...this.state, ...patch }
    this.listeners.forEach(listener => listener())
  }
  private invalidate() {
    ++this.generation
    ++this.contextGeneration
    this.controllers.forEach(controller => controller.abort())
    this.controllers.clear()
    this.inFlightRefresh = null
  }
  suspend() { this.invalidate(); this.update({ pending: false }) }
  private expire() {
    this.invalidate()
    this.update({ status: 'anonymous', user: null, roster: null, pending: false, error: '' })
  }
  externalAuthChanged = async () => {
    this.invalidate()
    this.update({ status: 'loading', user: null, roster: null, pending: false, error: '' })
    await this.refresh()
  }
  async refresh() {
    if (this.state.pending) return
    if (this.inFlightRefresh?.generation === this.generation) return this.inFlightRefresh.promise
    const generation = ++this.generation
    const request = { generation, promise: this.fetchSnapshot(generation) }
    this.inFlightRefresh = request
    try { await request.promise }
    finally { if (this.inFlightRefresh === request) this.inFlightRefresh = null }
  }
  private async fetchSnapshot(generation: number) {
    const controller = new AbortController()
    this.controllers.add(controller)
    try {
      const { user } = await this.api.currentUser(controller.signal)
      if (generation !== this.generation) return
      const roster = await this.api.listCharacters(user.id, controller.signal)
      if (generation !== this.generation) return
      if (this.state.user?.id !== user.id) ++this.contextGeneration
      this.update({ status: 'ready', user: { ...user, character_id: roster.active_character_id }, roster, error: '' })
    } catch (error) {
      if (generation !== this.generation) return
      if (accountChanged(error)) await this.externalAuthChanged()
      else if (isUnauthenticated(error)) this.expire()
      else this.update({ status: 'error', roster: null, error: apiError(error).message })
    } finally { this.controllers.delete(controller) }
  }
  private async mutate<T>(operation: (signal: AbortSignal) => Promise<T>, apply: (result: T) => void) {
    if (this.state.pending) throw new Error('Дождитесь завершения операции.')
    this.invalidate()
    const generation = this.generation
    const controller = new AbortController()
    this.controllers.add(controller)
    this.update({ pending: true })
    try {
      const result = await operation(controller.signal)
      if (generation !== this.generation) throw sessionChanged()
      apply(result)
      return result
    } catch (error) {
      if (generation === this.generation) {
        if (accountChanged(error)) await this.externalAuthChanged()
        else if (isUnauthenticated(error)) this.expire()
      }
      throw error
    } finally {
      this.controllers.delete(controller)
      if (generation === this.generation) { ++this.generation; this.update({ pending: false }) }
    }
  }
  private async displayedAccount() {
    if (this.state.user) return this.state.user.id
    if (this.state.status !== 'anonymous') await this.refresh()
    const user = this.getSnapshot().user
    if (!user) throw sessionChanged()
    return user.id
  }
  async signIn(email: string, password: string, rememberMe = false) {
    await this.mutate(signal => this.api.login(email, password, rememberMe, signal), ({ user }) => {
      this.update({ user, roster: null, status: 'loading', error: '' })
      this.broadcast()
    })
    await this.refresh()
  }
  async signOut() {
    if (this.state.status === 'anonymous') return
    // Unknown state discovers the cookie account without waiting for a stalled roster.
    const displayed = this.state.user?.id
    await this.mutate(async signal => {
      const accountId = displayed ?? (await this.api.currentUser(signal)).user.id
      return this.api.logout(accountId, signal)
    }, () => { this.expire(); this.broadcast() })
  }
  async resetPassword(token: string, password: string) {
    await this.mutate(signal => this.api.resetPassword(token, password, signal), () => { this.expire(); this.broadcast() })
  }
  async select(id: number | null) {
    const before = this.generation
    const accountId = this.state.user?.id ?? await this.displayedAccount()
    if (before !== this.generation && this.state.user?.id !== accountId) throw sessionChanged()
    return this.mutate(signal => this.api.setActiveCharacter(accountId, id, signal), ({ character }) => {
      if (!this.state.user || !this.state.roster) return
      const active = character?.id ?? null
      const characters = character ? this.state.roster.characters.map(item => item.id === character.id ? character : item) : this.state.roster.characters
      this.update({ user: { ...this.state.user, character_id: active }, roster: { characters, active_character_id: active } })
    })
  }
  async create(input: CreateCharacterInput) {
    const before = this.generation
    const accountId = this.state.user?.id ?? await this.displayedAccount()
    if (before !== this.generation && this.state.user?.id !== accountId) throw sessionChanged()
    return this.mutate(signal => this.api.createCharacter(accountId, input, signal), ({ character }) => {
      if (this.state.roster) this.update({ roster: { ...this.state.roster, characters: [...this.state.roster.characters, character] } })
    })
  }
  async catalog() {
    const before = this.generation
    const accountId = this.state.user?.id ?? await this.displayedAccount()
    if (before !== this.generation && this.state.user?.id !== accountId) throw sessionChanged()
    const generation = this.contextGeneration
    const currentContext = () => generation === this.contextGeneration && this.state.user?.id === accountId
    const controller = new AbortController()
    this.controllers.add(controller)
    try {
      const result = await this.api.characterCatalog(accountId, controller.signal)
      if (!currentContext()) throw sessionChanged()
      return result
    } catch (error) {
      if (currentContext()) {
        if (accountChanged(error)) await this.externalAuthChanged()
        else if (isUnauthenticated(error)) this.expire()
      }
      throw error
    } finally { this.controllers.delete(controller) }
  }
}
