import { useEffect, useState } from 'react'
import { useAccount } from '../../../entities/account/context'
import type { CharacterCatalog } from '../../../entities/character/types'
import { apiError } from '../../../shared/api/errors'

export function useCharacterCatalog() {
  const { store } = useAccount()
  const [catalog, setCatalog] = useState<CharacterCatalog | null>(null)
  const [error, setError] = useState('')
  const [attempt, setAttempt] = useState(0)
  useEffect(() => {
    let active = true
    store.catalog().then(value => { if (active) setCatalog(value) })
      .catch(error => { if (active) setError(apiError(error).message) })
    return () => { active = false }
  }, [store, attempt])
  const retry = () => { setError(''); setAttempt(value => value + 1) }
  return { catalog, error, retry }
}
