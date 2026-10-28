import assert from 'node:assert/strict'
import test from 'node:test'
import { SnapshotFeed, connectChat } from '../../frontend/src/features/social/model/lifecycle.ts'
import { canModerate, canSanction, messageLength, validReason } from '../../frontend/src/features/social/model/policy.ts'
import { SocialSession } from '../../frontend/src/features/social/model/session.ts'

const deferred = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
const tick = async () => { for (let i = 0; i < 8; i++) await Promise.resolve() }

test('notifications coalesce and refetch the displayed history cursor', async () => {
  const first = deferred(), calls = [], values = []
  const feed = new SnapshotFeed((cursor, signal) => { calls.push({ cursor, signal }); return calls.length === 1 ? first.promise : Promise.resolve('fresh') }, value => values.push(value), assert.fail)
  feed.page(80)
  feed.refresh(); feed.refresh(); feed.refresh()
  assert.equal(calls.length, 1)
  first.resolve('old'); await tick()
  assert.deepEqual(calls.map(item => item.cursor), [80, 80])
  assert.deepEqual(values, ['old', 'fresh'])
  feed.stop()
})

test('changing pages aborts old snapshots and ignores late completion and errors after stop', async () => {
  const old = deferred(), current = deferred(), calls = [], values = [], errors = []
  const feed = new SnapshotFeed((cursor, signal) => { calls.push({ cursor, signal }); return cursor === 70 ? old.promise : current.promise }, value => values.push(value), error => errors.push(error))
  feed.page(70); feed.page(null)
  assert.equal(calls[0].signal.aborted, true)
  old.resolve('obsolete'); await tick()
  assert.deepEqual(values, [])
  feed.stop(); current.reject(new Error('late 401')); await tick()
  assert.deepEqual(errors, []); assert.equal(calls[1].signal.aborted, true)
})

test('connection changes trigger authorized refresh, bounded retries, and cleanup closes the context', () => {
  const sockets = [], scheduled = [], statuses = []; let refreshes = 0
  const connect = () => { const socket = { close() { this.closed = true } }; sockets.push(socket); return socket }
  const connection = connectChat({ accountId: 17, origin: 'https://game.test', connect, refresh: () => refreshes++, status: status => statuses.push(status), schedule: (callback, delay) => { const task = { callback, delay }; scheduled.push(task); return task }, cancel: task => { task.cancelled = true } })
  sockets[0].onopen(); sockets[0].onmessage({ data: '{"type":"changed"}' }); sockets[0].onmessage({ data: 'invalid' })
  assert.equal(refreshes, 2)
  for (let i = 0; i < 8; i++) { sockets.at(-1).onclose(); scheduled.at(-1).callback() }
  assert.ok(scheduled.every(task => task.delay <= 15000)); assert.equal(scheduled.at(-1).delay, 15000)
  connection.stop(); const before = refreshes; sockets.at(-1).onmessage({ data: '{"type":"changed"}' }); sockets.at(-1).onclose()
  assert.equal(refreshes, before); assert.equal(sockets.at(-1).closed, true)
  assert.ok(statuses.includes('reconnecting'))
})

test('events from a replaced connection cannot change or close its replacement', () => {
  const sockets = [], statuses = []; let refreshes = 0
  const connection = connectChat({ accountId: 1, origin: 'http://game.test', refresh: () => refreshes++, status: value => statuses.push(value), connect: () => {
    const socket = { close() { this.closed = true } }; sockets.push(socket); return socket
  } })
  const oldOpen = sockets[0].onopen, oldMessage = sockets[0].onmessage, oldError = sockets[0].onerror
  connection.retry(); const before = statuses.length
  oldOpen(); oldMessage({ data: '{"type":"changed"}' }); oldError()
  assert.equal(refreshes, 0); assert.equal(statuses.length, before); assert.equal(sockets[1].closed, undefined)
  connection.stop()
})

