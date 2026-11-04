import type { ReactNode } from 'react'
import { WalletContext } from '../model/context'
import { useWalletBalance } from '../model/useWalletBalance'

export default function WalletProvider({ accountId, characterId, children }: { accountId: number; characterId: number; children: ReactNode }) {
  const balance = useWalletBalance(accountId, characterId)
  return <WalletContext value={{ accountId, characterId, ...balance }}>{children}</WalletContext>
}
