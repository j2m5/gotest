import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { apiError } from '../../../shared/api/errors'
import Layout from '../../../shared/ui/Layout'
import CharacterIdentity from '../../../entities/character/CharacterIdentity'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'
export default function Characters() {
  const { user, roster, pending, store } = useAccount()
  const catalogState = useCharacterCatalog()
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const play = async (id: number) => {
    setError('')
    try { await store.select(id); navigate('/game') }
    catch (error) { setError(apiError(error).message) }
  }
  const signOut = async () => {
    setError('')
    try { await store.signOut() }
    catch (error) { setError(apiError(error).message) }
  }
  return <Layout><h1>Персонажи</h1><p className="account-name muted">Аккаунт: {user?.login}</p>
    {error && <p role="alert" className="flash-error">{error}</p>}
    <CatalogFeedback {...catalogState} />
    {roster?.characters.length === 0 && <p>У вас пока нет персонажей. Создайте первого.</p>}
    <ul className="character-list">{roster?.characters.map(character => <li className="character-card" key={character.id}>
      <CharacterIdentity character={character} catalog={catalogState.catalog} />
      <p className="sm-text muted">Здоровье: {character.hp_current}/{character.hp} · Энергия: {character.mp_current}/{character.mp}</p>
      {roster.active_character_id === character.id && <p className="sm-text">Текущий персонаж</p>}
      <button className="btn-primary" disabled={pending} onClick={() => void play(character.id)}>Играть</button>
    </li>)}</ul>
    {!pending && <p><Link to="/characters/new">Создать персонажа</Link></p>}
    <button type="button" className="btn-primary" disabled={pending} onClick={() => void signOut()}>{pending ? 'Подождите…' : 'Выйти'}</button>
  </Layout>
}
