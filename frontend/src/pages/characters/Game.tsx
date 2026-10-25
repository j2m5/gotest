import { useState } from 'react'
import { Navigate, useNavigate } from 'react-router-dom'
import { useAccount } from '../../account/context'
import { authError } from '../../api'
import Layout from '../../components/layout/Layout'
export default function Game() {
  const { roster, pending, store } = useAccount()
  const [error, setError] = useState('')
  const navigate = useNavigate()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!character) return <Navigate to="/characters" replace />
  const clear = async () => {
    setError('')
    try { await store.select(null); navigate('/characters', { replace: true }) }
    catch (error) { setError(authError(error).message) }
  }
  return <Layout><h1>{character.name}</h1>
    {error && <p role="alert" className="flash-error">{error}</p>}
    <p>Уровень {character.level} · Опыт {character.experience}</p>
    <p>Здоровье: {character.hp_current}/{character.hp}</p><p>Мана: {character.mp_current}/{character.mp}</p><p>Персонаж готов к игре.</p>
    <button className="btn-primary" disabled={pending} onClick={() => void clear()}>{pending ? 'Подождите…' : 'К выбору персонажа'}</button>
  </Layout>
}
