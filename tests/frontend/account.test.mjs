import assert from 'node:assert/strict'
import test from 'node:test'
import { AccountStore } from '../../frontend/src/entities/account/store.ts'

const user = { id: 1, email: 'a@example.com', login: 'Alice', character_id: null }
const character = { id: 7, name: 'Алиса', faction_id: 1, class_id: 2, role: 'dd', level: 1, experience: 0, avatar: null, hp: 10, mp: 5, hp_current: 10, mp_current: 5 }
const deferred = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
function setup(overrides = {}) {
  return new AccountStore({ currentUser: async () => ({ user }), listCharacters: async () => ({ characters: [character], active_character_id: null }), login: async () => ({ user }), logout: async () => ({}), resetPassword: async () => ({ message: 'ok' }), setActiveCharacter: async (_accountId, id) => ({ character: id === null ? null : character }), createCharacter: async () => ({ character }), characterCatalog: async () => ({ factions: [], classes: [] }), ...overrides })
}

test('sign-in forwards an explicit persistence choice and defaults to a session cookie', async () => {
  const calls = []
  const store = setup({ login: async (...args) => { calls.push(args); return { user } } })
  await store.signIn('a@example.com', 'password')
  await store.signIn('a@example.com', 'password', true)
  await store.signIn('a@example.com', 'password', false)
  assert.deepEqual(calls.map(args => args.slice(0, 3)), [false, true, false].map(flag => ['a@example.com', 'password', flag]))
  assert.equal(store.getSnapshot().status, 'ready')
})
test('initial roster and cross-device refresh derive the account choice from the same server snapshot', async () => {
  let active = null
  const store = setup({ listCharacters: async () => ({ characters: [character], active_character_id: active }) })
  await store.refresh()
  assert.equal(store.getSnapshot().status, 'ready')
  assert.equal(store.getSnapshot().user.character_id, null)
  active = 7
  await store.refresh()
  assert.equal(store.getSnapshot().user.character_id, 7)
})
test('a roster started before selection cannot undo it', async () => {
  const old = deferred()
  let calls = 0
  const store = setup({ listCharacters: () => ++calls === 2 ? old.promise : Promise.resolve({ characters: [character], active_character_id: null }) })
  await store.refresh()
  const refresh = store.refresh()
  await Promise.resolve()
  await store.select(7)
  old.resolve({ characters: [character], active_character_id: null })
  await refresh
  assert.equal(store.getSnapshot().user.character_id, 7)
  await store.select(null)
  assert.equal(store.getSnapshot().user.character_id, null)
})
test('an old unauthorized response cannot clear a later login', async () => {
  const old = deferred()
  let calls = 0
  const store = setup({ currentUser: () => ++calls === 1 ? old.promise : Promise.resolve({ user }) })
  const refresh = store.refresh()
  await store.signIn('a@example.com', 'password')
  old.reject({ response: { status: 401 } })
  await refresh
  assert.equal(store.getSnapshot().user.id, 1)
})
test('logout invalidates outstanding roster data', async () => {
  const old = deferred()
  const store = setup({ listCharacters: () => old.promise })
  const refresh = store.refresh()
  await Promise.resolve()
  await store.signOut()
  old.resolve({ characters: [character], active_character_id: 7 })
  await refresh
  assert.equal(store.getSnapshot().status, 'anonymous')
  assert.equal(store.getSnapshot().user, null)
})
test('refresh failures hide stale game data but do not pretend the account is logged out', async () => {
  let fails = false
  const store = setup({ listCharacters: async () => { if (fails) throw new Error('offline'); return { characters: [character], active_character_id: 7 } } })
  await store.refresh()
  fails = true
  await store.refresh()
  assert.equal(store.getSnapshot().status, 'error')
  assert.equal(store.getSnapshot().user.id, 1)
  assert.equal(store.getSnapshot().roster, null)
})
test('unauthorized character operations expire the shared session', async () => {
  const store = setup({ setActiveCharacter: async () => { throw { response: { status: 401 } } } })
  await store.refresh()
  await assert.rejects(store.select(7))
  assert.equal(store.getSnapshot().status, 'anonymous')
})
test('creation adds to the roster without selecting and pending mutations cannot duplicate', async () => {
  const request = deferred()
  let calls = 0
  const store = setup({ createCharacter: () => { calls++; return request.promise } })
  await store.refresh()
  const first = store.create({ name: 'Алиса', faction_id: 1, class_id: 2 })
  await assert.rejects(store.create({ name: 'Алиса', faction_id: 1, class_id: 2 }))
  request.resolve({ character: { ...character, id: 8 } })
  await first
  assert.equal(calls, 1)
  assert.equal(store.getSnapshot().user.character_id, null)
  assert.equal(store.getSnapshot().roster.characters.length, 2)
})
test('post-login refresh wins when an earlier roster arrives out of order', async () => {
  const old = deferred()
  let calls = 0
  const store = setup({ listCharacters: () => ++calls === 1 ? old.promise : Promise.resolve({ characters: [character], active_character_id: 7 }) })
  const first = store.refresh()
  await Promise.resolve()
  await store.signIn('a@example.com', 'password')
  old.resolve({ characters: [], active_character_id: null })
  await first
  assert.equal(store.getSnapshot().user.character_id, 7)
  assert.equal(store.getSnapshot().roster.characters.length, 1)
})
test('repeated refreshes share in-flight account and roster requests until ready', async () => {
  const account = deferred()
  const roster = deferred()
  let accountCalls = 0
  let rosterCalls = 0
  const store = setup({
    currentUser: () => { accountCalls++; return account.promise },
    listCharacters: () => { rosterCalls++; return roster.promise },
  })
  const refreshes = [store.refresh(), store.refresh(), store.refresh()]
  account.resolve({ user })
  await Promise.resolve()
  refreshes.push(store.refresh(), store.refresh())
  roster.resolve({ characters: [character], active_character_id: 7 })
  await Promise.all(refreshes)
  assert.equal(accountCalls, 1)
  assert.equal(rosterCalls, 1)
  assert.equal(store.getSnapshot().status, 'ready')
  assert.equal(store.getSnapshot().user.character_id, 7)
})
test('completion of an invalidated refresh does not release the newer login refresh', async () => {
  const oldRoster = deferred()
  const newRoster = deferred()
  let accountCalls = 0
  let rosterCalls = 0
  const store = setup({
    currentUser: async () => { accountCalls++; return { user } },
    listCharacters: () => ++rosterCalls === 1 ? oldRoster.promise : newRoster.promise,
  })
  const oldRefresh = store.refresh()
  await Promise.resolve()
  const login = store.signIn('a@example.com', 'password')
  await new Promise(resolve => setImmediate(resolve))
  oldRoster.resolve({ characters: [], active_character_id: null })
  await oldRefresh
  const repeated = store.refresh()
  newRoster.resolve({ characters: [character], active_character_id: 7 })
  await Promise.all([login, repeated])
  assert.equal(accountCalls, 2)
  assert.equal(rosterCalls, 2)
  assert.equal(store.getSnapshot().user.character_id, 7)
})
test('initial network failure is an error; a valid empty roster remains signed in', async () => {
  let fails = true
  const store = setup({ listCharacters: async () => { if (fails) throw new Error('offline'); return { characters: [], active_character_id: null } } })
  await store.refresh()
  assert.equal(store.getSnapshot().status, 'error')
  assert.equal(store.getSnapshot().roster, null)
  fails = false
  await store.refresh()
  assert.equal(store.getSnapshot().status, 'ready')
  assert.deepEqual(store.getSnapshot().roster.characters, [])
})
test('catalog unauthorized response expires the session', async () => {
  const store = setup({ characterCatalog: async () => { throw { response: { status: 401 } } } })
  await store.refresh()
  await assert.rejects(store.catalog())
  assert.equal(store.getSnapshot().status, 'anonymous')
})
test('a same-account background refresh does not reject an in-flight catalog', async () => {
  const request = deferred()
  const store = setup({ characterCatalog: () => request.promise })
  await store.refresh()
  const catalog = store.catalog()
  await store.refresh()
  request.resolve({ factions: [{ id: 5, name: 'Republic' }], classes: [] })
  assert.deepEqual(await catalog, { factions: [{ id: 5, name: 'Republic' }], classes: [] })
})
test('catalog responses stay invalid after a silent account switch and return to the original account', async () => {
  for (const rejected of [false, true]) {
    const request = deferred()
    let current = user
    const store = setup({ currentUser: async () => ({ user: current }), characterCatalog: () => request.promise })
    await store.refresh()
    const catalog = store.catalog()
    const rejection = assert.rejects(catalog)
    current = { ...user, id: 2 }
    await store.refresh()
    current = user
    await store.refresh()
    if (rejected) request.reject({ response: { status: 401 } })
    else request.resolve({ factions: [], classes: [] })
    await rejection
    assert.equal(store.getSnapshot().status, 'ready')
    assert.equal(store.getSnapshot().user.id, 1)
  }
})
test('a current catalog 401 still expires the session after a same-account refresh', async () => {
  const request = deferred()
  const store = setup({ characterCatalog: () => request.promise })
  await store.refresh()
  const rejected = assert.rejects(store.catalog())
  await store.refresh()
  request.reject({ response: { status: 401 } })
  await rejected
  assert.equal(store.getSnapshot().status, 'anonymous')
})
test('failed mutations preserve the current selection and allow retry', async () => {
  let fails = true
  const store = setup({ setActiveCharacter: async () => { if (fails) throw { response: { status: 500 } }; return { character } } })
  await store.refresh()
  await assert.rejects(store.select(7))
  assert.equal(store.getSnapshot().user.character_id, null)
  assert.equal(store.getSnapshot().pending, false)
  fails = false
  await store.select(7)
  assert.equal(store.getSnapshot().user.character_id, 7)
})

