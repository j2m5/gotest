import { useCallback, useEffect, useRef, useState } from 'react'
import { moveToLocation, planetLocations } from '../api'
import { useGameActor } from '../../../entities/account/useGameActor'

export function useLocations(accountId: number, characterId: number, planetId: number) {
  const { store, current, report } = useGameActor(accountId, characterId)
  const [data, setData] = useState<Awaited<ReturnType<typeof planetLocations>> | null>(null)
  const [error, setError] = useState('')
  const [actionError, setActionError] = useState('')
  const [attempt, setAttempt] = useState(0)
  const [target, setTarget] = useState<number | null>(null)
  const read = useRef<AbortController | null>(null)
  const mutation = useRef<AbortController | null>(null)
  const reload = useCallback(() => {
    if (!current() || mutation.current) return
    setData(null); setError(''); setAttempt(value => value + 1)
  }, [current])
  useEffect(() => {
    const controller = new AbortController()
    read.current = controller
    const valid = () => !controller.signal.aborted && read.current === controller && current()
    planetLocations(accountId, characterId, planetId, controller.signal)
      .then(value => { if (valid()) setData(value) })
      .catch(cause => { if (valid()) report(cause, setError) })
    return () => controller.abort()
  }, [accountId, characterId, planetId, attempt, current, report])
  useEffect(() => () => { mutation.current?.abort(); mutation.current = null }, [])
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') reload() }
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => { window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh) }
  }, [reload])

  const move = async (locationId: number) => {
    if (!data || data.return_required || mutation.current || !current()) return
    read.current?.abort()
    const controller = new AbortController()
    mutation.current = controller
    setTarget(locationId); setActionError('')
    const valid = () => !controller.signal.aborted && mutation.current === controller && current()
    try {
      const result = await moveToLocation(accountId, { character_id: characterId, location_id: locationId, from_location_id: data.position.location.id }, controller.signal)
      if (!valid()) return
      if (!result.returned_home) setData(value => value && ({ ...value, position: result.position }))
      await store.refresh({ fresh: true })
      if (valid()) return result
    } catch (cause) {
      if (!valid()) return
      report(cause, setActionError)
      // A write may have reached the server even if its response was lost.
      // Read the authoritative position; never automatically repeat a move.
      setData(null); setError(''); setAttempt(value => value + 1)
    } finally {
      if (mutation.current === controller) {
        mutation.current = null
        if (current()) setTarget(null)
      }
    }
  }
  return { data, error, actionError, target, reload, move }
}
