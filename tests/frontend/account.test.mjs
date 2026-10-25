import assert from 'node:assert/strict'
import test from 'node:test'
import { AccountStore } from '../../frontend/src/account/store.ts'

const user = { id: 1, email: 'a@example.com', login: 'Alice', character_id: null }
const character = { id: 7, name: 'Алиса', faction_id: 1, class_id: 2, role: 'dd', level: 1, experience: 0, avatar: null, hp: 10, mp: 5, hp_current: 10, mp_current: 5 }
const deferred = () => { let resolve, reject; const promise = new Promise((a, b) => { resolve = a; reject = b }); return { promise, resolve, reject } }
function setup(overrides = {}) {
  return new AccountStore({ currentUser: async () => ({ user }), listCharacters: async () => ({ characters: [character], active_character_id: null }), login: async () => ({ user }), logout: async () => ({}), setActiveCharacter: async id => ({ character: id === null ? null : character }), createCharacter: async () => ({ character }), characterCatalog: async () => ({ factions: [], classes: [] }), ...overrides })
}
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
