import { useCallback } from 'react'
import { useAccount } from './context'
import { apiError } from '../../shared/api/errors'

export function useGameActor(accountId: number, characterId: number) {
  const { store } = useAccount()
  const current = useCallback(() => {
    const state = store.getSnapshot()
    return state.user?.id === accountId && state.roster?.active_character_id === characterId
  }, [store, accountId, characterId])
  const report = useCallback((cause: unknown, apply: (message: string) => void) => {
    if (!current()) return
    const failure = apiError(cause)
    apply(failure.message)
    if (failure.code === 'account_changed') void store.externalAuthChanged()
    else if (failure.status === 401 || failure.code === 'character_changed' || failure.code === 'character_required') void store.refresh()
  }, [store, current])
  return { store, current, report }
}
