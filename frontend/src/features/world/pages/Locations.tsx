import { Link, Navigate, useLocation, useNavigate, useParams } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { Alert, Button, ButtonLink } from '../../../shared/ui/kit'
import { useCharacterCatalog } from '../../character/model/useCharacterCatalog'
import CatalogFeedback from '../../character/ui/CatalogFeedback'
import GameLayout from '../../character/ui/GameLayout'
import LocationNPCs from '../../npc/ui/LocationNPCs'
import { useLocations } from '../model/useLocations'
import '../world.css'

function routeID(value: string | undefined) {
  if (!value || !/^[1-9]\d*$/.test(value)) return null
  const id = Number(value)
  return Number.isSafeInteger(id) && id <= 2147483647 ? id : null
}

export default function Locations() {
  const { user, roster } = useAccount()
  const catalogState = useCharacterCatalog()
  const params = useParams()
  const planetId = routeID(params.planetId)
  const locationId = routeID(params.locationId)
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!user || !character) return <Navigate to="/characters" replace />
  return <GameLayout character={character} catalog={catalogState.catalog}>
    <CatalogFeedback {...catalogState} />
    {!planetId || (params.locationId !== undefined && !locationId)
      ? <><h1>Локация не найдена</h1><Link to="/galaxy">К галактике</Link></>
      : <LocationContent key={`${user.id}:${character.id}:${planetId}:${locationId}`} accountId={user.id} characterId={character.id} planetId={planetId} locationId={locationId} />}
  </GameLayout>
}

function LocationContent({ accountId, characterId, planetId, locationId }: { accountId: number; characterId: number; planetId: number; locationId: number | null }) {
  const { data, error, actionError, target, reload, move } = useLocations(accountId, characterId, planetId)
  const navigate = useNavigate()
  const route = useLocation()
  const listPath = `/galaxy/planets/${planetId}/locations`
  const enter = async (id: number) => {
    const result = await move(id)
    if (result) navigate(`/galaxy/planets/${result.position.planet.id}/locations/${result.position.location.id}`, { replace: locationId === result.position.location.id, state: { returnedHome: result.returned_home } })
  }
  if (error) return <><h1>Локации планеты</h1><Alert variant="error">{error}</Alert><div className="location-actions"><Button onClick={reload}>Повторить</Button><Link to="/galaxy">К галактике</Link></div></>
  if (!data) return <><h1>Локации планеты</h1>{actionError && <Alert variant="error">{actionError}</Alert>}<p role="status">Загрузка локаций…</p></>
  const selected = locationId === null ? null : data.locations.find(item => item.id === locationId)
  if (locationId !== null && !selected) return <><h1>Локация не найдена</h1><Alert variant="error">На этой планете такой локации нет.</Alert><Link to={listPath}>Локации планеты</Link></>
  const here = selected?.id === data.position.location.id
  return <div className="world-locations" aria-busy={target !== null}>
    <h1>{selected ? selected.name : data.position.planet.name}</h1>
    <nav className="location-navigation" aria-label="Навигация по миру"><Link to="/galaxy">К галактике</Link>{selected && <Link to={listPath}>Локации планеты</Link>}</nav>
    {route.state?.returnedHome === true && <Alert>Доступ к прежней планете закрыт. Вы вернулись в столицу своей фракции.</Alert>}
    <p className="location-current" data-current-location-id={data.position.location.id}>Вы находитесь: <strong>{data.position.location.name}</strong></p>
    {actionError && <Alert variant="error">{actionError}</Alert>}
    {data.return_required && <Alert variant="warning">Доступ к планете закрыт. Вернитесь в столицу через экран галактики.</Alert>}
    {selected ? <section className="location-detail">
      <p className="muted">{data.position.planet.name} · Рекомендуемый уровень: {selected.recommended_level}</p>
      <p>{selected.description}</p>
      {!here && !data.return_required && <><p className="muted">Перейдите в локацию, чтобы увидеть противников.</p><Button className="location-enter" disabled={target !== null} pending={target === selected.id} onClick={() => void enter(selected.id)}>Перейти в локацию</Button></>}
      {here && !data.return_required && target === null && <LocationNPCs key={`${accountId}:${characterId}:${selected.id}`} accountId={accountId} characterId={characterId} locationId={selected.id} onContextChanged={reload} />}
    </section> : <>
      <p className="muted">Рекомендуемый уровень показывает сложность. Посещать локации можно на любом уровне.</p>
      <ul className="location-list" aria-label="Локации планеты">{data.locations.map(item => <li key={item.id} data-location-id={item.id}>
        <h2><Link className="location-title" to={`${listPath}/${item.id}`}>{item.name}</Link></h2>
        <p className="location-level muted">Рекомендуемый уровень: {item.recommended_level}</p>
        <p>{item.description}</p>
        <div className="location-actions">{item.id === data.position.location.id
          ? <><span className="muted">Вы здесь</span><ButtonLink variant="secondary" to={`${listPath}/${item.id}`}>Открыть</ButtonLink></>
          : !data.return_required && <Button disabled={target !== null} pending={target === item.id} onClick={() => void enter(item.id)}>Перейти</Button>}</div>
      </li>)}</ul>
    </>}
  </div>
}
