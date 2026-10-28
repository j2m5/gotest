import assert from 'node:assert/strict'
import test from 'node:test'
import { connectAuthSync } from '../../frontend/src/entities/account/authSync.ts'

function fixture(BroadcastChannel) {
  const listeners = new Set(), auth = new Set(), stored = [], channels = []
  let refreshes = 0
  const host = {
    BroadcastChannel: BroadcastChannel && class {
      constructor(name) { this.name = name; channels.push(this) }
      postMessage(data) { this.sent = [...(this.sent ?? []), data] }
      close() { this.closed = true }
    },
    addEventListener: (name, fn) => { assert.equal(name, 'storage'); listeners.add(fn) },
    removeEventListener: (name, fn) => listeners.delete(fn),
    localStorage: { setItem: (...args) => stored.push(args) },
  }
  const store = {
    subscribeAuthChanges: fn => { auth.add(fn); return () => auth.delete(fn) },
    externalAuthChanged: async () => { refreshes++ },
  }
  const connect = () => connectAuthSync(store, host)
  const publish = () => auth.forEach(fn => fn())
  return { host, store, connect, publish, listeners, stored, channels, refreshes: () => refreshes }
}

test('auth channel sends no secrets, handles external changes without loops, releases and reconnects', async () => {
  const f = fixture(true)
  const release = f.connect()
  f.publish()
  const channel = f.channels[0]
  assert.deepEqual(channel.sent, [{ type: 'auth_changed' }])
  channel.onmessage({ data: { type: 'auth_changed' } })
  await Promise.resolve()
  assert.equal(f.refreshes(), 1)
  assert.equal(channel.sent.length, 1)
  release()
  assert.equal(channel.closed, true)
  assert.equal(f.listeners.size, 0)
  f.publish()
  assert.equal(channel.sent.length, 1)
  const releaseAgain = f.connect()
  f.publish()
  assert.equal(f.channels[1].sent.length, 1)
  releaseAgain()
})

test('storage fallback invalidates other tabs, tolerates blocked storage and cleans listeners', async () => {
  const f = fixture(false)
  const release = f.connect()
  f.publish()
  assert.equal(f.stored.length, 1)
  const [key, newValue] = f.stored[0]
  const payload = JSON.parse(newValue)
  assert.deepEqual(Object.keys(payload).sort(), ['nonce', 'type'])
  for (const fn of f.listeners) { fn({ key: 'unrelated', newValue }); fn({ key, newValue: null }); fn({ key, newValue: '{' }); fn({ key, newValue }) }
  await Promise.resolve()
  assert.equal(f.refreshes(), 1)
  assert.equal(f.stored.length, 1)
  f.host.localStorage.setItem = () => { throw new Error('denied') }
  assert.doesNotThrow(f.publish)
  release()
  assert.equal(f.listeners.size, 0)
})
