import type { AccountStore } from '../account/store.ts'
import type { PresenceStore } from './store.ts'
import { apiError } from '../../shared/api/errors.ts'

export function connectPresenceAccount(
  accountStore: Pick<AccountStore, 'getSnapshot' | 'subscribe' | 'refresh' | 'externalAuthChanged'>,
  store: PresenceStore,
  visible: () => boolean,
  now = () => performance.now(),
) {
  let activeKey = '', lastReadyKey = ''
  let lastActivity = visible() ? now() : -Infinity
  let pendingInput = visible()
  const activity = (foreground = false) => {
    lastActivity = now(); pendingInput = !activeKey
    store.activity(foreground)
  }
  const sync = () => {
    const account = accountStore.getSnapshot()
    const id = account.status === 'ready' ? account.user?.id : undefined
    const character = account.roster?.active_character_id ?? null
    const next = id ? `${id}:${character}` : ''
    if (activeKey === next) return
    activeKey = next
    if (!id) { store.stop(); return }
    store.start(id, character, error => {
      if (apiError(error).code === 'account_changed') void accountStore.externalAuthChanged()
      else void accountStore.refresh()
    })
    // Account polling/error recovery may restart reads; only real recent input
    // or the initial visible page opening may submit activity for the new context.
    if (visible() && now() - lastActivity < 15000 && (pendingInput || next !== lastReadyKey)) store.activity(true)
    pendingInput = false; lastReadyKey = next
  }
  const release = accountStore.subscribe(sync)
  sync()
  return { activity, stop: () => { release(); store.stop() } }
}
