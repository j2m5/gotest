import type { CatalogEntry, CharacterClass } from './types'

export function defaultPortrait(characterClass?: CharacterClass, faction?: CatalogEntry) {
  switch (characterClass?.code) {
    case 'sith_warrior': return '/images/avatars/sith_warrior.gif'
    case 'sith_inquisitor': return '/images/avatars/sith_inquisitor.gif'
    case 'jedi_knight': return '/images/avatars/jedi_knight.jpg'
    case 'jedi_consular': return '/images/avatars/jedi_consular.gif'
    default: return faction?.icon ?? '/images/icons/neutral.png'
  }
}

// Codes belong to the catalog; database IDs may differ between environments.
export function factionArtwork(faction?: CatalogEntry) {
  switch (faction?.code) {
    case 'sith_empire':
      return { icon: '/images/icons/empire.png', scene: '/images/scenes/dromund_kaas.jpg', sceneAlt: 'Панорама Дромунд-Кааса' }
    case 'galactic_republic':
      return { icon: '/images/icons/republic.png', scene: '/images/scenes/coruscant.jpg', sceneAlt: 'Панорама Корусканта' }
    default:
      return { icon: '/images/icons/neutral.png', scene: undefined, sceneAlt: '' }
  }
}
