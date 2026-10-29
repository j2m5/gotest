import assert from 'node:assert/strict'
import test from 'node:test'
import axios from '../../frontend/node_modules/axios/index.js'
import { projectResource, projectCharacterResources, observeCharacter, characterResources } from '../../frontend/src/entities/character/regeneration.ts'

const snapshot = () => ({
  hp: 100, mp: 40, hp_current: 20, mp_current: 5,
  regeneration: {
    hp_baseline: 10, mp_baseline: 0,
    hp_updated_at: '2026-10-29T10:00:00.000Z',
    mp_updated_at: '2026-10-29T10:00:00.000Z',
    observed_at: '2026-10-29T10:00:01.000Z',
    hp_full_restore_ms: 10000, mp_full_restore_ms: 8000,
  },
})

test('health and energy use independent server periods and baselines', () => {
  const character = snapshot()
  assert.deepEqual(projectCharacterResources(character, 1000), { hp: 30, mp: 10 })
  assert.deepEqual(projectCharacterResources(character, 0), { hp: 20, mp: 5 })
  assert.deepEqual(projectCharacterResources(character, 9000), { hp: 100, mp: 40 })
  assert.deepEqual(projectCharacterResources(character, 1e20), { hp: 100, mp: 40 })
})

test('fractional progress is retained across repeated projections', () => {
  assert.equal(projectResource(3, 0, 1000, 10000), 0)
  assert.equal(projectResource(3, 0, 2000, 10000), 0)
  assert.equal(projectResource(3, 0, 4000, 10000), 1)
  assert.equal(projectResource(3, 0, 10000, 10000), 3)
})

test('zero maximum, future timestamps and large int32 maxima stay bounded', () => {
  assert.equal(projectResource(0, 0, 1e20, 10000), 0)
  assert.equal(projectResource(2147483647, 0, 5000, 10000), 1073741823)
  assert.equal(projectResource(2147483647, 0, 1e20, 10000), 2147483647)
  const character = snapshot()
  character.regeneration.observed_at = '2026-10-29T09:59:59.000Z'
  assert.deepEqual(projectCharacterResources(character, 0), { hp: 10, mp: 0 })
})

test('receipt uses monotonic elapsed time and ignores client wall clock skew', () => {
  const character = snapshot()
  observeCharacter(character, 500)
  const original = Date.now
  try {
    Date.now = () => 0
    assert.deepEqual(characterResources(character, 1500), { hp: 30, mp: 10 })
    Date.now = () => 1e15
    assert.deepEqual(characterResources(character, 1500), { hp: 30, mp: 10 })
  } finally { Date.now = original }
})

test('same object remount retains original receipt while distinct snapshots have separate clocks', () => {
  const first = snapshot(), second = snapshot()
  observeCharacter(first, 1000)
  observeCharacter(first, 4000)
  assert.deepEqual(characterResources(first, 4000), { hp: 50, mp: 20 })
  observeCharacter(second, 4000)
  assert.deepEqual(characterResources(second, 4000), { hp: 20, mp: 5 })
  assert.deepEqual(characterResources(first, 5000), { hp: 60, mp: 25 })
})

test('snapshots without regeneration metadata use safe static resources', () => {
  assert.deepEqual(characterResources({ hp: 10, mp: 0, hp_current: 50, mp_current: 3 }, 5000), { hp: 10, mp: 0 })
  assert.deepEqual(projectCharacterResources({ hp: 10, mp: 5, hp_current: -1, mp_current: 2 }, 5000), { hp: 0, mp: 2 })
})

test('character APIs observe every returned snapshot before callers receive it', async t => {
  let now = 1000
  t.mock.method(performance, 'now', () => now)
  const adapter = axios.defaults.adapter
  axios.defaults.adapter = async config => ({
    status: 200, statusText: 'OK', headers: {}, config,
    data: config.method === 'get'
      ? { characters: [snapshot(), snapshot()], active_character_id: null }
      : { character: config.data === '{"character_id":null}' ? null : snapshot() },
  })
  try {
    const { listCharacters, createCharacter, setActiveCharacter } = await import('../../frontend/src/features/character/api.ts')
    const roster = await listCharacters(1)
    now = 2000
    const created = await createCharacter(1, { name: 'Test', faction_id: 1, class_id: 1 })
    now = 3000
    const selected = await setActiveCharacter(1, 7)
    now = 4000
    assert.deepEqual(roster.characters.map(character => characterResources(character)), [{ hp: 50, mp: 20 }, { hp: 50, mp: 20 }])
    assert.deepEqual(characterResources(created.character), { hp: 40, mp: 15 })
    assert.deepEqual(characterResources(selected.character), { hp: 30, mp: 10 })
    assert.equal((await setActiveCharacter(1, null)).character, null)
  } finally { axios.defaults.adapter = adapter }
})
