import { useCallback, useEffect, useRef, useState } from 'react'
import { useGameActor } from '../../../entities/account/useGameActor'
import { walletBalances } from '../api'

export function useWalletBalance(accountId: number, characterId: number) {
  const { current, report } = useGameActor(accountId, characterId)
  const [data, setData] = useState<Awaited<ReturnType<typeof walletBalances>> | null>(null)
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
    walletBalances(accountId, characterId, controller.signal)
      .then(value => { if (valid()) setData(value) })
      .catch(cause => { if (valid()) report(cause, setError) })
    return () => controller.abort()
  }, [accountId, characterId, attempt, current, report])
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') reload() }
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => { window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh) }
  }, [reload])
  return { data, error, reload }
}