test('moderation permissions protect self and privileged accounts; text counts Unicode code points', () => {
  assert.equal(canModerate('user'), false); assert.equal(canModerate('moderator'), true)
  assert.equal(canSanction('moderator', 1, 2, 'administrator'), false)
  assert.equal(canSanction('moderator', 1, 2, 'user'), true)
  assert.equal(canSanction('administrator', 1, 1, 'user'), false)
  assert.equal(canSanction('administrator', 1, 2, 'administrator'), true)
  assert.equal(messageLength(' 😀😀 '), 2)
  assert.equal(validReason(' 😀😀😀 '), true); assert.equal(validReason('a'), false)
})

test('late unauthorized mutation cannot refresh or alter a replacement account', async () => {
  let snapshot = { user: { id: 1 }, roster: { active_character_id: 7 } }, refreshes = 0
  const store = { getSnapshot: () => snapshot, refresh: async () => { refreshes++ }, externalAuthChanged: async () => { refreshes++ } }
  const session = new SocialSession(store, 1, 7); session.start()
  const pending = deferred(), applied = [], errors = []
  const operation = session.run(() => pending.promise, value => applied.push(value), error => errors.push(error))
  snapshot = { user: { id: 2 }, roster: { active_character_id: 9 } }
  pending.reject({ response: { status: 401 } }); await operation
  assert.equal(refreshes, 0); assert.deepEqual(errors, []); assert.deepEqual(applied, [])
  session.stop()
})

test('character switch rejects old success; unmount aborts pending mutations and prevents retries', async () => {
  let snapshot = { user: { id: 1 }, roster: { active_character_id: 7 } }
  const store = { getSnapshot: () => snapshot, refresh: async () => {}, externalAuthChanged: async () => {} }
  const session = new SocialSession(store, 1, 7); session.start()
  const pending = deferred(), applied = []; let signal, calls = 0
  const operation = session.run(contextSignal => { signal = contextSignal; calls++; return pending.promise }, value => applied.push(value), assert.fail)
  snapshot = { user: { id: 1 }, roster: { active_character_id: 8 } }
  session.stop(); assert.equal(signal.aborted, true)
  pending.resolve('old message'); await operation
  assert.deepEqual(applied, [])
  await session.run(() => { calls++; return Promise.resolve('unexpected') }, assert.fail, assert.fail)
  assert.equal(calls, 1)
})

test('current authentication and character conflicts refresh context once without retrying writes', async () => {
  let refreshes = 0, external = 0, calls = 0
  const store = { getSnapshot: () => ({ user: { id: 1 }, roster: { active_character_id: 7 } }), refresh: async () => { refreshes++ }, externalAuthChanged: async () => { external++ } }
  const session = new SocialSession(store, 1, 7); session.start()
  const errors = []
  for (const failure of [{ status: 401 }, { status: 409, data: { code: 'character_changed' } }, { status: 409, data: { code: 'account_changed' } }, { status: 429 }]) {
    await session.run(async () => { calls++; throw { response: failure } }, assert.fail, error => errors.push(error))
  }
  assert.equal(refreshes, 2); assert.equal(external, 1); assert.equal(calls, 4); assert.equal(errors.length, 4)
  session.stop()
})

test('cleared active character response refreshes the roster without replaying the message', async () => {
  let snapshot = { user: { id: 1 }, roster: { active_character_id: 7 } }, refreshes = 0, writes = 0
  const store = {
    getSnapshot: () => snapshot,
    refresh: async () => { refreshes++; snapshot = { user: { id: 1 }, roster: { active_character_id: null } } },
    externalAuthChanged: async () => assert.fail('character conflict must not reset authentication'),
  }
  const session = new SocialSession(store, 1, 7); session.start()
  const errors = []
  await session.run(async () => {
    writes++
    throw { response: { status: 409, data: { code: 'character_required', error: 'Выберите персонажа.' } } }
  }, assert.fail, failure => errors.push(failure))
  assert.equal(refreshes, 1)
  assert.equal(store.getSnapshot().roster.active_character_id, null)
  assert.equal(session.current(), false)
  assert.equal(writes, 1)
  assert.equal(errors[0].code, 'character_required')
  session.stop()
})
