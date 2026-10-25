import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import * as api from '../api'
import { AccountStore } from './store'
import { AccountContext } from './context'
export default function AccountProvider({ children }: { children: ReactNode }) {
  const [store] = useState(() => new AccountStore(api))
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') void store.refresh() }
    void store.refresh()
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    const timer = window.setInterval(refresh, 15000)
    return () => {
      window.removeEventListener('focus', refresh)
      document.removeEventListener('visibilitychange', refresh)
      window.clearInterval(timer)
    }
  }, [store])
  return <AccountContext value={store}>{children}</AccountContext>
}
