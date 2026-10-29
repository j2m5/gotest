import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import Layout from '../../../shared/ui/Layout'
import { useCharacterCatalog } from '../../character/model/useCharacterCatalog'
import GameLayout from '../../character/ui/GameLayout'
import { canModerate } from '../model/policy'

export default function SocialLayout({ children }: { children: ReactNode }) {
  const { user, roster } = useAccount()
  const { catalog } = useCharacterCatalog()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  const body = <div className="social-page">
    <nav className="social-tabs" aria-label="Общение"><Link to="/chat">Общий чат</Link><Link to="/forum">Форум</Link>{user && canModerate(user.role) && <Link to="/moderation">Модерация</Link>}</nav>
    {children}
  </div>
  return character ? <GameLayout character={character} catalog={catalog}>{body}</GameLayout> : <Layout footer={<Link to="/characters">К выбору персонажа</Link>}>{body}</Layout>
}
