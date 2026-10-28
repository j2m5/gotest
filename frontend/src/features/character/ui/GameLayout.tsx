import { useState } from 'react'
import type { ReactNode } from 'react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import type { Character, CharacterCatalog } from '../../../entities/character/types'
import { factionArtwork } from '../../../entities/character/appearance'
import ResourceBars from '../../../entities/character/ResourceBars'
import { apiError } from '../../../shared/api/errors'
import Layout from '../../../shared/ui/Layout'

export default function GameLayout({ character, catalog, children }: { character: Character; catalog: CharacterCatalog | null; children: ReactNode }) {
  const { pending, store, user } = useAccount()
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const faction = catalog?.factions.find(item => item.id === character.faction_id)
  const clear = async () => {
    setError('')
    try { await store.select(null); navigate('/characters', { replace: true }) }
    catch (error) { setError(apiError(error).message) }
  }
  const signOut = async () => {
    setError('')
    try { await store.signOut() }
    catch (error) { setError(apiError(error).message) }
  }
  return <Layout statusPanel={<ResourceBars character={character} />} footer={<>
    <div className="character-strip">
      <Link to="/profile"><img className="icon" src={factionArtwork(faction).icon} alt="" width={25} height={25} />[{character.level}] {character.name}</Link>
    </div>
    <nav className="game-navigation" aria-label="Навигация персонажа">
      <NavLink to="/chat">Чат</NavLink>
      {user && user.role !== 'user' && <NavLink to="/moderation">Модерация</NavLink>}
      <NavLink to="/game">На главную</NavLink>
      <NavLink to="/profile">Мой персонаж</NavLink>
      <button type="button" className="link-button" disabled={pending} onClick={() => void clear()}>К выбору персонажа</button>
      <button type="button" className="link-button" disabled={pending} onClick={() => void signOut()}>{pending ? 'Подождите…' : 'Выйти'}</button>
    </nav>
  </>}>
    {error && <p role="alert" className="flash-error">{error}</p>}
    {children}
  </Layout>
}