test('explicit invalid login credentials preserve an existing valid account', async () => {
  const store = setup({ login: async () => { throw { response: { status: 401, data: { code: 'invalid_credentials', error: 'Wrong credentials' } } } } })
  await store.refresh()
  const before = store.getSnapshot()
  await assert.rejects(store.signIn('other@example.com', 'wrong'))
  assert.equal(store.getSnapshot().status, 'ready')
  assert.equal(store.getSnapshot().user, before.user)
  assert.equal(store.getSnapshot().roster, before.roster)
  assert.equal(store.getSnapshot().pending, false)
})

test('explicit unauthenticated failures expire the shared session', async () => {
  const store = setup({ setActiveCharacter: async () => { throw { response: { status: 401, data: { code: 'unauthenticated' } } } } })
  await store.refresh()
  await assert.rejects(store.select(7))
  assert.equal(store.getSnapshot().status, 'anonymous')
  assert.equal(store.getSnapshot().user, null)
})

test('password reset forwards token and password, then expires the current account', async () => {
  const calls = []
  const store = setup({ resetPassword: async (...args) => { calls.push(args); return { message: 'ok' } } })
  await store.refresh()
  await store.resetPassword('reset-token', 'new-password')
  assert.deepEqual(calls.map(args => args.slice(0, 2)), [['reset-token', 'new-password']])
  assert.equal(store.getSnapshot().status, 'anonymous')
  assert.equal(store.getSnapshot().user, null)
  assert.equal(store.getSnapshot().roster, null)
})

