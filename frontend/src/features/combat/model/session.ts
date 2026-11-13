import { apiError } from '../../../shared/api/errors.ts'
import type { CombatAPI, CombatScope, CombatSnapshot, Command, BattleResult, BattleDetail, BattleEvent, Character } from './types.ts'

export class CombatSession {
  private api: CombatAPI
  private scope: CombatScope
  private now: () => number
  private newKey: () => string
  private active = false
  private generation = 0
  private battleId: string | null = null
  private read: AbortController | null = null
  private write: AbortController | null = null
  private again = false
  private command: Command | null = null
  private listeners = new Set<() => void>()
  private state: CombatSnapshot = { current: null, character: null, detail: null, loadout: null, events: [],
    receivedAt: 0, loading: true, error: '', actionError: '', pending: null, unconfirmed: false }
  constructor(api: CombatAPI, scope: CombatScope, options: { now?: () => number; newKey?: () => string } = {}) {
    this.api = api; this.scope = scope
    this.now = options.now ?? (() => performance.now()); this.newKey = options.newKey ?? (() => crypto.randomUUID())
  }
  getSnapshot = () => this.state
  subscribe = (listener: () => void) => { this.listeners.add(listener); return () => { this.listeners.delete(listener) } }
  private update(patch: Partial<CombatSnapshot>) {
    this.state = { ...this.state, ...patch }; this.listeners.forEach(listener => listener())
  }
  private current = () => this.active && this.scope.current()
  start() { if (!this.active) { this.active = true; void this.refresh() } }
  stop() {
    this.active = false; ++this.generation; this.read?.abort(); this.write?.abort()
    this.read = null; this.write = null; this.again = false; this.command = null
    this.update({ pending: null, unconfirmed: false })
  }
  watch = (id: string | null) => {
    if (id === this.battleId) return
    this.battleId = id; ++this.generation; this.read?.abort(); this.write?.abort()
    this.read = null; this.write = null; this.command = null; this.again = false
    this.update({ detail: null, events: [], loadout: null, loading: true, error: '', actionError: '', pending: null, unconfirmed: false })
    void this.refresh()
  }
  private async journal(id: string, signal: AbortSignal): Promise<{ detail: BattleDetail; events: BattleEvent[] }> {
    const existing = this.state.detail?.battle.id === id ? this.state.events : []
    const events = new Map(existing.map(event => [event.sequence, event]))
    let after = existing.at(-1)?.sequence ?? 0
    for (let page = 0; page < 100; page++) {
      const detail = await this.api.detail(id, after, signal)
      if (signal.aborted) throw new Error('Read cancelled')
      if (detail.log_cleared_at) return { detail, events: detail.events }
      for (const event of detail.events) events.set(event.sequence, event)
      if (detail.next_after === null) return { detail, events: [...events.values()].sort((a, b) => a.sequence - b.sequence) }
      if (detail.next_after <= after) throw new Error('Некорректная последовательность журнала')
      after = detail.next_after
    }
    throw new Error('Журнал слишком большой. Обновите страницу.')
  }
  refresh = async () => {
    if (!this.current()) return
    if (this.read || this.write) { this.again = true; return }
    const controller = new AbortController(), generation = this.generation, id = this.battleId
    this.read = controller
    const valid = () => this.current() && !controller.signal.aborted && this.read === controller && generation === this.generation
    try {
      const [current, journal, abilities] = await Promise.all([
        this.api.current(controller.signal),
        id ? this.journal(id, controller.signal) : Promise.resolve(null),
        id ? this.api.abilities(controller.signal).then(loadout => ({ loadout, at: this.now() })) : Promise.resolve(null),
      ])
      if (!valid()) return
      const characters: Character[] = [current.character]
      if (journal) characters.push(journal.detail.character)
      if (abilities) characters.push(abilities.loadout.character)
      characters.sort((a,b) => Date.parse(b.regeneration.observed_at) - Date.parse(a.regeneration.observed_at))
      this.update({ current, character: characters[0], detail: journal?.detail ?? null, events: journal?.events ?? [],
        loadout: abilities?.loadout ?? null, receivedAt: abilities?.at ?? this.now(), loading: false, error: '' })
    } catch (cause) {
      if (!valid()) return
      const failure = apiError(cause)
      this.update({ loading: false, error: failure.message })
      this.scope.report(failure)
    } finally {
      if (this.read === controller) {
        this.read = null
        if (this.again) { this.again = false; void this.refresh() }
      }
    }
  }
  begin = (locationId: number, npcId: number) => this.perform({ kind: 'start', locationId, npcId, key: this.newKey() })
  use = (abilityId: number) => {
    if (!this.battleId || this.state.detail?.battle.status !== 'active' || !this.state.loadout || this.state.loading || this.state.error) return Promise.resolve(null)
    return this.perform({ kind: 'use', battleId: this.battleId, abilityId, key: this.newKey() })
  }
  retreat = () => {
    if (!this.battleId || this.state.detail?.battle.status !== 'active') return Promise.resolve(null)
    return this.perform({ kind: 'retreat', battleId: this.battleId })
  }
  retry = () => this.command && this.state.unconfirmed ? this.perform(this.command, true) : Promise.resolve(null)
  private async perform(command: Command, retry = false): Promise<BattleResult | null> {
    if (!this.current() || this.write || this.state.unconfirmed && !retry) return null
    this.read?.abort(); this.read = null; this.again = false
    const generation = ++this.generation, controller = new AbortController()
    this.write = controller; this.command = command
    this.update({ pending: command.kind, unconfirmed: false, actionError: '', loading: true })
    const valid = () => this.current() && generation === this.generation && !controller.signal.aborted && this.write === controller
    try {
      const result = command.kind === 'start' ? await this.api.begin(command.locationId, command.npcId, command.key, controller.signal)
        : command.kind === 'use' ? await this.api.use(command.battleId, command.abilityId, command.key, controller.signal)
        : await this.api.retreat(command.battleId, controller.signal)
      if (!valid()) return null
      this.command = null
      this.update({ character: result.character, current: { character: result.character, battle: result.battle.status === 'active' ? result.battle : null },
        unconfirmed: false, actionError: '' })
      return result
    } catch (cause) {
      if (!valid()) return null
      const failure = apiError(cause)
      const uncertain = !failure.status || failure.status >= 500
      if (!uncertain) this.command = null
      this.update({ actionError: failure.message, unconfirmed: uncertain })
      this.scope.report(failure)
      return null
    } finally {
      if (valid()) { this.write = null; this.update({ pending: null }); void this.refresh() }
    }
  }
}
