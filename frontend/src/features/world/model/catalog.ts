import type { components } from '../../../shared/api/schema'

export type Planet = components['schemas']['Planet']
export type CoreFilter = 'all' | 'outer' | 'galactic_republic' | 'sith_empire'
export type ControlFilter = 'all' | 'neutral' | 'galactic_republic' | 'sith_empire'

export function filterPlanets(planets: Planet[], query: string, core: CoreFilter, control: ControlFilter): Planet[] {
  const search = query.trim().toLocaleLowerCase('ru')
  return planets.filter(planet => {
    if (core === 'outer' && planet.core_faction !== null) return false
    if (core !== 'all' && core !== 'outer' && planet.core_faction?.code !== core) return false
    if (control === 'neutral' && planet.controlling_faction !== null) return false
    if (control !== 'all' && control !== 'neutral' && planet.controlling_faction?.code !== control) return false
    return `${planet.name}\n${planet.name_en}`.toLocaleLowerCase('ru').includes(search)
  })
}
