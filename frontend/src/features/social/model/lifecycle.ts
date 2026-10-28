// Owns the displayed cursor. Notifications during a read request one further snapshot.
export class SnapshotFeed<T> {
  private cursor: number | null = null
  private controller: AbortController | null = null
  private again = false
  private stopped = false
  constructor(
    privateRead: (cursor: number | null, signal: AbortSignal) => Promise<T>,
    privateApply: (value: T) => void,
    privateError: (error: unknown) => void,
  ) { this.read = privateRead; this.apply = privateApply; this.error = privateError }
  private read: (cursor: number | null, signal: AbortSignal) => Promise<T>
  private apply: (value: T) => void
  private error: (error: unknown) => void
  page(cursor: number | null) {
    if (this.stopped) return
    this.controller?.abort(); this.controller = null; this.again = false
    this.cursor = cursor; this.refresh()
  }
  refresh = () => {
    if (this.stopped) return
    if (this.controller) { this.again = true; return }
    const controller = new AbortController()
    this.controller = controller
    void this.read(this.cursor, controller.signal).then(value => {
      if (!controller.signal.aborted && !this.stopped) this.apply(value)
    }, error => {
      if (!controller.signal.aborted && !this.stopped) this.error(error)
    }).finally(() => {
      if (this.controller !== controller) return
      this.controller = null
      if (this.again) { this.again = false; this.refresh() }
    })
  }
  stop() { this.stopped = true; this.controller?.abort(); this.controller = null }
}

export type ConnectionStatus = 'connecting' | 'connected' | 'reconnecting'
interface Socket {
  onopen: (() => void) | null
  onmessage: ((event: { data: string }) => void) | null
  onclose: (() => void) | null
  onerror: (() => void) | null
  close(): void
}
export function connectChat<Timer = ReturnType<typeof setTimeout>>(options: {
  accountId: number; origin: string
  connect?: (url: string) => Socket
  refresh: () => void; status: (status: ConnectionStatus) => void
  schedule?: (callback: () => void, delay: number) => Timer
  cancel?: (timer: Timer) => void
}) {
  const connect = options.connect ?? (url => new WebSocket(url) as unknown as Socket)
  const schedule = options.schedule ?? ((callback, delay) => setTimeout(callback, delay) as Timer)
  const cancel = options.cancel ?? (timer => clearTimeout(timer as ReturnType<typeof setTimeout>))
  let socket: Socket | null = null, timer: Timer | undefined, stopped = false, attempt = 0, generation = 0
  const open = () => {
    if (stopped) return
    const context = ++generation
    const current = () => !stopped && context === generation
    timer = undefined
    options.status(attempt ? 'reconnecting' : 'connecting')
    const url = new URL('/api/chat/events', options.origin)
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    url.searchParams.set('account_id', String(options.accountId))
    const retry = () => {
      if (!current() || timer !== undefined) return
      options.status('reconnecting'); options.refresh()
      timer = schedule(open, Math.min(1000 * 2 ** attempt++, 15000))
    }
    try {
      const opened = connect(url.href)
      socket = opened
      opened.onopen = () => { if (current()) { attempt = 0; options.status('connected'); options.refresh() } }
      opened.onmessage = event => {
        if (!current()) return
        try { if (JSON.parse(event.data)?.type === 'changed') options.refresh() } catch { /* Ignore malformed notifications. */ }
      }
      opened.onclose = retry
      opened.onerror = () => { if (current()) opened.close() }
    } catch { retry() }
  }
  open()
  return {
    retry() { if (stopped) return; if (timer !== undefined) cancel(timer); timer = undefined; if (socket) { socket.onclose = null; socket.close() }; open() },
    stop() { stopped = true; if (timer !== undefined) cancel(timer); socket?.close() },
  }
}
