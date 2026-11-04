import { useCallback, useEffect, useRef, useState } from 'react'
import { apiError } from '../../../shared/api/errors'
import { planetCatalog, worldNavigation, travelToPlanet } from '../api'
import { useGameActor } from '../../../entities/account/useGameActor'

async function readWorld(accountId: number, characterId: number, signal: AbortSignal) {
  const [catalog, navigation] = await Promise.all([planetCatalog(accountId, signal), worldNavigation(accountId, characterId, signal)])
  return { planets: catalog.planets, navigation }
}

export function useWorldNavigation(accountId: number, characterId: number) {
  const { store, current, report } = useGameActor(accountId, characterId)
  const [data, setData] = useState<Awaited<ReturnType<typeof readWorld>> | null>(null)
  const [error, setError] = useState('')
  const [actionError, setActionError] = useState('')
  const [notice, setNotice] = useState('')
  const [attempt, setAttempt] = useState(0)
  const [target, setTarget] = useState<number | null>(null)
  const mutation = useRef<AbortController | null>(null)
  const reload = useCallback(() => {
    if (!current() || mutation.current) return
    setData(null); setError(''); setAttempt(value => value + 1)
  }, [current])
  useEffect(() => {
    const controller = new AbortController()
    readWorld(accountId, characterId, controller.signal)
      .then(value => { if (!controller.signal.aborted && current()) setData(value) })
      .catch(cause => { if (!controller.signal.aborted) report(cause, setError) })
    return () => controller.abort()
  }, [accountId, characterId, attempt, current, report])
  useEffect(() => () => { mutation.current?.abort(); mutation.current = null }, [])
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') reload() }
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => { window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh) }
  }, [reload])

  const travel = async (planetId: number) => {
    if (!data || mutation.current || !current()) return
    const controller = new AbortController()
    mutation.current = controller
    setTarget(planetId); setActionError(''); setNotice('')
    const valid = () => !controller.signal.aborted && mutation.current === controller && current()
    try {
      const result = await travelToPlanet(accountId, { character_id: characterId, planet_id: planetId, from_location_id: data.navigation.position.location.id }, controller.signal)
      if (!valid()) return
      setData(value => value && ({ ...value, navigation: { ...value.navigation, position: result.position, return_required: false } }))
      setNotice(result.returned_home ? 'Доступ к прежней планете закрыт. Вы вернулись в столицу своей фракции.' : `Вы прибыли: ${result.position.planet.name} — ${result.position.location.name}.`)
      await store.refresh({ fresh: true })
      if (!valid()) return
      const updated = await readWorld(accountId, characterId, controller.signal)
      if (valid()) setData(updated)
    } catch (cause) {
      if (!valid()) return
      report(cause, setActionError)
      // Reload authoritative position/permissions after a conflict. Never retry a move.
      if (['position_changed', 'world_access_denied', 'locations_unavailable'].includes(apiError(cause).code ?? '')) {
        setData(null); setError(''); setAttempt(value => value + 1)
      }
    } finally {
      if (mutation.current === controller) {
        mutation.current = null
        if (current()) setTarget(null)
      }
    }
  }
  return { data, error, actionError, notice, target, reload, travel }
}
