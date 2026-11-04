import { createContext, useContext } from 'react'
import type { useWalletBalance } from './useWalletBalance'

type WalletContextValue = ReturnType<typeof useWalletBalance> & { accountId: number; characterId: number }
export const WalletContext = createContext<WalletContextValue | null>(null)
export function useWallet() {
  const value = useContext(WalletContext)
  if (!value) throw new Error('WalletProvider is required')
  return value
}
