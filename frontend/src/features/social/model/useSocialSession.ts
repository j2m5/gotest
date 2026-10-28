import { useEffect, useMemo } from 'react'
import { useAccount } from '../../../entities/account/context'
import { SocialSession } from './session'

// Every operation belongs to the account AND character that displayed its form.
export function useSocialSession() {
  const { store, user, roster } = useAccount()
  const accountId = user!.id
  const characterId = roster?.active_character_id ?? null
  const session = useMemo(() => new SocialSession(store, accountId, characterId), [store, accountId, characterId])
  useEffect(() => {
    session.start()
    return () => session.stop()
  }, [session])
  return { accountId, characterId, run: session.run, report: session.report, current: session.current }
}
