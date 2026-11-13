import assert from 'node:assert/strict'
import test from 'node:test'
import { CombatSession } from '../../frontend/src/features/combat/model/session.ts'
import { connectNotifications } from '../../frontend/src/shared/api/notifications.ts'

const deferred = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
const tick = async () => { for (let i = 0; i < 40; i++) await Promise.resolve() }
const character = () => ({ id: 3, hp: 100, hp_current: 100, mp: 100, mp_current: 100, regeneration: { observed_at: '2026-11-08T10:00:00Z' } })
const battle = () => ({ id: '9007199254740993', status: 'active', character_id: 3, location_id: 1, npc: { hp: 100, hp_current: 100 } })
const detail = (events = [{ sequence: 1, kind: 'started' }], next_after = null) => ({ character: character(), battle: battle(), events, next_after, log_cleared_at: null })
function setup(overrides = {}) {
  let valid = true, counter = 0
  const failures = [], writes = [], cursors = []
  const api = {
    current: async () => ({ character: character(), battle: battle() }),
    abilities: async () => ({ character: character(), abilities: [], observed_at: '2026-11-08T10:00:00Z', global_ready_at: null, global_cooldown_ms: 1500 }),
    detail: async (id, cursor) => { cursors.push(cursor); return detail(cursor ? [] : undefined) },
    begin: async (...args) => { writes.push(args); return { character: character(), battle: battle(), event_sequence: 900, replayed: false } },
    use: async (...args) => { writes.push(args); return { character: character(), battle: battle(), event_sequence: 900, replayed: false } },
    retreat: async (...args) => { writes.push(args); return { character: character(), battle: { ...battle(), status: 'retreated' }, event_sequence: 2, replayed: false } },
    ...overrides,
  }
  const session = new CombatSession(api, { current: () => valid, report: error => failures.push(error) }, { now: () => 100, newKey: () => 'key:' + ++counter })
  session.watch(battle().id); session.start()
  return { session, writes, cursors, failures, replace: () => { valid = false } }
}
test('combat pages merge by sequence and receipts never advance the journal cursor', async () => {
  let calls = 0
  const cursors = []
  const f = setup({ detail: async (id, cursor) => {
    cursors.push(cursor); calls++
    if (calls === 1) return detail([{ sequence: 1, kind: 'started' }], 1)
    if (calls === 2) return detail([{ sequence: 1, kind: 'started' }, { sequence: 2, kind: 'npc_ability' }])
    return detail([])
  } })
  await tick()
  assert.deepEqual(f.session.getSnapshot().events.map(e => e.sequence), [1, 2])
  await f.session.use(1); await tick()
  assert.equal(cursors.at(-1), 2)
  assert.equal(f.writes.length, 1)
  f.session.stop()
})
test('cleared combat history replaces cached detail and preserves the final reward', async () => {
  let archived = false
  const reward = { sequence: 3, kind: 'won', experience: '10', credits: '3' }
  const f = setup({ detail: async () => archived
    ? { ...detail([reward]), battle: { ...battle(), status: 'won' }, log_cleared_at: '2026-12-13T10:00:00Z' }
    : detail([{ sequence: 1, kind: 'started' }, { sequence: 2, kind: 'player_ability' }]) })
  await tick()
  archived = true
  await f.session.refresh()
  assert.deepEqual(f.session.getSnapshot().events, [reward])
  assert.equal(f.session.getSnapshot().detail.log_cleared_at, '2026-12-13T10:00:00Z')
  await f.session.refresh()
  assert.deepEqual(f.session.getSnapshot().events, [reward])
  f.session.stop()
})
test('cleanup between journal pages discards the previous page', async () => {
  let calls = 0
  const cursors = [], reward = { sequence: 53, kind: 'won', experience: '12', credits: '8' }
  const f = setup({ detail: async (_id, cursor) => {
    cursors.push(cursor)
    if (++calls === 1) return detail([{ sequence: 1, kind: 'started' }], 1)
    return { ...detail([reward]), log_cleared_at: '2026-12-13T10:00:00Z', battle: { ...battle(), status: 'won' } }
  } })
  await tick()
  assert.deepEqual(cursors, [0, 1])
  assert.deepEqual(f.session.getSnapshot().events, [reward])
  f.session.stop()
})
test('write cancels older reads; late snapshots never undo the accepted effect', async () => {
  const old = deferred(); let reads = 0, signal
  const f = setup({ current: async contextSignal => { reads++; if (reads === 2) { signal = contextSignal; return old.promise }; return { character: character(), battle: battle() } } })
  await tick()
  f.session.refresh(); await tick()
  await f.session.use(1)
  assert.equal(signal.aborted, true)
  old.resolve({ character: { ...character(), hp_current: 999 }, battle: battle() }); await tick()
  assert.notEqual(f.session.getSnapshot().character.hp_current, 999)
  f.session.stop()
})
test('uncertain write exposes explicit same-key retry and blocks double clicks', async () => {
  const pending = deferred(), calls = []
  const f = setup({ use: (...args) => { calls.push(args); return calls.length === 1 ? pending.promise : Promise.resolve({ character: character(), battle: battle(), event_sequence: 2, replayed: true }) } })
  await tick()
  const first = f.session.use(1)
  await f.session.use(2)
  assert.equal(calls.length, 1)
  pending.reject(new Error('network')); await first; await tick()
  assert.equal(f.session.getSnapshot().unconfirmed, true)
  await f.session.use(2); assert.equal(calls.length, 1)
  await f.session.retry(); await tick()
  assert.deepEqual(calls.map(c => c.slice(0, 3)), [[battle().id, 1, 'key:1'], [battle().id, 1, 'key:1']])
  assert.equal(f.session.getSnapshot().unconfirmed, false)
  f.session.stop()
})
test('ability controls keep their snapshot while a command reconciles, but cannot spend again', async () => {
  const pending = deferred(), reconciled = deferred(); let reads = 0, writes = 0
  const f = setup({ abilities: async () => ++reads === 1 ? { character: character(), abilities: [] } : reconciled.promise,
    use: () => { writes++; return pending.promise } })
  await tick()
  const loadout = f.session.getSnapshot().loadout
  const first = f.session.use(1)
  assert.equal(f.session.getSnapshot().loadout, loadout, 'keep mounted controls during the command')
  assert.equal(f.session.getSnapshot().loading, true)
  pending.resolve({ character: character(), battle: battle(), event_sequence: 2, replayed: false }); await first
  await f.session.use(1)
  assert.equal(writes, 1, 'old loadout remains blocked until the follow-up read')
  reconciled.resolve({ character: character(), abilities: [] }); await tick()
  assert.equal(f.session.getSnapshot().loading, false)
  f.session.stop()
})
test('replacement account and unmount abort requests and ignore late errors', async () => {
  const pending = deferred(); let signal
  const f = setup({ use: (_id, _ability, _key, contextSignal) => { signal = contextSignal; return pending.promise } })
  await tick()
  const first = f.session.use(1)
  f.replace(); f.session.stop()
  assert.equal(signal.aborted, true)
  pending.reject({ response: { status: 401 } }); await first
  assert.deepEqual(f.failures, [])
  await f.session.retry(); assert.equal(f.session.getSnapshot().unconfirmed, false)
})
test('notifications coalesce and a rejected cast reconciles due NPC damage without retrying', async () => {
  const first = deferred(); let reads = 0, writes = 0
  const f = setup({ current: async () => ++reads === 1 ? first.promise : { character: character(), battle: battle() }, use: async () => { writes++; throw { response: { status: 409, data: { code: 'ability_cooldown', error: 'Ещё восстанавливается' } } } } })
  f.session.refresh(); f.session.refresh(); f.session.refresh()
  assert.equal(reads, 1)
  first.resolve({ character: character(), battle: battle() }); await tick()
  assert.equal(reads, 2)
  await f.session.use(1); await tick()
  assert.equal(writes, 1); assert.equal(f.session.getSnapshot().unconfirmed, false)
  assert.ok(reads >= 3); assert.match(f.session.getSnapshot().actionError, /восстанавливается/)
  f.session.stop()
})
test('combat notification URL retains account and character while using the shared connection', () => {
  let url
  const socket = { close() {} }
  const connection = connectNotifications({ path: '/api/battles/events', query: { account_id: 7, character_id: 3 }, origin: 'https://game.test', connect: value => { url = value; return socket }, refresh() {}, status() {} })
  assert.equal(url, 'wss://game.test/api/battles/events?account_id=7&character_id=3')
  connection.stop()
})
