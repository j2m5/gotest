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

export { type ConnectionStatus } from '../../../shared/api/notifications.ts'
import { connectNotifications, type NotificationOptions } from '../../../shared/api/notifications.ts'

export function connectChat<Timer = ReturnType<typeof setTimeout>>(options: Omit<NotificationOptions<Timer>, 'path' | 'query'> & { accountId: number }) {
  return connectNotifications({ ...options, path: '/api/chat/events', query: { account_id: options.accountId } })
}
