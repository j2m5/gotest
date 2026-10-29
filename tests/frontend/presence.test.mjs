import assert from 'node:assert/strict'
import test from 'node:test'
import { PresenceStore } from '../../frontend/src/entities/presence/store.ts'
import { connectPresenceActivity } from '../../frontend/src/entities/presence/activity.ts'
import { connectPresenceAccount } from '../../frontend/src/entities/presence/account.ts'

const tick = async () => { for (let i = 0; i < 15; i++) await Promise.resolve() }
const deferred = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
function harness(overrides = {}) {
  let now = 0, id = 0
  const tasks = new Map(), writes = [], reads = []
  const clock = { now: () => now, schedule: (fn, delay) => { tasks.set(++id, { fn, at: now + delay }); return id }, cancel: id => tasks.delete(id) }
  const api = {
    heartbeat: async (account, character, signal) => { writes.push({ account, character, signal }) },
    statuses: async (account, ids, signal) => { reads.push({ account, ids, signal }); return { total: 2, characters: ids.map(character_id => ({ character_id, is_online: true })) } },
    ...overrides,
  }
  const store = new PresenceStore(api, clock)
  const advance = async ms => { const end = now + ms; while (true) { const next = [...tasks].filter(([, t]) => t.at <= end).sort((a, b) => a[1].at - b[1].at)[0]; if (!next) break; now = next[1].at; tasks.delete(next[0]); next[1].fn(); await tick() } now = end; await tick() }
  return { store, writes, reads, tasks, advance, jump: ms => { now += ms } }
}

test('passive polling never renews activity, while input has a leading and trailing write', async () => {
  const h = harness(); h.store.start(1, 7); await tick()
  await h.advance(60000); assert.equal(h.writes.length, 0)
  h.store.activity(); await tick(); assert.equal(h.writes.length, 1)
  await h.advance(1000); h.store.activity(); await h.advance(1000); h.store.activity()
  await h.advance(12999); assert.equal(h.writes.length, 1)
  await h.advance(1); assert.equal(h.writes.length, 2)
  await h.advance(180000); assert.equal(h.writes.length, 2)
  h.store.stop(); assert.equal(h.tasks.size, 0)
})

test('input during an in-flight heartbeat survives; focus and visibility coalesce', async () => {
  const pending = deferred(); let calls = 0
  const h = harness({ heartbeat: () => { calls++; return calls === 1 ? pending.promise : Promise.resolve() } })
  h.store.start(1, 7); h.store.activity(true); h.store.activity(true)
  await h.advance(1000); h.store.activity(); await h.advance(15000)
  assert.equal(calls, 1); pending.resolve(); await tick(); assert.equal(calls, 2)
  h.store.stop()
})

test('watchers deduplicate and split requests at 100 IDs; releasing a watcher removes its IDs', async () => {
  const h = harness(); const release = h.store.watch(Array.from({ length: 215 }, (_, i) => i + 1)); const other = h.store.watch([1, 216])
  h.store.start(1, null); await tick()
  assert.deepEqual(h.reads.map(r => r.ids.length), [100, 100, 16])
  assert.equal(h.store.getSnapshot().statuses[216], true)
  release(); h.reads.length = 0; await h.advance(30000)
  assert.deepEqual(h.reads.flatMap(r => r.ids), [1, 216]); other(); h.store.stop()
})

test('replacement aborts requests and rejects late results and auth failures', async () => {
  const old = deferred(), calls = []; let refreshes = 0
  const h = harness({ statuses: (account, ids, signal) => { calls.push({ account, signal }); return account === 1 ? old.promise : Promise.resolve({ total: 0, characters: [{ character_id: 7, is_online: false }] }) } })
  h.store.watch([7]); h.store.start(1, 7, () => refreshes++); await tick()
  h.store.start(2, 8, () => refreshes++); await tick(); assert.equal(calls[0].signal.aborted, true)
  old.reject({ response: { status: 401 } }); await tick()
  assert.equal(refreshes, 0); assert.equal(h.store.getSnapshot().statuses[7], false)
  h.store.stop()
})

test('failed and suspended snapshots become unknown, controlled auth refresh runs once', async () => {
  let fail = false, refreshes = 0
  const h = harness({ statuses: async (_, ids) => { if (fail) throw { response: { status: 401 } }; return { total: 1, characters: ids.map(character_id => ({ character_id, is_online: true })) } } })
  h.store.watch([7]); h.store.start(1, 7, () => refreshes++); await tick()
  assert.equal(h.store.getSnapshot().statuses[7], true)
  h.jump(121000); fail = true; h.store.refresh(); assert.equal(h.store.getSnapshot().statuses[7], undefined); await tick()
  assert.equal(h.store.getSnapshot().total, null); assert.equal(refreshes, 1)
  await h.advance(60000); assert.equal(refreshes, 1); h.store.stop()
})

