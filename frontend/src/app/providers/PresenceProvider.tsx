import { useEffect, useState, type ReactNode } from 'react'
import { useAccount } from '../../entities/account/context'
import { PresenceContext } from '../../entities/presence/context'
import { PresenceStore } from '../../entities/presence/store'
import { presenceAPI } from '../../features/presence/api'
import { connectPresenceActivity } from '../../entities/presence/activity'
import { connectPresenceAccount } from '../../entities/presence/account'
import PresenceFooter from '../../entities/presence/PresenceFooter'
import { LayoutFooterContext } from '../../shared/ui/LayoutFooterContext'

export default function PresenceProvider({ children }: { children: ReactNode }) {
  const { store: accountStore } = useAccount()
  const [store] = useState(() => new PresenceStore(presenceAPI))
  useEffect(() => {
    // Subscribe directly: replacement invalidates requests before any late promise
    // continuation, and ordinary account refreshes do not restart activity.
    const account = connectPresenceAccount(accountStore, store, () => document.visibilityState === 'visible')
    const disconnect = connectPresenceActivity(account.activity, window, document)
    return () => {
      account.stop()
      disconnect()
    }
  }, [accountStore, store])
  return <PresenceContext value={store}><LayoutFooterContext value={<PresenceFooter />}>{children}</LayoutFooterContext></PresenceContext>
}
