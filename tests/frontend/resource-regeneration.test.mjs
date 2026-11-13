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

test('server carry survives resource changes and is included in both projections', () => {
  assert.equal(projectResource(7, 1, 2000, 300000, 290000), 2)
  const character = snapshot()
  character.regeneration.hp_remainder = 9000
  character.regeneration.mp_remainder = 7000
  assert.deepEqual(projectCharacterResources(character, 100), { hp: 21, mp: 6 })
  assert.equal(projectResource(100, 100, 1000, 300000, 299999), 100)
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
  axios.defaults.adapter = async config => {
    if (config.url === '/api/battles/current') {
      assert.equal(config.headers.get('X-Account-ID'), '7')
      assert.deepEqual(config.params, { character_id: 12 })
      return { status: 200, statusText: 'OK', headers: {}, config, data: { character: snapshot(), battle: null } }
    }
    if (config.url.startsWith('/api/battles/')) {
      assert.equal(config.headers.get('X-Account-ID'), '7')
      if (config.url.endsWith('/detail')) {
        assert.equal(config.method, 'get')
        assert.deepEqual(config.params, { battle_id: '9007199254740993', character_id: 12, after_sequence: 5 })
      } else {
        assert.equal(config.method, 'post')
        const bodies = {
          '/api/battles/start': { character_id: 12, location_id: 2, npc_spawn_id: 4, request_key: 'start:1' },
          '/api/battles/use': { character_id: 12, battle_id: '9007199254740993', ability_id: 3, request_key: 'cast:1' },
          '/api/battles/retreat': { battle_id: '9007199254740993', character_id: 12 },
        }
        assert.deepEqual(JSON.parse(config.data), bodies[config.url])
      }
      return { status: 200, statusText: 'OK', headers: {}, config, data: { character: snapshot(), battle: { id: '9007199254740993' } } }
    }
    if (config.url === '/api/abilities') {
      assert.equal(config.headers.get('X-Account-ID'), '7')
      assert.deepEqual(config.params, { character_id: 12 })
      return { status: 200, statusText: 'OK', headers: {}, config, data: { character: snapshot(), abilities: [] } }
    }
    return {
      status: 200, statusText: 'OK', headers: {}, config,
      data: config.method === 'get'
        ? { characters: [snapshot(), snapshot()], active_character_id: null }
        : { character: config.data === '{"character_id":null}' ? null : snapshot() },
    }
  }
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
    const { characterAbilities } = await import('../../frontend/src/features/abilities/api.ts')
    const loadout = await characterAbilities(7, 12)
    now = 6000
    assert.deepEqual(characterResources(loadout.character), { hp: 40, mp: 15 })
    const { currentBattle, startBattle, useBattleAbility, retreatBattle, battleDetail } = await import('../../frontend/src/features/combat/api.ts')
    const battleState = await currentBattle(7, 12)
    now = 8000
    assert.equal(battleState.battle, null)
    assert.deepEqual(characterResources(battleState.character), { hp: 40, mp: 15 })
    const started = await startBattle(7, { character_id: 12, location_id: 2, npc_spawn_id: 4, request_key: 'start:1' })
    const used = await useBattleAbility(7, { character_id: 12, battle_id: started.battle.id, ability_id: 3, request_key: 'cast:1' })
    const retreated = await retreatBattle(7, started.battle.id, 12)
    const detail = await battleDetail(7, started.battle.id, 12, 5)
    now = 10000
    for (const state of [started, used, retreated, detail]) {
      assert.deepEqual(characterResources(state.character), { hp: 40, mp: 15 })
    }
  } finally { axios.defaults.adapter = adapter }
})
