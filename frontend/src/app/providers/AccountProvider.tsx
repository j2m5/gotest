import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { currentUser, login, logout, resetPassword } from '../../features/auth/api'
import { characterCatalog, createCharacter, listCharacters, setActiveCharacter } from '../../features/character/api'
import { connectAuthSync } from '../../entities/account/authSync'
import { AccountStore } from '../../entities/account/store'
import { AccountContext } from '../../entities/account/context'
export default function AccountProvider({ children }: { children: ReactNode }) {
  const [store] = useState(() => new AccountStore({
    currentUser, login, logout, resetPassword, listCharacters, setActiveCharacter, createCharacter, characterCatalog,
  }))
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') void store.refresh() }
    const releaseSync = connectAuthSync(store)
    void store.refresh()
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    const timer = window.setInterval(refresh, 15000)
    return () => {
      releaseSync()
      store.suspend()
      window.removeEventListener('focus', refresh)
      document.removeEventListener('visibilitychange', refresh)
      window.clearInterval(timer)
    }
  }, [store])
  return <AccountContext value={store}>{children}</AccountContext>
}
