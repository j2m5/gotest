import type { ReactNode } from 'react'
import type { Character, CharacterCatalog } from './types'
import { defaultPortrait, factionArtwork } from './appearance'
import LocalImage from '../../shared/ui/LocalImage'

interface Props {
  character: Pick<Character, 'name' | 'level' | 'avatar' | 'faction_id' | 'class_id'>
  catalog: CharacterCatalog | null
  heading?: 'h1' | 'h2'
  status?: ReactNode
}

export default function CharacterIdentity({ character, catalog, heading: Heading = 'h2', status }: Props) {
  const faction = catalog?.factions.find(item => item.id === character.faction_id)
  const characterClass = catalog?.classes.find(item => item.id === character.class_id)
  return <div className="character-identity">
    <LocalImage className="character-portrait" src={character.avatar || defaultPortrait(characterClass, faction)} alt="" width={90} height={90} />
    <div className="character-description">
      <Heading>{character.name}</Heading>
      {status && <p>{status}</p>}
      <p>Уровень {character.level}</p>
      {characterClass && <p>{characterClass.name}</p>}
      {faction && <p className="faction-name"><img className="icon" src={factionArtwork(faction).icon} alt="" width={25} height={25} />{faction.name}</p>}
    </div>
  </div>
}
