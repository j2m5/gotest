import { Alert, Button } from '../../../shared/ui/kit'
import { useEffect, useState } from 'react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { roleLabels } from '../../../entities/account/roles'
import CharacterIdentity from '../../../entities/character/CharacterIdentity'
import PresenceIndicator from '../../../entities/presence/PresenceIndicator'
import type { components } from '../../../shared/api/schema'
import { apiError } from '../../../shared/api/errors'
import Layout from '../../../shared/ui/Layout'
import { publicCharacterProfile } from '../api'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'

export default function PublicProfile() {
  const { characterId = '' } = useParams()
  const { user } = useAccount()
  return <ProfileContent key={`${user!.id}:${characterId}`} characterId={characterId} />
}

function ProfileContent({ characterId }: { characterId: string }) {
  const { user, store } = useAccount()
  const location = useLocation()
  const catalogState = useCharacterCatalog()
  const [character, setCharacter] = useState<components['schemas']['PublicCharacterProfile'] | null>(null)
  const [error, setError] = useState('')
  const [attempt, setAttempt] = useState(0)
  const id = Number(characterId)
  const validId = /^\d+$/.test(characterId) && Number.isInteger(id) && id > 0 && id <= 2147483647
  const accountId = user!.id
  useEffect(() => {
    if (!validId) return
    const controller = new AbortController()
    const current = () => !controller.signal.aborted && store.getSnapshot().user?.id === accountId
    void publicCharacterProfile(accountId, id, controller.signal).then(result => {
      if (current()) setCharacter(result.character)
    }, failure => {
      if (!current()) return
      const result = apiError(failure)
      if (result.code === 'account_changed') { void store.externalAuthChanged(); return }
      if (result.status === 401) { void store.refresh(); return }
      setError(result.status === 404 ? 'Персонаж не найден.' : result.message)
    })
    return () => controller.abort()
  }, [accountId, id, validId, attempt, store])

  const state: unknown = location.state
  const origin = state && typeof state === 'object' && 'profileOrigin' in state && typeof state.profileOrigin === 'string' ? state.profileOrigin : ''
  const forumOrigin = /^\/forum(?:\/(?:topics|categories)\/\d+)?$/.test(origin) ? origin : '/forum'
  const navigation = <nav className="game-navigation" aria-label="Навигация профиля">
    {origin === '/online' && <Link to="/online">Вернуться к списку онлайн</Link>}
    {origin === '/chat/moderators' && <Link to={origin}>Вернуться к списку модераторов</Link>}
    <Link to="/chat">Вернуться в чат</Link>
    <Link to={forumOrigin}>Вернуться в форум</Link>
    <Link to="/characters">К выбору персонажа</Link>
  </nav>
  return <Layout footer={navigation}>
    {character ? <>
      <CharacterIdentity character={character} catalog={catalogState.catalog} heading="h1" status={<PresenceIndicator characterId={character.id} />} />
      <p className={`account-nickname account-nickname--${character.account_role}`}>{roleLabels[character.account_role]}</p>
      <CatalogFeedback {...catalogState} />
    </> : <>
      <h1>Профиль персонажа</h1>
      {(!validId || error) ? <>
        <Alert variant="error" role="alert">{validId ? error : 'Персонаж не найден.'}</Alert>
        {validId && <Button type="button" onClick={() => { setError(''); setAttempt(value => value + 1) }}>Повторить</Button>}
      </> : <p role="status">Загрузка персонажа…</p>}
    </>}
  </Layout>
}
