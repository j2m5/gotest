import { Navigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import { factionArtwork } from '../../../entities/character/appearance'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'
import GameLayout from '../ui/GameLayout'

export default function Game() {
  const { roster } = useAccount()
  const catalogState = useCharacterCatalog()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!character) return <Navigate to="/characters" replace />
  const faction = catalogState.catalog?.factions.find(item => item.id === character.faction_id)
  const artwork = factionArtwork(faction)
  return <GameLayout character={character} catalog={catalogState.catalog}>
    <h1 className="visually-hidden">Главная</h1>
    {artwork.scene && <img className="scene-image" src={artwork.scene} alt={artwork.sceneAlt} width={550} height={120} />}
    <CatalogFeedback {...catalogState} />
    <nav className="section-menu section-menu--home" aria-labelledby="battles-heading">
      <h2 id="battles-heading">Сражения</h2>
      <span className="section-menu-inactive" role="link" aria-disabled="true">
        <img className="icon" src="/images/icons/galaxy.png" alt="" width={25} height={25} />
        Галактическая война
      </span>
    </nav>
  </GameLayout>
}
