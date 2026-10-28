import { Link, Navigate } from 'react-router-dom'
import { useAccount } from '../../../entities/account/context'
import CharacterIdentity from '../../../entities/character/CharacterIdentity'
import { useCharacterCatalog } from '../model/useCharacterCatalog'
import CatalogFeedback from '../ui/CatalogFeedback'
import GameLayout from '../ui/GameLayout'

export default function Profile() {
  const { roster } = useAccount()
  const catalogState = useCharacterCatalog()
  const character = roster?.characters.find(item => item.id === roster.active_character_id)
  if (!character) return <Navigate to="/characters" replace />
  return <GameLayout character={character} catalog={catalogState.catalog}>
    <CharacterIdentity character={character} catalog={catalogState.catalog} heading="h1" />
    <CatalogFeedback {...catalogState} />
    <p className="experience-row">Опыт: {character.experience}</p>
    <dl className="character-stats" aria-label="Характеристики персонажа">
      <div><dt>Здоровье</dt><dd>{character.hp_current} / {character.hp}</dd></div>
      <div><dt>Энергия</dt><dd>{character.mp_current} / {character.mp}</dd></div>
    </dl>
    <nav className="section-menu" aria-label="Разделы игры">
      <Link to="/game"><img className="icon" src="/images/icons/home.png" alt="" width={25} height={25} />На главную</Link>
    </nav>
  </GameLayout>
}
