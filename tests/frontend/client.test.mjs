import assert from 'node:assert/strict'
import test from 'node:test'
import http from 'node:http'
import { apiRequest } from '../../frontend/src/shared/api/client.ts'
import { apiError } from '../../frontend/src/shared/api/errors.ts'
import { AccountStore } from '../../frontend/src/entities/account/store.ts'

test('client forwards typed cursor queries and the expected account header', async () => {
  let received
  const server = http.createServer((request, response) => {
    received = { url: request.url, account: request.headers['x-account-id'] }
    response.end(JSON.stringify({ messages: [], next_before_id: null, chat_ban: null }))
  })
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve))
  try {
    await apiRequest('get', `http://127.0.0.1:${server.address().port}/api/chat/messages`, { query: { before_id: 17 }, headers: { 'X-Account-ID': 9 } })
    assert.deepEqual(received, { url: '/api/chat/messages?before_id=17', account: '9' })
  } finally { server.closeAllConnections(); await new Promise(resolve => server.close(resolve)) }
})

test('actual client bounds a stalled transport at 20 seconds, preserves caller abort and permits manual retry', async () => {
  const nativeTimer = globalThis.setTimeout
  const nativeClear = globalThis.clearTimeout
  const activeTimers = new Set()
  globalThis.clearTimeout = timer => { activeTimers.delete(timer); return nativeClear(timer) }
  const deadlines = []
  globalThis.setTimeout = (fn, ms, ...args) => {
    if (ms === 20000) { deadlines.push(ms); ms = 40 }
    const timer = nativeTimer(fn, ms, ...args)
    if (ms === 40) activeTimers.add(timer)
    return timer
  }
  let calls = 0
  const server = http.createServer((request, response) => {
    calls++
    if (request.url === '/ok') response.end(JSON.stringify({ user: { id: 1 } }))
  })
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve))
  const base = `http://127.0.0.1:${server.address().port}`
  try {
    const start = Date.now()
    await assert.rejects(Promise.race([apiRequest('get', base + '/stall'), new Promise((_, reject) => nativeTimer(() => reject(new Error('transport remained unbounded')), 300))]), error => {
      assert.match(apiError(error).message, /Время ожидания/)
      return true
    })
    assert(Date.now() - start < 1000)
    assert(deadlines.length > 0, 'production deadline must be 20_000ms')
    assert.equal(calls, 1, 'deadline does not retry')
    assert.equal(activeTimers.size, 0, 'deadline timer cleaned')
    assert.equal((await apiRequest('get', base + '/ok')).user.id, 1)
    assert.equal(activeTimers.size, 0, 'success timer cleaned')
    const controller = new AbortController()
    let listeners = 0
    const add = controller.signal.addEventListener.bind(controller.signal)
    const remove = controller.signal.removeEventListener.bind(controller.signal)
    controller.signal.addEventListener = (...args) => { listeners++; return add(...args) }
    controller.signal.removeEventListener = (...args) => { listeners--; return remove(...args) }
    const aborted = apiRequest('get', base + '/stall', { signal: controller.signal })
    controller.abort()
    await assert.rejects(aborted, error => error.code === 'ERR_CANCELED')
    assert.equal(listeners, 0, 'caller listener cleaned')
    assert.equal(activeTimers.size, 0, 'abort timer cleaned')
    const already = new AbortController(); already.abort()
    await assert.rejects(apiRequest('get', base + '/ok', { signal: already.signal }), error => error.code === 'ERR_CANCELED')
  } finally {
    globalThis.setTimeout = nativeTimer
    globalThis.clearTimeout = nativeClear
    server.closeAllConnections()
    await new Promise(resolve => server.close(resolve))
  }
})

test('mutation deadline frees pending and never resends the operation', async () => {
  const nativeTimer = globalThis.setTimeout
  globalThis.setTimeout = (fn, ms, ...args) => nativeTimer(fn, ms === 20000 ? 30 : ms, ...args)
  let requests = 0
  const server = http.createServer((request, response) => {
    requests++
    if (requests === 2) response.end(JSON.stringify({ character: null }))
  })
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve))
  const user = { id: 1, character_id: null }
  const store = new AccountStore({
    currentUser: async () => ({ user }),
    listCharacters: async () => ({ characters: [], active_character_id: null }),
    setActiveCharacter: () => apiRequest('put', `http://127.0.0.1:${server.address().port}/stall`, { body: { character_id: null } }),
  })
  try {
    await store.refresh()
    await assert.rejects(Promise.race([store.select(null), new Promise((_, reject) => nativeTimer(() => reject(new Error('unbounded mutation')), 300))]))
    assert.equal(store.getSnapshot().pending, false)
    assert.equal(requests, 1)
    await store.select(null)
    assert.equal(requests, 2)
  } finally {
    globalThis.setTimeout = nativeTimer
    server.closeAllConnections()
    await new Promise(resolve => server.close(resolve))
  }
})
