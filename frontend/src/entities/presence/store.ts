import { apiError } from '../../shared/api/errors.ts'

export interface PresenceAPI {
  heartbeat(account: number, character: number | null, signal: AbortSignal): Promise<unknown>
  statuses(account: number, ids: number[], signal: AbortSignal): Promise<{ total: number; characters: { character_id: number; is_online: boolean }[] }>
}
interface Clock {
  now(): number
  schedule(callback: () => void, delay: number): ReturnType<typeof setTimeout>
  cancel(timer: ReturnType<typeof setTimeout>): void
}
interface Snapshot { total: number | null; statuses: Record<number, boolean | undefined>; error: string }
const browserClock: Clock = { now: () => performance.now(), schedule: (fn, delay) => setTimeout(fn, delay), cancel: timer => clearTimeout(timer) }

// Reading presence and submitting player activity deliberately have separate paths.
export class PresenceStore {
  private api: PresenceAPI
  private clock: Clock
  private generation = 0
  private account: number | null = null
  private character: number | null = null
  private controllers = new Set<AbortController>()
  private listeners = new Set<() => void>()
  private watchers = new Map<number, number>()
  private requestedIds = new Set<number>()
  private poll: ReturnType<typeof setTimeout> | null = null
  private trailing: ReturnType<typeof setTimeout> | null = null
  private lastWrite = -Infinity
  private lastReturn = -Infinity
  private lastActivityRead = -Infinity
  private observedAt = -Infinity
  private writing = false
  private pendingActivity = false
  private pendingActivityAt = -Infinity
  private reading = false
  private queuedRead = false
  private authReported = false
  private onConflict: (error: unknown) => void = () => {}
  private state: Snapshot = { total: null, statuses: {}, error: '' }
  constructor(api: PresenceAPI, clock: Clock = browserClock) { this.api = api; this.clock = clock }
  getSnapshot = () => this.state
  subscribe = (listener: () => void) => { this.listeners.add(listener); return () => { this.listeners.delete(listener) } }
  private update(state: Snapshot) { this.state = state; this.listeners.forEach(listener => listener()) }
  start(account: number, character: number | null, onConflict: (error: unknown) => void = () => {}) {
    this.stop()
    this.account = account; this.character = character; this.onConflict = onConflict
    this.refresh(); this.schedulePoll()
  }
  stop() {
    this.generation++; this.account = null
    this.controllers.forEach(controller => controller.abort()); this.controllers.clear()
    if (this.poll !== null) this.clock.cancel(this.poll)
    if (this.trailing !== null) this.clock.cancel(this.trailing)
    this.poll = null; this.trailing = null; this.writing = false; this.reading = false; this.queuedRead = false
    this.pendingActivity = false; this.pendingActivityAt = -Infinity; this.authReported = false; this.lastWrite = -Infinity; this.lastReturn = -Infinity; this.lastActivityRead = -Infinity; this.observedAt = -Infinity
    this.update({ total: null, statuses: {}, error: '' })
  }
  watch(ids: readonly number[]) {
    const unique = [...new Set(ids.filter(id => Number.isInteger(id) && id > 0))]
    unique.forEach(id => this.watchers.set(id, (this.watchers.get(id) ?? 0) + 1))
    // All indicators mounted in the same render share a single batched read.
    const generation = this.generation
    if (this.account !== null) queueMicrotask(() => {
      if (generation === this.generation && unique.some(id => !this.requestedIds.has(id))) this.refresh()
    })
    return () => { unique.forEach(id => { const count = (this.watchers.get(id) ?? 1) - 1; if (count) this.watchers.set(id, count); else this.watchers.delete(id) }) }
  }
  private schedulePoll() {
    this.poll = this.clock.schedule(() => { this.poll = null; this.refresh(); if (this.account !== null) this.schedulePoll() }, 30000)
  }
  private expired() {
    if (this.clock.now() - this.observedAt >= 120000 && (this.state.total !== null || Object.keys(this.state.statuses).length)) {
      this.update({ total: null, statuses: {}, error: '' })
    }
  }
  private report(error: unknown) {
    const failure = apiError(error)
    if (!this.authReported && (failure.status === 401 || failure.code === 'account_changed' || failure.code === 'character_changed')) {
      this.authReported = true; this.onConflict(error)
    }
  }
  refresh = () => {
    if (this.account === null) return
    this.expired()
    if (this.reading) { this.queuedRead = true; return }
    this.reading = true
    const generation = this.generation, account = this.account, started = this.clock.now()
    const ids = [...this.watchers.keys()].sort((a, b) => a - b)
    this.requestedIds = new Set(ids)
    const chunks: number[][] = []
    for (let i = 0; i < ids.length; i += 100) chunks.push(ids.slice(i, i + 100))
    if (!chunks.length) chunks.push([])
    const controller = new AbortController(); this.controllers.add(controller)
    void Promise.all(chunks.map(chunk => this.api.statuses(account, chunk, controller.signal))).then(results => {
      if (generation !== this.generation || controller.signal.aborted) return
      if (this.clock.now() - started >= 120000) { this.update({ total: null, statuses: {}, error: '' }); return }
      this.observedAt = started
      const statuses: Snapshot['statuses'] = {}
      results.forEach(result => result.characters.forEach(item => { statuses[item.character_id] = item.is_online }))
      this.update({ total: results[0].total, statuses, error: '' })
    }, error => {
      if (generation !== this.generation || controller.signal.aborted) return
      this.update({ total: null, statuses: {}, error: apiError(error).message }); this.report(error)
    }).finally(() => {
      this.controllers.delete(controller)
      if (generation !== this.generation) return
      this.reading = false
      if (this.queuedRead) { this.queuedRead = false; this.refresh() }
    })
  }
  activity = (foregroundReturn = false) => {
    if (this.account === null) return
    this.expired()
    const now = this.clock.now()
    // Browsers commonly dispatch both focus and visibilitychange for one return.
    if (foregroundReturn && now - this.lastReturn < 1000) return
    if (foregroundReturn) this.lastReturn = now
    this.pendingActivity = true
    this.pendingActivityAt = now
    this.flushActivity(foregroundReturn)
    if (foregroundReturn || now - this.lastActivityRead >= 15000) { this.lastActivityRead = now; this.refresh() }
  }
  private flushActivity(immediate = false) {
    // Frozen tabs can resume timers or promise continuations long after the input.
    // Only a fresh action may renew presence; a delayed callback is not activity.
    if (this.pendingActivity && this.clock.now() - this.pendingActivityAt >= 120000) {
      this.pendingActivity = false
      if (this.trailing !== null) this.clock.cancel(this.trailing)
      this.trailing = null
    }
    if (!this.pendingActivity || this.writing || this.account === null || this.character === null) return
    const delay = this.lastWrite + 15000 - this.clock.now()
    if (!immediate && delay > 0) {
      if (this.trailing === null) this.trailing = this.clock.schedule(() => { this.trailing = null; this.flushActivity() }, delay)
      return
    }
    if (this.trailing !== null) this.clock.cancel(this.trailing)
    this.trailing = null; this.pendingActivity = false; this.writing = true; this.lastWrite = this.clock.now()
    const generation = this.generation, controller = new AbortController(); this.controllers.add(controller)
    void this.api.heartbeat(this.account, this.character, controller.signal).then(() => {
      if (generation === this.generation && !controller.signal.aborted) this.refresh()
    }, error => {
      if (generation !== this.generation || controller.signal.aborted) return
      this.update({ total: null, statuses: {}, error: apiError(error).message }); this.report(error)
    }).finally(() => {
      this.controllers.delete(controller)
      if (generation !== this.generation) return
      this.writing = false; this.flushActivity()
    })
  }
}
