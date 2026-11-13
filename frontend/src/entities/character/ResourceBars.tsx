import type { Character } from './types'
import { useCharacterResources } from './useCharacterResources'

import { ResourceMeter } from '../../shared/ui/kit'

export default function ResourceBars({ character }: { character: Character }) {
  const resources = useCharacterResources(character)
  return <div className="resource-panel" aria-label="Ресурсы персонажа">
    <ResourceMeter label="Здоровье" value={resources.hp} max={character.hp} kind="health" />
    <ResourceMeter label="Энергия" value={resources.mp} max={character.mp} kind="energy" />
  </div>
}
