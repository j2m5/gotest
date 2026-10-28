import { Link, Navigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { factionArtwork } from '../../../entities/character/appearance'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'
import GameLayout from '../ui/GameLayout'

export default function Game() {
  const { roster, user } = useAccount()
  const catalogState = useCharacterCatalog()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!character) return <Navigate to="/characters" replace />
  const faction = catalogState.catalog?.factions.find(item => item.id === character.faction_id)
  const artwork = factionArtwork(faction)
  return <GameLayout character={character} catalog={catalogState.catalog}>
    {artwork.scene && <img className="scene-image" src={artwork.scene} alt={artwork.sceneAlt} width={550} height={120} />}
    <h1>{character.name}</h1>
    <p className="text-center">Уровень {character.level} · Опыт {character.experience}</p>
    <CatalogFeedback {...catalogState} />
    {faction && <p className="text-center muted">{faction.name}</p>}
    <nav className="section-menu" aria-label="Разделы игры">
      <Link to="/chat">Общий чат</Link>
      {user && user.role !== 'user' && <Link to="/moderation">Модерация</Link>}
      <Link to="/profile"><img className="icon" src={artwork.icon} alt="" width={25} height={25} />Мой персонаж</Link>
    </nav>
  </GameLayout>
}
