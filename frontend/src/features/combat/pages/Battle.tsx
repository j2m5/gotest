import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useParams } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { Alert, Breadcrumbs, Button, ButtonLink, ResourceMeter } from '../../../shared/ui/kit'
import { useCharacterCatalog } from '../../character/model/useCharacterCatalog'
import GameLayout from '../../character/ui/GameLayout'
import AbilityBar from '../../abilities/ui/AbilityBar'
import { projectActions } from '../../abilities/model/readiness'
import { useWallet } from '../../wallet/model/context'
import { formatAmount } from '../../wallet/format'
import { worldNavigation } from '../../world/api'
import { useCombat } from '../model/context'
import BattleLog from '../ui/BattleLog'
import CombatFeedback from '../ui/CombatFeedback'
import '../combat.css'

function validBattleID(value: string | undefined): value is string {
  return !!value && /^[1-9]\d{0,18}$/.test(value) && BigInt(value) <= 9223372036854775807n
}
export default function Battle() {
  const { user, roster } = useAccount()
  const { catalog } = useCharacterCatalog()
  const { battleId } = useParams()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!user || !character) return <Navigate to="/characters" replace />
  return <GameLayout character={character} catalog={catalog}>
    <BattleContent accountId={user.id} characterId={character.id} battleId={validBattleID(battleId) ? battleId : null} />
  </GameLayout>
}

function BattleContent({ accountId, characterId, battleId }: { accountId: number; characterId: number; battleId: string | null }) {
  const { session, snapshot, connection } = useCombat()
  const { store } = useAccount()
  const { reload: reloadWallet } = useWallet()
  const settled = useRef(new Set<string>())
  const [now, setNow] = useState(() => performance.now())
  const [returnPath, setReturnPath] = useState<string | null>(null)
  useEffect(() => { session.watch(battleId) }, [session, battleId])
  useEffect(() => {
    const timer = window.setInterval(() => { if (document.visibilityState === 'visible') setNow(performance.now()) }, 100)
    return () => window.clearInterval(timer)
  }, [])
  useEffect(() => {
    const controller = new AbortController()
    worldNavigation(accountId, characterId, controller.signal).then(value => {
      if (!controller.signal.aborted) setReturnPath(`/galaxy/planets/${value.position.planet.id}/locations/${value.position.location.id}`)
    }).catch(() => { /* The galaxy link remains available when navigation cannot be read. */ })
    return () => controller.abort()
  }, [accountId, characterId, battleId])
  const battle = snapshot.detail?.battle.id === battleId ? snapshot.detail.battle : null
  useEffect(() => {
    if (!battle || battle.status === 'active' || settled.current.has(battle.id)) return
    settled.current.add(battle.id); reloadWallet(); void store.refresh({ fresh: true })
  }, [battle, reloadWallet, store])
  if (!battleId) return <><h1>Бой не найден</h1><Link to="/galaxy">К галактике</Link></>
  if (!battle) return <><h1>Бой</h1>{snapshot.error ? <><Alert variant="error">{snapshot.error}</Alert><Button onClick={() => void session.refresh()}>Повторить загрузку</Button></> : <p role="status">Загрузка боя…</p>}<p><Link to="/galaxy">К галактике</Link></p></>
  const active = battle.status === 'active'
  const blocked = !active ? 'Бой завершён' : snapshot.pending ? 'Подождите…' : snapshot.unconfirmed ? 'Нужно подтверждение' : snapshot.error ? 'Обновите данные' : snapshot.loading ? 'Обновляем умения…' : ''
  const actions = snapshot.loadout ? projectActions(snapshot.loadout, Math.max(0, now - snapshot.receivedAt), blocked) : []
  const reward = snapshot.events.find(event => event.kind === 'won')
  const outcomes = { active: 'Бой идёт', won: 'Победа', retreated: 'Вы отступили', defeated: 'Поражение', expired: 'Время боя истекло' }
  const ranks = { normal: 'Обычный', strong: 'Сильный', elite: 'Элитный' }
  return <div className="combat-page" data-battle-status={battle.status} data-battle-id={battle.id}>
    <h1>{active ? 'Бой' : outcomes[battle.status]}</h1>
    <Breadcrumbs items={[{ label: 'К галактике', to: '/galaxy' }, ...(returnPath ? [{ label: 'Локация', to: returnPath }] : []), { label: 'Бой' }]} />
    <section className="combat-opponent" aria-label="Противник">
      <div className="combat-opponent-heading"><h2>{battle.npc.name}</h2><span className="combat-rank">Ур. {battle.npc.level} · {ranks[battle.npc.rank]}</span></div>
      <ResourceMeter label="Здоровье противника" value={battle.npc.hp_current} max={battle.npc.hp} kind="health" />
    </section>
    {connection !== 'connected' && active && <p className="muted combat-connection" role="status">Восстанавливаем связь. Данные обновляются автоматически.</p>}
    {snapshot.error && <Alert variant="error">{snapshot.error} <Button variant="ghost" onClick={() => void session.refresh()}>Обновить</Button></Alert>}
    <CombatFeedback />
    {!active && <Alert variant={battle.status === 'won' ? 'success' : battle.status === 'defeated' ? 'error' : 'info'}>
      {battle.status === 'won' ? <span className="combat-rewards">{reward ? <>Опыт: +{formatAmount(reward.experience)} · Кредиты: +{formatAmount(reward.credits)}</> : 'Победа. Загружаем награду…'}</span>
        : battle.status === 'defeated' ? 'Вы остались в локации. Здоровье восстанавливается; опыт и кредиты сохранены.'
        : battle.status === 'retreated' ? 'Противник восстановил здоровье. Ваши ресурсы и кулдауны сохраняются.'
        : 'Встреча завершилась без награды. Противник снова доступен.'}
    </Alert>}
    <section aria-labelledby="abilities-heading"><h2 id="abilities-heading">Умения</h2>
      {snapshot.loadout ? actions.length ? <AbilityBar actions={actions} onUse={id => { void session.use(id) }} /> : <p className="muted">У персонажа пока нет изученных умений.{active && ' Можно отступить из боя.'}</p>
        : <p className="muted" role="status">{snapshot.pending ? 'Применяем действие…' : 'Обновляем умения…'}</p>}
    </section>
    <div className="combat-actions">
      {active && <Button variant="danger" data-combat-retreat disabled={snapshot.pending !== null || snapshot.unconfirmed} pending={snapshot.pending === 'retreat'} onClick={() => { void session.retreat() }}>Отступить</Button>}
      {!active && <ButtonLink data-combat-return to={returnPath ?? '/galaxy'} variant="secondary">{returnPath ? 'К локациям' : 'К галактике'}</ButtonLink>}
    </div>
    <BattleLog events={snapshot.events} npcName={battle.npc.name} clearedAt={snapshot.detail?.log_cleared_at ?? null} />
  </div>
}
