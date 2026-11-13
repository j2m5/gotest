import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

// Go-driven tests attach to their parent's Chrome. Standalone scripts own theirs.
export async function launchChrome(binary, profile) {
  const browser = process.env.COMBAT_BROWSER_PARENT === '1' ? null : spawn(binary, ['--headless=new', '--remote-debugging-port=0', `--user-data-dir=${profile}`,
    '--disable-background-networking', '--no-first-run', '--no-default-browser-check', '--disable-extensions', '--disable-gpu', 'about:blank'],
  { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] })
  const exited = browser ? new Promise(resolve => browser.once('close', resolve)) : Promise.resolve()
  let socket, sequence = 0, log = '', startupError
  const requests = new Map(), listeners = new Set(), eventTimers = new Set(), exceptions = []
  browser?.once('error', error => { startupError = error })
  browser?.stderr.on('data', data => { log = (log + data).slice(-4000) })
  function command(method, params = {}, sessionId) {
    return new Promise((resolve, reject) => {
      const id = ++sequence, timer = setTimeout(() => { requests.delete(id); reject(new Error(`CDP timeout ${method}`)) }, 10000)
      requests.set(id, { resolve, reject, timer }); socket.send(JSON.stringify({ id, method, params, sessionId }))
    })
  }
  async function close() {
    if (socket?.readyState === WebSocket.OPEN) { await command('Browser.close').catch(() => {}); socket.close() }
    if (browser?.exitCode === null) browser.kill()
    await exited
    for (const pending of requests.values()) clearTimeout(pending.timer)
    for (const timer of eventTimers) clearTimeout(timer)
    listeners.clear()
  }
  try {
    let port, endpoint
    for (let i = 0; i < 150 && !endpoint; i++) {
      if (startupError) throw startupError
      try { [port, endpoint] = (await readFile(path.join(profile, 'DevToolsActivePort'), 'utf8')).trim().split('\n') } catch { await delay(100) }
    }
    assert(endpoint, 'Chrome unavailable: ' + log)
    socket = new WebSocket(`ws://127.0.0.1:${port}${endpoint}`)
    socket.addEventListener('message', ({ data }) => {
      const message = JSON.parse(data)
      if (message.method === 'Runtime.exceptionThrown') exceptions.push(message.params.exceptionDetails.exception?.description ?? message.params.exceptionDetails.text)
      if (message.method) for (const listener of [...listeners]) listener(message)
      const pending = requests.get(message.id)
      if (!pending) return
      clearTimeout(pending.timer); requests.delete(message.id)
      if (message.error) pending.reject(new Error(message.error.message)); else pending.resolve(message.result)
    })
    await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once: true }); socket.addEventListener('error', reject, { once: true }) })
    const on = listener => { listeners.add(listener); return () => listeners.delete(listener) }
    function nextEvent(sessionId, method, predicate = () => true) {
      const promise = new Promise((resolve, reject) => {
        const timer = setTimeout(() => { eventTimers.delete(timer); off(); reject(new Error('No event: ' + method)) }, 10000)
        eventTimers.add(timer)
        const off = on(event => {
          if (event.sessionId !== sessionId || event.method !== method || !predicate(event.params)) return
          clearTimeout(timer); eventTimers.delete(timer); off(); resolve(event.params)
        })
      })
      // A preceding failed step must not leave an unhandled timeout rejection.
      promise.catch(() => {})
      return promise
    }
    async function evaluate(sessionId, expression) {
      const result = await command('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true }, sessionId)
      if (result.exceptionDetails) throw new Error(result.exceptionDetails.exception?.description ?? result.exceptionDetails.text)
      return result.result.value
    }
    async function waitFor(sessionId, expression, label) {
      const deadline = Date.now() + 10000
      while (Date.now() < deadline) { if (await evaluate(sessionId, `Boolean(${expression})`)) return; await delay(50) }
      throw new Error(`${label}: ${await evaluate(sessionId, 'document.body.innerText')}`)
    }
    async function newTab(browserContextId) {
      browserContextId ??= (await command('Target.createBrowserContext')).browserContextId
      const { targetId } = await command('Target.createTarget', { url: 'about:blank', browserContextId })
      const { sessionId } = await command('Target.attachToTarget', { targetId, flatten: true })
      await command('Page.enable', {}, sessionId); await command('Runtime.enable', {}, sessionId); await command('Network.enable', {}, sessionId)
      await command('Emulation.setFocusEmulationEnabled', { enabled: true }, sessionId)
      return { sessionId, targetId, browserContextId }
    }
    async function press(sessionId, key) {
      const code = key === ' ' ? 'Space' : key, windowsVirtualKeyCode = key === ' ' ? 32 : 13
      await command('Input.dispatchKeyEvent', { type: 'keyDown', key, code, windowsVirtualKeyCode, text: key === 'Enter' ? '\r' : key }, sessionId)
      await command('Input.dispatchKeyEvent', { type: 'keyUp', key, code, windowsVirtualKeyCode }, sessionId)
    }
    return { command, evaluate, waitFor, newTab, nextEvent, on, press, close, exceptions }
  } catch (error) { await close(); throw error }
}
