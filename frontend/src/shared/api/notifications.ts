export type ConnectionStatus = 'connecting' | 'connected' | 'reconnecting'
interface Socket {
  onopen: (() => void) | null
  onmessage: ((event: { data: string }) => void) | null
  onclose: (() => void) | null
  onerror: (() => void) | null
  close(): void
}
export interface NotificationOptions<Timer = ReturnType<typeof setTimeout>> {
  path: string; query: Record<string, string | number>; origin: string
  connect?: (url: string) => Socket
  refresh: () => void; status: (status: ConnectionStatus) => void
  schedule?: (callback: () => void, delay: number) => Timer
  cancel?: (timer: Timer) => void
}
export function connectNotifications<Timer = ReturnType<typeof setTimeout>>(options: NotificationOptions<Timer>) {
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
    const url = new URL(options.path, options.origin)
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
    for (const [key, value] of Object.entries(options.query)) url.searchParams.set(key, String(value))
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
