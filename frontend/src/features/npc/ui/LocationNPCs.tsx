import { Alert, Button } from '../../../shared/ui/kit'
import { useLocationNPCs } from '../model/useLocationNPCs'
import NPCKindIcon from './NPCKindIcon'
import '../npc.css'
import { useNavigate } from 'react-router-dom'
import { useCombat } from '../../combat/model/context'
import CombatFeedback from '../../combat/ui/CombatFeedback'
import { useCharacterResources } from '../../../entities/character/useCharacterResources'

const ranks = { normal: 'Обычный', strong: 'Сильный', elite: 'Элитный' }
const kinds = { humanoid: 'Гуманоид', droid: 'Дроид', beast: 'Зверь' }

export default function LocationNPCs({ accountId, characterId, locationId, onContextChanged }: { accountId: number; characterId: number; locationId: number; onContextChanged: () => void }) {
  const { data, error, reload } = useLocationNPCs(accountId, characterId, locationId, onContextChanged)
  const { session, snapshot } = useCombat()
  const resources = useCharacterResources(snapshot.character ?? undefined)
  const navigate = useNavigate()
  const attack = async (npcId: number) => {
    const result = await session.begin(locationId, npcId)
    if (result) navigate(`/battle/${result.battle.id}`)
    else if (!session.getSnapshot().unconfirmed) reload()
  }
  const unavailable = !snapshot.current || !!snapshot.error || snapshot.pending !== null || snapshot.unconfirmed || !!snapshot.current.battle || resources.hp <= 0
  const loading = !data && !error
  return <section className="npc-section" aria-labelledby="npc-heading" aria-busy={loading}>
    <div className="npc-toolbar"><h2 id="npc-heading">Противники</h2>{!error && <Button variant="secondary" disabled={loading} onClick={reload}>Обновить список</Button>}</div>
    <CombatFeedback />
    {snapshot.error && <Alert variant="error">{snapshot.error} <Button variant="ghost" onClick={() => void session.refresh()}>Обновить данные боя</Button></Alert>}
    {error ? <><Alert variant="error">{error}</Alert><Button onClick={reload}>Повторить</Button></>
      : !data ? <p role="status">Загрузка противников…</p>
        : data.npcs.length === 0 ? <p className="muted">Пока нет противников.</p>
          : <ul className="npc-list" aria-label="Противники в локации">{data.npcs.map(npc => <li key={npc.id} data-npc-id={npc.id}>
            <div className={`npc-portrait npc-rank--${npc.rank}`}><NPCKindIcon kind={npc.kind} /></div>
            <div className="npc-content">
              <h3>{npc.name}</h3>
              <p className="npc-meta"><span>Уровень {npc.level}</span><span className={`npc-rank--${npc.rank}`}>{ranks[npc.rank]}</span><span>{kinds[npc.kind]}</span></p>
              <p className="npc-description">{npc.description}</p>
              <Button className="npc-attack" data-attack-npc={npc.id} disabled={unavailable} onClick={() => void attack(npc.id)}>Атаковать</Button>
            </div>
          </li>)}</ul>}
  </section>
}
