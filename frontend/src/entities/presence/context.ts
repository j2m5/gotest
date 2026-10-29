import { createContext, useContext, useSyncExternalStore } from 'react'
import type { PresenceStore } from './store'

export const PresenceContext = createContext<PresenceStore | null>(null)
const empty = { total: null, statuses: {} as Record<number, boolean | undefined>, error: '' }
const snapshot = () => empty
const subscribe = () => () => {}
export function usePresence() {
  const store = useContext(PresenceContext)
  const state = useSyncExternalStore(store?.subscribe ?? subscribe, store?.getSnapshot ?? snapshot)
  return { ...state, store }
}
