import { apiError } from '../../../shared/api/errors.ts'
import type { APIError } from '../../../shared/api/errors.ts'

interface SessionStore {
  getSnapshot(): { user: { id: number } | null; roster: { active_character_id: number | null } | null }
  refresh(): Promise<void>
  externalAuthChanged(): Promise<void>
}
// Every operation belongs to the account AND character that displayed its form.
export class SocialSession {
  private active = false
  private controllers = new Set<AbortController>()
  private store: SessionStore
  private accountId: number
  private characterId: number | null
  constructor(store: SessionStore, accountId: number, characterId: number | null) {
    this.store = store; this.accountId = accountId; this.characterId = characterId
  }
  start() { this.active = true }
  stop() { this.active = false; this.controllers.forEach(controller => controller.abort()); this.controllers.clear() }
  current = () => {
    const snapshot = this.store.getSnapshot()
    return this.active && snapshot.user?.id === this.accountId && (snapshot.roster?.active_character_id ?? null) === this.characterId
  }
  report = (error: unknown, fail: (error: APIError) => void) => {
    if (!this.current()) return
    const failure = apiError(error)
    fail(failure)
    if (failure.code === 'account_changed') void this.store.externalAuthChanged()
    else if (failure.status === 401 || failure.code === 'character_changed' || failure.code === 'character_required' || failure.code === 'forbidden') void this.store.refresh()
  }
  run = async <T,>(operation: (signal: AbortSignal) => Promise<T>, apply: (value: T) => void, fail: (error: APIError) => void) => {
    if (!this.current()) return
    const controller = new AbortController()
    this.controllers.add(controller)
    try {
      const result = await operation(controller.signal)
      if (!controller.signal.aborted && this.current()) apply(result)
    } catch (error) { if (!controller.signal.aborted) this.report(error, fail) }
    finally { this.controllers.delete(controller) }
  }
}
