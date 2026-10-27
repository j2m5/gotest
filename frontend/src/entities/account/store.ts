import type { AuthResponse, User } from './types.ts'
import type { Character, CharacterCatalog, CharacterRoster, CreateCharacterInput } from '../character/types.ts'
import { apiError, isUnauthenticated } from '../../shared/api/errors.ts'

interface AccountAPI {
  currentUser(): Promise<AuthResponse>
  listCharacters(): Promise<CharacterRoster>
  login(email: string, password: string, rememberMe: boolean): Promise<AuthResponse>
  logout(): Promise<unknown>
  resetPassword(token: string, password: string): Promise<unknown>
  setActiveCharacter(id: number | null): Promise<{ character: Character | null }>
  createCharacter(input: CreateCharacterInput): Promise<{ character: Character }>
  characterCatalog(): Promise<CharacterCatalog>
}
interface Snapshot {
  status: 'loading' | 'ready' | 'anonymous' | 'error'
  user: User | null
  roster: CharacterRoster | null
  pending: boolean
  error: string
}

export class AccountStore {
  private api: AccountAPI
  private generation = 0
  private inFlightRefresh: { generation: number; promise: Promise<void> } | null = null
  private listeners = new Set<() => void>()
  private state: Snapshot = { status: 'loading', user: null, roster: null, pending: false, error: '' }
  constructor(api: AccountAPI) { this.api = api }
  getSnapshot = () => this.state
  subscribe = (listener: () => void) => { this.listeners.add(listener); return () => { this.listeners.delete(listener) } }
  private update(patch: Partial<Snapshot>) {
    this.state = { ...this.state, ...patch }
    this.listeners.forEach(listener => listener())
  }
  private expire() {
    ++this.generation
    this.update({ status: 'anonymous', user: null, roster: null, pending: false, error: '' })
  }
  async refresh() {
    if (this.state.pending) return
    if (this.inFlightRefresh?.generation === this.generation) return this.inFlightRefresh.promise
    const generation = ++this.generation
    const request = { generation, promise: this.fetchSnapshot(generation) }
    this.inFlightRefresh = request
    try { await request.promise }
    finally {
      if (this.inFlightRefresh === request) this.inFlightRefresh = null
    }
  }
  private async fetchSnapshot(generation: number) {
    try {
      const { user } = await this.api.currentUser()
      if (generation !== this.generation) return
      const roster = await this.api.listCharacters()
      if (generation !== this.generation) return
      this.update({ status: 'ready', user: { ...user, character_id: roster.active_character_id }, roster, error: '' })
    } catch (error) {
      if (generation !== this.generation) return
      if (isUnauthenticated(error)) this.expire()
      else this.update({ status: 'error', roster: null, error: apiError(error).message })
    }
  }
  private async mutate<T>(operation: () => Promise<T>, apply: (result: T) => void) {
    if (this.state.pending) throw new Error('Дождитесь завершения операции.')
    const generation = ++this.generation
    this.update({ pending: true })
    try {
      const result = await operation()
      if (generation !== this.generation) throw new Error('Сеанс изменился. Повторите действие.')
      apply(result)
      return result
    } catch (error) {
      if (generation === this.generation && isUnauthenticated(error)) this.expire()
      throw error
    } finally {
      if (generation === this.generation) { ++this.generation; this.update({ pending: false }) }
    }
  }
  async signIn(email: string, password: string, rememberMe = false) {
    await this.mutate(() => this.api.login(email, password, rememberMe), ({ user }) => {
      this.update({ user, roster: null, status: 'loading', error: '' })
    })
    await this.refresh()
  }
  async signOut() {
    await this.mutate(() => this.api.logout(), () => this.expire())
  }
  async resetPassword(token: string, password: string) {
    await this.mutate(() => this.api.resetPassword(token, password), () => this.expire())
  }
  async select(id: number | null) {
    return this.mutate(() => this.api.setActiveCharacter(id), ({ character }) => {
      if (!this.state.user || !this.state.roster) return
      const active = character?.id ?? null
      const characters = character ? this.state.roster.characters.map(item => item.id === character.id ? character : item) : this.state.roster.characters
      this.update({ user: { ...this.state.user, character_id: active }, roster: { characters, active_character_id: active } })
    })
  }
  async create(input: CreateCharacterInput) {
    return this.mutate(() => this.api.createCharacter(input), ({ character }) => {
      if (this.state.roster) this.update({ roster: { ...this.state.roster, characters: [...this.state.roster.characters, character] } })
    })
  }
  async catalog() {
    const generation = this.generation
    try { return await this.api.characterCatalog() }
    catch (error) {
      if (generation === this.generation && isUnauthenticated(error)) this.expire()
      throw error
    }
  }
}
