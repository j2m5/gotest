import { Alert, Button } from '../../../shared/ui/kit'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { apiError } from '../../../shared/api/errors'
import Layout from '../../../shared/ui/Layout'
import CharacterIdentity from '../../../entities/character/CharacterIdentity'
import type { Character } from '../../../entities/character/types'
import { useCharacterResources } from '../../../entities/character/useCharacterResources'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'

function ResourceLabel({ character }: { character: Character }) {
  const resources = useCharacterResources(character)
  return <p className="sm-text muted">Здоровье: {resources.hp}/{character.hp} · Энергия: {resources.mp}/{character.mp}</p>
}

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
    {error && <Alert variant="error" role="alert">{error}</Alert>}
    <CatalogFeedback {...catalogState} />
    {roster?.characters.length === 0 && <p>У вас пока нет персонажей. Создайте первого.</p>}
    <ul className="character-list">{roster?.characters.map(character => <li className="character-card" key={character.id}>
      <CharacterIdentity character={character} catalog={catalogState.catalog} />
      <ResourceLabel character={character} />
      {roster.active_character_id === character.id && <p className="sm-text">Текущий персонаж</p>}
      <Button pending={pending} disabled={pending} onClick={() => void play(character.id)}>Играть</Button>
    </li>)}</ul>
    {!pending && <p><Link to="/characters/new">Создать персонажа</Link></p>}
    <Button variant="secondary" pending={pending} type="button" disabled={pending} onClick={() => void signOut()}>{pending ? 'Подождите…' : 'Выйти'}</Button>
  </Layout>
}
