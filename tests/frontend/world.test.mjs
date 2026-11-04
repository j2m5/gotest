import assert from 'node:assert/strict'
import test from 'node:test'
import { filterPlanets } from '../../frontend/src/features/world/model/catalog.ts'

const republic = { id: 101, code: 'galactic_republic', name: 'Республика' }
const empire = { id: 202, code: 'sith_empire', name: 'Империя' }
const planets = [
  { id: 1, code: 'coruscant', name: 'Корускант', name_en: 'Coruscant', description: null, core_faction: republic, controlling_faction: republic },
  { id: 2, code: 'dromund_kaas', name: 'Дромунд-Каас', name_en: 'Dromund Kaas', description: null, core_faction: empire, controlling_faction: empire },
  { id: 3, code: 'tatooine', name: 'Татуин', name_en: 'Tatooine', description: null, core_faction: null, controlling_faction: empire },
  { id: 4, code: 'hoth', name: 'Хот', name_en: 'Hoth', description: null, core_faction: null, controlling_faction: null },
]

test('planet search accepts Russian and English names regardless of case and surrounding spaces', () => {
  assert.deepEqual(filterPlanets(planets, '  КОРУСКАНТ ', 'all', 'all').map(p => p.id), [1])
  assert.deepEqual(filterPlanets(planets, 'dromund kaas', 'all', 'all').map(p => p.id), [2])
  assert.deepEqual(filterPlanets(planets, 'несуществующая планета', 'all', 'all'), [])
})

test('captured planets remain outside the protected core and combined filters intersect', () => {
  const before = structuredClone(planets)
  assert.deepEqual(filterPlanets(planets, '', 'sith_empire', 'all').map(p => p.id), [2])
  assert.deepEqual(filterPlanets(planets, '', 'outer', 'sith_empire').map(p => p.id), [3])
  assert.deepEqual(filterPlanets(planets, 'TATOOINE', 'outer', 'sith_empire').map(p => p.id), [3])
  assert.deepEqual(filterPlanets(planets, 'TATOOINE', 'all', 'galactic_republic'), [])
  assert.deepEqual(filterPlanets(planets, '', 'all', 'neutral').map(p => p.id), [4])
  assert.deepEqual(filterPlanets(planets, '', 'galactic_republic', 'all').map(p => p.id), [1])
  assert.deepEqual(filterPlanets(planets, '', 'all', 'all'), before)
  assert.deepEqual(planets, before)
})
