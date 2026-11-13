import { createContext, useContext } from 'react'
import type { CombatSession } from './session'
import type { CombatSnapshot } from './types'
import type { ConnectionStatus } from '../../../shared/api/notifications'

export const CombatContext = createContext<{ session: CombatSession; snapshot: CombatSnapshot; connection: ConnectionStatus } | null>(null)
export function useCombat() {
  const value = useContext(CombatContext)
  if (!value) throw new Error('CombatProvider is required')
  return value
}
