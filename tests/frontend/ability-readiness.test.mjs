import assert from 'node:assert/strict'
import test from 'node:test'
import { projectActions } from '../../frontend/src/features/abilities/model/readiness.ts'

const loadout = () => ({
  observed_at: '2026-11-08T10:00:00Z', global_ready_at: '2026-11-08T10:00:01.500Z', global_cooldown_ms: 1500,
  character: { hp: 100, hp_current: 100, mp: 100, mp_current: 0, regeneration: {
    hp_baseline: 100, mp_baseline: 0, hp_remainder: 0, mp_remainder: 0,
    hp_updated_at: '2026-11-08T10:00:00Z', mp_updated_at: '2026-11-08T10:00:00Z',
    observed_at: '2026-11-08T10:00:00Z', hp_full_restore_ms: 300000, mp_full_restore_ms: 60000,
  }},
  abilities: [{ id: 1, name: 'Удар', description: 'Удар по цели', icon: null, damage: 10, heal: 0, mp_cost: 2, cooldown_ms: 1000, ready_at: '2026-11-08T10:00:01Z', availability: 'cooldown' }],
})
test('ability deadlines use server time plus elapsed time and preserve priority', () => {
  const source = loadout()
  assert.equal(projectActions(source, 0)[0].reason, 'cooldown')
  assert.equal(projectActions(source, 1000)[0].reason, 'global_cooldown')
  assert.equal(projectActions(source, 1499)[0].ready, false)
  assert.equal(projectActions(source, 1500)[0].ready, true)
  assert.equal(projectActions(source, -500)[0].remainingMs, 1000)
})
test('energy regenerates into readiness while contextual blocks and defeat remain explicit', () => {
  const source = loadout()
  source.global_ready_at = null; source.abilities[0].ready_at = null
  assert.equal(projectActions(source, 0)[0].reason, 'insufficient_energy')
  assert.equal(projectActions(source, 1200)[0].ready, true)
  assert.equal(projectActions(source, 1200, 'Бой завершён')[0].ready, false)
  source.character.regeneration.hp_baseline = 0
  assert.equal(projectActions(source, 0)[0].reason, 'defeated')
  source.abilities = []
  assert.deepEqual(projectActions(source, 1000), [])
})
test('wall clock changes cannot skip a cooldown', t => {
  t.mock.method(Date, 'now', () => 1e15)
  assert.equal(projectActions(loadout(), 500)[0].remainingMs, 500)
})
