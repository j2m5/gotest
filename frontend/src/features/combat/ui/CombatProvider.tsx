import { useEffect, useMemo, useState, useSyncExternalStore, type ReactNode } from 'react'
import { useAccount } from '../../../entities/account/context'
import { connectNotifications, type ConnectionStatus } from '../../../shared/api/notifications'
import { characterAbilities } from '../../abilities/api'
import { battleDetail, currentBattle, retreatBattle, startBattle, useBattleAbility } from '../api'
import { CombatContext } from '../model/context'
import { CombatSession } from '../model/session'

export default function CombatProvider({ accountId, characterId, children }: { accountId: number; characterId: number; children: ReactNode }) {
  const { store } = useAccount()
  const [connection, setConnection] = useState<ConnectionStatus>('connecting')
  const session = useMemo(() => new CombatSession({
    current: signal => currentBattle(accountId, characterId, signal),
    detail: (id, after, signal) => battleDetail(accountId, id, characterId, after, signal),
    abilities: signal => characterAbilities(accountId, characterId, signal),
    begin: (locationId, npcId, key, signal) => startBattle(accountId, { character_id: characterId, location_id: locationId, npc_spawn_id: npcId, request_key: key }, signal),
    use: (id, abilityId, key, signal) => useBattleAbility(accountId, { character_id: characterId, battle_id: id, ability_id: abilityId, request_key: key }, signal),
    retreat: (id, signal) => retreatBattle(accountId, id, characterId, signal),
  }, {
    current: () => {
      const state = store.getSnapshot()
      return state.user?.id === accountId && state.roster?.active_character_id === characterId
    },
    report: failure => {
      if (failure.code === 'account_changed') void store.externalAuthChanged()
      else if (failure.status === 401 || failure.code === 'character_changed' || failure.code === 'character_required') void store.refresh()
    },
  }), [store, accountId, characterId])
  const snapshot = useSyncExternalStore(session.subscribe, session.getSnapshot)
  useEffect(() => {
    const refresh = () => { if (document.visibilityState === 'visible') void session.refresh() }
    session.start()
    const socket = connectNotifications({
      path: '/api/battles/events', query: { account_id: accountId, character_id: characterId },
      origin: window.location.origin, refresh, status: setConnection,
    })
    const timer = window.setInterval(refresh, 5000)
    window.addEventListener('focus', refresh)
    document.addEventListener('visibilitychange', refresh)
    return () => {
      socket.stop(); session.stop(); window.clearInterval(timer)
      window.removeEventListener('focus', refresh); document.removeEventListener('visibilitychange', refresh)
    }
  }, [session, accountId, characterId])
  return <CombatContext value={{ session, snapshot, connection }}>{children}</CombatContext>
}
