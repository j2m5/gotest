import type { AccountStore } from './store'

const channelName = 'gotest-auth'
const storageKey = 'gotest-auth-change'
const isChange = (data: unknown) => typeof data === 'object' && data !== null && 'type' in data && data.type === 'auth_changed'

// The payload carries only an invalidation hint. Each tab discovers its session itself.
export function connectAuthSync(store: Pick<AccountStore, 'subscribeAuthChanges' | 'externalAuthChanged'>, host: Window & typeof globalThis = window) {
  let channel: BroadcastChannel | undefined
  let released = false
  const changed = () => { if (!released) void store.externalAuthChanged() }
  try {
    if (host.BroadcastChannel) {
      channel = new host.BroadcastChannel(channelName)
      channel.onmessage = event => { if (isChange(event.data)) changed() }
    }
  } catch { /* Private browser policies can also disable BroadcastChannel. */ }
  const storage = (event: StorageEvent) => {
    if (event.key !== storageKey || !event.newValue) return
    try { if (isChange(JSON.parse(event.newValue))) changed() } catch { /* Ignore unrelated or malformed storage. */ }
  }
  host.addEventListener('storage', storage)
  const unsubscribe = store.subscribeAuthChanges(() => {
    const message = { type: 'auth_changed' }
    if (channel) {
      try { channel.postMessage(message); return } catch { /* Fall back when the channel is unavailable. */ }
    }
    try { host.localStorage.setItem(storageKey, JSON.stringify({ ...message, nonce: `${Date.now()}-${Math.random()}` })) }
    catch { /* Server account preconditions remain authoritative if storage is blocked. */ }
  })
  return () => {
    released = true
    unsubscribe()
    host.removeEventListener('storage', storage)
    if (channel) { channel.onmessage = null; channel.close() }
  }
}