test('password reset exposes pending state and preserves the account on failure', async () => {
  const request = deferred()
  const store = setup({ resetPassword: () => request.promise })
  await store.refresh()
  const before = store.getSnapshot()
  const reset = store.resetPassword('reset-token', 'new-password')
  assert.equal(store.getSnapshot().pending, true)
  request.reject({ response: { status: 400, data: { code: 'password_reset_token_invalid', error: 'Invalid token' } } })
  await assert.rejects(reset)
  assert.equal(store.getSnapshot().status, 'ready')
  assert.equal(store.getSnapshot().user, before.user)
  assert.equal(store.getSnapshot().roster, before.roster)
  assert.equal(store.getSnapshot().pending, false)
})

test('password reset invalidates a refresh started before session revocation', async () => {
  const old = deferred()
  let calls = 0
  const store = setup({ currentUser: () => ++calls === 1 ? old.promise : Promise.resolve({ user }) })
  const refresh = store.refresh()
  await store.resetPassword('reset-token', 'new-password')
  old.resolve({ user })
  await refresh
  assert.equal(store.getSnapshot().status, 'anonymous')
  assert.equal(store.getSnapshot().user, null)
})

test('roster pins the discovered account and account_changed reboots without publishing mixed state', async () => {
  const seen = []
  let current = user
  const store = setup({
    currentUser: async () => ({ user: current }),
    listCharacters: async accountId => {
      seen.push(accountId)
      if (seen.length === 1) { current = { ...user, id: 2 }; throw { response: { status: 409, data: { code: 'account_changed' } } } }
      return { characters: [], active_character_id: null }
    },
  })
  const snapshots = []
  store.subscribe(() => snapshots.push(store.getSnapshot()))
  await store.refresh()
  assert.deepEqual(seen, [1, 2])
  assert.equal(store.getSnapshot().user.id, 2)
  assert(snapshots.some(s => s.user === null && s.roster === null))
})

