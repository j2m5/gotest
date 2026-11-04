import { Alert, Button } from '../../../shared/ui/kit'
import { useLocationNPCs } from '../model/useLocationNPCs'
import NPCKindIcon from './NPCKindIcon'
import '../npc.css'

const ranks = { normal: 'Обычный', strong: 'Сильный', elite: 'Элитный' }
const kinds = { humanoid: 'Гуманоид', droid: 'Дроид', beast: 'Зверь' }

export default function LocationNPCs({ accountId, characterId, locationId, onContextChanged }: { accountId: number; characterId: number; locationId: number; onContextChanged: () => void }) {
  const { data, error, reload } = useLocationNPCs(accountId, characterId, locationId, onContextChanged)
  const loading = !data && !error
  return <section className="npc-section" aria-labelledby="npc-heading" aria-busy={loading}>
    <div className="npc-toolbar"><h2 id="npc-heading">Противники</h2>{!error && <Button variant="secondary" disabled={loading} onClick={reload}>Обновить список</Button>}</div>
    {error ? <><Alert variant="error">{error}</Alert><Button onClick={reload}>Повторить</Button></>
      : !data ? <p role="status">Загрузка противников…</p>
        : data.npcs.length === 0 ? <p className="muted">Пока нет противников.</p>
          : <ul className="npc-list" aria-label="Противники в локации">{data.npcs.map(npc => <li key={npc.id} data-npc-id={npc.id}>
            <div className={`npc-portrait npc-rank--${npc.rank}`}><NPCKindIcon kind={npc.kind} /></div>
            <div className="npc-content">
              <h3>{npc.name}</h3>
              <p className="npc-meta"><span>Уровень {npc.level}</span><span className={`npc-rank--${npc.rank}`}>{ranks[npc.rank]}</span><span>{kinds[npc.kind]}</span></p>
              <p className="npc-description">{npc.description}</p>
            </div>
          </li>)}</ul>}
  </section>
}