test('visible browser actions count, background input and removed listeners do not', () => {
  const browser = new EventTarget(), page = new EventTarget(), actions = []
  page.visibilityState = 'visible'
  const release = connectPresenceActivity(value => actions.push(Boolean(value)), browser, page)
  page.dispatchEvent(new Event('scroll')); assert.deepEqual(actions, [], 'programmatic scrolling is passive')
  page.dispatchEvent(new Event('pointerdown')); page.dispatchEvent(new Event('keydown')); page.dispatchEvent(new Event('wheel')); page.dispatchEvent(new Event('touchmove'))
  page.dispatchEvent(new Event('scroll')); page.dispatchEvent(new Event('mousemove')); assert.deepEqual(actions, [false, false, false, false])
  page.visibilityState = 'hidden'; page.dispatchEvent(new Event('visibilitychange')); page.dispatchEvent(new Event('pointerdown')); browser.dispatchEvent(new Event('focus'))
  assert.equal(actions.length, 4)
  page.visibilityState = 'visible'; page.dispatchEvent(new Event('visibilitychange')); browser.dispatchEvent(new Event('focus'))
  assert.deepEqual(actions.slice(4), [true, true])
  release(); page.dispatchEvent(new Event('keydown')); page.dispatchEvent(new Event('wheel')); page.dispatchEvent(new Event('touchmove')); browser.dispatchEvent(new Event('focus')); assert.equal(actions.length, 6)
})

test('continuous input keeps read and write traffic bounded and replacement clears trailing timers', async () => {
  const h = harness(); h.store.start(1, 7); await tick(); h.reads.length = 0
  for (let i = 0; i < 60; i++) { h.store.activity(); await h.advance(100) }
  assert.equal(h.writes.length, 1); assert.ok(h.reads.length <= 2)
  h.store.start(1, 8); await h.advance(20000)
  assert.equal(h.writes.length, 1); h.store.activity(); await tick()
  assert.equal(h.writes[1].character, 8); h.store.stop()
})

test('passive account refresh and error recovery do not invent player activity', async () => {
  const h = harness(); let listener, now = 0
  let snapshot = { status: 'ready', user: { id: 1 }, roster: { active_character_id: 7 } }
  const account = { getSnapshot: () => snapshot, subscribe: value => { listener = value; return () => { listener = null } }, refresh: async () => {}, externalAuthChanged: async () => {} }
  const connection = connectPresenceAccount(account, h.store, () => true, () => now)
  await tick(); assert.equal(h.writes.length, 1)
  listener(); await tick(); assert.equal(h.writes.length, 1)
  snapshot = { ...snapshot, status: 'error' }; listener()
  snapshot = { ...snapshot, status: 'ready' }; listener(); await tick()
  assert.equal(h.writes.length, 1)
  now = 1000; connection.activity()
  snapshot = { ...snapshot, roster: { active_character_id: 8 } }; listener(); await tick()
  assert.equal(h.writes.at(-1).character, 8)
  connection.stop(); assert.equal(listener, null); assert.equal(h.tasks.size, 0)
})

test('foreground return sends promptly inside the input throttle and stop rejects late heartbeat errors', async () => {
  const pending = deferred(); let calls = 0, refreshes = 0
  const h = harness({ heartbeat: () => { calls++; return calls === 1 ? Promise.resolve() : pending.promise } })
  h.store.start(1, 7, () => refreshes++); h.store.activity(); await tick()
  await h.advance(2000); h.store.activity(true); h.store.activity(true)
  assert.equal(calls, 2)
  h.store.stop(); pending.reject({ response: { status: 401 } }); await tick()
  assert.equal(refreshes, 0); assert.equal(h.store.getSnapshot().total, null); assert.equal(h.tasks.size, 0)
})

test('late successful snapshots cannot republish a replaced character context', async () => {
  const old = deferred(); let reads = 0
  const h = harness({ statuses: async () => { reads++; return reads === 1 ? old.promise : { total: 0, characters: [{ character_id: 7, is_online: false }] } } })
  h.store.watch([7]); h.store.start(1, 7); h.store.start(1, 8); await tick()
  old.resolve({ total: 999, characters: [{ character_id: 7, is_online: true }] }); await tick()
  assert.equal(h.store.getSnapshot().total, 0); assert.equal(h.store.getSnapshot().statuses[7], false)
  h.store.stop()
})

test('a meaningful action near expiry reaches the server, then silence schedules no renewals', async () => {
  const h = harness(); h.store.start(1, 7); h.store.activity(); await tick()
  await h.advance(119000); assert.equal(h.writes.length, 1)
  h.store.activity(); await tick(); assert.equal(h.writes.length, 2)
  await h.advance(240000); assert.equal(h.writes.length, 2); h.store.stop()
})

test('resuming delayed timers discards expired input instead of renewing an idle character', async () => {
  const h = harness(); h.store.start(1, 7); h.store.activity(); await tick()
  await h.advance(1000); h.store.activity()
  h.jump(200000)
  // A frozen tab resumes overdue callbacks at its new time, not at their old deadlines.
  for (const [id, task] of [...h.tasks]) { h.tasks.delete(id); task.fn(); await tick() }
  assert.equal(h.writes.length, 1, 'no real input occurred during suspension')
  h.store.activity(true); await tick(); assert.equal(h.writes.length, 2, 'fresh foreground activity restores online')
  h.store.stop()
})

test('a heartbeat completing after suspension cannot flush expired queued input', async () => {
  const pending = deferred(); let calls = 0
  const h = harness({ heartbeat: () => { calls++; return calls === 1 ? pending.promise : Promise.resolve() } })
  h.store.start(1, 7); h.store.activity()
  await h.advance(1000); h.store.activity()
  h.jump(200000); pending.resolve(); await tick()
  assert.equal(calls, 1, 'old queued action must expire while request completion is suspended')
  h.store.activity(); await tick(); assert.equal(calls, 2)
  h.store.stop()
})
