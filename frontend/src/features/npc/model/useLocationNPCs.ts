import { useCallback, useEffect, useRef, useState } from 'react'
import { useGameActor } from '../../../entities/account/useGameActor'
import { apiError } from '../../../shared/api/errors'
import { locationNPCs } from '../api'

export function useLocationNPCs(accountId: number, characterId: number, locationId: number, onContextChanged: () => void) {
  const { current, report } = useGameActor(accountId, characterId)
  const [data, setData] = useState<Awaited<ReturnType<typeof locationNPCs>> | null>(null)
  const [error, setError] = useState('')
  const [attempt, setAttempt] = useState(0)
  const read = useRef<AbortController | null>(null)
  const reload = useCallback(() => {
    if (!current()) return
    read.current?.abort()
    setData(null); setError(''); setAttempt(value => value + 1)
  }, [current])
  useEffect(() => {
    const controller = new AbortController()
    read.current = controller
    const valid = () => !controller.signal.aborted && read.current === controller && current()
    locationNPCs(accountId, characterId, locationId, controller.signal)
      .then(value => { if (valid()) setData(value) })
      .catch(cause => {
        if (!valid()) return
        setData(null)
        report(cause, setError)
        if (['position_changed', 'world_access_denied', 'planet_changed'].includes(apiError(cause).code ?? '')) onContextChanged()
      })
    return () => controller.abort()
  }, [accountId, characterId, locationId, attempt, current, report, onContextChanged])
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') reload() }
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => { window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh) }
  }, [reload])
  return { data, error, reload }
}
