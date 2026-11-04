import { useCallback, useEffect, useRef, useState } from 'react'
import { useGameActor } from '../../../entities/account/useGameActor'
import { apiError } from '../../../shared/api/errors'
import { walletHistory } from '../api'

export function useWalletHistory(accountId: number, characterId: number) {
  const { current, report } = useGameActor(accountId, characterId)
  const [data, setData] = useState<Awaited<ReturnType<typeof walletHistory>> | null>(null)
  const [error, setError] = useState('')
  const [moreError, setMoreError] = useState('')
  const [morePending, setMorePending] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const read = useRef<AbortController | null>(null)
  const more = useRef<AbortController | null>(null)
  const reload = useCallback(() => {
    if (!current()) return
    read.current?.abort(); more.current?.abort(); more.current = null
    setData(null); setError(''); setMoreError(''); setMorePending(false); setAttempt(value => value + 1)
  }, [current])
  useEffect(() => {
    const controller = new AbortController()
    read.current = controller
    const valid = () => !controller.signal.aborted && read.current === controller && current()
    walletHistory(accountId, characterId, undefined, controller.signal)
      .then(value => { if (valid()) setData(value) })
      .catch(cause => { if (valid()) report(cause, setError) })
    return () => controller.abort()
  }, [accountId, characterId, attempt, current, report])
  useEffect(() => () => { more.current?.abort() }, [])
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') reload() }
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => { window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh) }
  }, [reload])

  const loadMore = async () => {
    const cursor = data?.next_cursor
    if (!cursor || more.current || !current()) return
    const controller = new AbortController()
    more.current = controller
    setMorePending(true); setMoreError('')
    const valid = () => !controller.signal.aborted && more.current === controller && current()
    try {
      const page = await walletHistory(accountId, characterId, cursor, controller.signal)
      if (valid()) setData(previous => previous?.next_cursor === cursor ? { entries: [...previous.entries, ...page.entries], next_cursor: page.next_cursor } : previous)
    } catch (cause) {
      if (!valid()) return
      const failure = apiError(cause)
      if (failure.status === 401 || ['account_changed', 'character_changed', 'character_required'].includes(failure.code ?? '')) {
        setData(null); report(cause, setError)
      } else report(cause, setMoreError)
    } finally {
      if (valid()) { more.current = null; setMorePending(false) }
    }
  }
  return { data, error, moreError, morePending, reload, loadMore }
}
