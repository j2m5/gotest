import { useState } from 'react'
import { Link, Navigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { Alert, Button, Checkbox, Input, Select } from '../../../shared/ui/kit'
import { useCharacterCatalog } from '../../character/model/useCharacterCatalog'
import CatalogFeedback from '../../character/ui/CatalogFeedback'
import GameLayout from '../../character/ui/GameLayout'
import { filterPlanets, type CoreFilter, type ControlFilter } from '../model/catalog'
import { useWorldNavigation } from '../model/useWorldNavigation'
import '../world.css'

export default function Galaxy() {
  const { user, roster } = useAccount()
  const catalogState = useCharacterCatalog()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!user || !character) return <Navigate to="/characters" replace />
  return <GameLayout character={character} catalog={catalogState.catalog}>
    <h1>Галактическая война</h1>
    <CatalogFeedback {...catalogState} />
    <GalaxyCatalog key={`${user.id}:${character.id}`} accountId={user.id} characterId={character.id} />
  </GameLayout>
}

function GalaxyCatalog({ accountId, characterId }: { accountId: number; characterId: number }) {
  const { data, error, actionError, notice, target, reload, travel } = useWorldNavigation(accountId, characterId)
  const [query, setQuery] = useState('')
  const [core, setCore] = useState<CoreFilter>('all')
  const [control, setControl] = useState<ControlFilter>('all')
  const [availableOnly, setAvailableOnly] = useState(false)

  if (error) return <>
    <Alert variant="error" role="alert">{error}</Alert>
    <Button onClick={reload}>Повторить</Button>
  </>
  if (!data) return <>{actionError && <Alert variant="error" role="alert">{actionError}</Alert>}<p role="status">Загрузка планет…</p></>
  const { planets, navigation } = data
  const { position } = navigation
  const destinations = new Map(navigation.destinations.map(item => [item.planet_id, item]))
  const visible = filterPlanets(planets, query, core, control).filter(planet => !availableOnly || destinations.get(planet.id)?.can_travel)
  const filtered = query !== '' || core !== 'all' || control !== 'all' || availableOnly
  return <>
    <section className="world-position" data-planet-code={position.planet.code} aria-label="Текущее местоположение">
      <p className="muted">Текущее местоположение</p>
      <h2>{position.planet.name}</h2>
      <p>{position.location.name}</p>
      <p className="muted">Рекомендуемый уровень: {position.location.recommended_level}</p>
      <Link to={`/galaxy/planets/${position.planet.id}/locations`}>Локации планеты</Link>
    </section>
    {notice && <Alert role="status">{notice}</Alert>}
    {actionError && <Alert variant="error" role="alert">{actionError}</Alert>}
    {navigation.return_required && <Alert variant="warning">
      Доступ к этой планете закрыт. Следующее перемещение вернёт вас в столицу фракции.
      <Button disabled={target !== null} pending={target !== null} onClick={() => void travel(position.planet.id)}>Вернуться в столицу</Button>
    </Alert>}
    <div className="world-filters" role="search" aria-label="Поиск планет">
      <label className="world-search" htmlFor="planet-search">Название планеты
        <Input id="planet-search" type="search" placeholder="Например, Корускант" value={query} onChange={event => setQuery(event.target.value)} />
      </label>
      <label htmlFor="planet-core">Территория
        <Select id="planet-core" value={core} onChange={event => setCore(event.target.value as CoreFilter)}>
          <option value="all">Все миры</option>
          <option value="galactic_republic">Ядро Республики</option>
          <option value="sith_empire">Ядро Империи</option>
          <option value="outer">Вне ядра фракций</option>
        </Select>
      </label>
      <label htmlFor="planet-control">Контроль
        <Select id="planet-control" value={control} onChange={event => setControl(event.target.value as ControlFilter)}>
          <option value="all">Любая сторона</option>
          <option value="galactic_republic">Республика</option>
          <option value="sith_empire">Империя</option>
          <option value="neutral">Нейтральные</option>
        </Select>
      </label>
    </div>
    <label className="world-available"><Checkbox checked={availableOnly} onChange={event => setAvailableOnly(event.target.checked)} />Доступны для посещения</label>
    <div className="world-results">
      <p className="muted" role="status">Планеты: {visible.length} из {planets.length}</p>
      {filtered && <Button variant="ghost" onClick={() => { setQuery(''); setCore('all'); setControl('all'); setAvailableOnly(false) }}>Сбросить фильтры</Button>}
    </div>
    {visible.length === 0 && <p className="muted">{planets.length ? 'Планеты не найдены. Измените название или фильтры.' : 'Список планет пока пуст.'}</p>}
    <ul className="planet-list" aria-label="Планеты галактики">
      {visible.map(planet => <li key={planet.id} data-planet-code={planet.code}>
        <div className="planet-name"><strong>{planet.name}</strong><span className="muted" lang="en">{planet.name_en}</span></div>
        <div className="planet-affiliation">
          <span className={`planet-control planet-control--${planet.controlling_faction?.code ?? 'neutral'}`}>{planet.controlling_faction?.name ?? 'Нейтральная'}</span>
          {planet.core_faction && <span className="planet-core">Неотчуждаемое ядро</span>}
        </div>
        <div className="planet-travel">
          {planet.id === position.planet.id && !navigation.return_required ? <span className="muted">Вы здесь</span>
            : destinations.get(planet.id)?.can_travel ? <Button variant="secondary" disabled={target !== null} pending={target === planet.id} onClick={() => void travel(planet.id)}>{target === planet.id ? 'Перелёт…' : 'Переместиться'}</Button>
              : <span className="muted">{destinations.get(planet.id)?.reason === 'locations_unavailable' ? 'Локации недоступны' : planet.core_faction ? 'Требуется вторжение' : 'Требуется вторжение или главный приказ'}</span>}
        </div>
      </li>)}
    </ul>
  </>
}
