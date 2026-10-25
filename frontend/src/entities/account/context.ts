import { createContext, useContext, useSyncExternalStore } from 'react'
import type { AccountStore } from './store'
export const AccountContext = createContext<AccountStore | null>(null)
export function useAccount() {
  const store = useContext(AccountContext)
  if (!store) throw new Error('AccountProvider is required')
  return { ...useSyncExternalStore(store.subscribe, store.getSnapshot), store }
}