test('external auth invalidates mutation, aborts it and old finally cannot release a newer pending action', async () => {
  const old = deferred(), fresh = deferred()
  let current = user
  const calls = []
  const store = setup({
    currentUser: async () => ({ user: current }),
    createCharacter: (accountId, input, signal) => { calls.push({ accountId, signal }); return calls.length === 1 ? old.promise : fresh.promise },
  })
  await store.refresh()
  const first = store.create({ name: 'Old', faction_id: 1, class_id: 2 })
  await Promise.resolve()
  current = { ...user, id: 2 }
  await store.externalAuthChanged()
  assert.equal(calls[0].signal.aborted, true)
  const second = store.create({ name: 'New', faction_id: 1, class_id: 2 })
  old.resolve({ character: { ...character, id: 88 } })
  await assert.rejects(first)
  assert.equal(store.getSnapshot().pending, true)
  assert.equal(store.getSnapshot().user.id, 2)
  fresh.resolve({ character: { ...character, id: 99 } })
  await second
  assert.deepEqual(calls.map(c => c.accountId), [1, 2])
  assert(!store.getSnapshot().roster.characters.some(c => c.id === 88))
})

test('guarded mutation captures displayed account and mismatch refreshes without replay', async () => {
  let current = user
  const calls = []
  const store = setup({ currentUser: async () => ({ user: current }), setActiveCharacter: async (...args) => {
    calls.push(args); current = { ...user, id: 2 }
    throw { response: { status: 409, data: { code: 'account_changed' } } }
  } })
  await store.refresh()
  await assert.rejects(store.select(7))
  assert.equal(calls.length, 1)
  assert.equal(calls[0][0], 1)
  assert.equal(store.getSnapshot().user.id, 2)
  assert.equal(store.getSnapshot().pending, false)
})

test('successful auth broadcasts once and external refresh does not rebroadcast', async () => {
  let changes = 0
  const store = setup()
  const release = store.subscribeAuthChanges(() => changes++)
  await store.signIn('a', 'b')
  await store.externalAuthChanged()
  await store.signOut()
  await store.resetPassword('token', 'password')
  assert.equal(changes, 3)
  release()
  await store.signIn('a', 'b')
  assert.equal(changes, 3)
})


test('all guarded operations carry the displayed account, unknown state bootstraps and anonymous logout stays local', async () => {
  const calls = []
  const store = setup({
    characterCatalog: async accountId => { calls.push(['catalog', accountId]); return { factions: [], classes: [] } },
    setActiveCharacter: async (accountId, id) => { calls.push(['select', accountId, id]); return { character: null } },
    createCharacter: async (accountId, input) => { calls.push(['create', accountId, input.name]); return { character } },
    logout: async accountId => { calls.push(['logout', accountId]) },
  })
  await store.catalog()
  await store.select(null)
  await store.create({ name: 'Name', faction_id: 1, class_id: 2 })
  await store.signOut()
  await store.signOut()
  assert.deepEqual(calls, [['catalog', 1], ['select', 1, null], ['create', 1, 'Name'], ['logout', 1]])
})

test('failed authentication operations never broadcast', async () => {
  const failure = async () => { throw { response: { status: 500 } } }
  const store = setup({ login: failure, logout: failure, resetPassword: failure })
  let broadcasts = 0
  store.subscribeAuthChanges(() => broadcasts++)
  await store.refresh()
  await assert.rejects(store.signIn('a', 'b'))
  await assert.rejects(store.signOut())
  await assert.rejects(store.resetPassword('t', 'p'))
  assert.equal(broadcasts, 0)
  assert.equal(store.getSnapshot().pending, false)
})
