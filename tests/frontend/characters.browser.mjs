// Optional real-browser smoke test, launched by tests/db/character_browser_test.go.
// Node's built-in WebSocket talks to a disposable headless Chromium profile.
import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile, writeFile, mkdir } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

const [baseURL, browserPath, profilePath] = process.argv.slice(2)
assert(baseURL && browserPath && profilePath, 'Expected test server, browser and temporary profile')
const browser = spawn(browserPath, [
  '--headless=new', '--remote-debugging-port=0', `--user-data-dir=${profilePath}`,
  '--disable-background-networking', '--no-first-run', '--no-default-browser-check',
  '--disable-extensions', '--disable-gpu', 'about:blank',
], { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] })
let browserLog = ''
browser.stderr.on('data', chunk => { browserLog = (browserLog + chunk).slice(-6000) })
let socket
let nextID = 0
const pending = new Map()
const exceptions = []
const browserExited = new Promise(resolve => browser.once('exit', resolve))
let launchError
browser.on('error', error => { launchError = error })

function command(method, params = {}, sessionId) {
  return new Promise((resolve, reject) => {
    const id = ++nextID
    const timer = setTimeout(() => { pending.delete(id); reject(new Error(`CDP timeout: ${method}`)) }, 10000)
    pending.set(id, { resolve, reject, timer })
    socket.send(JSON.stringify({ id, method, params, sessionId }))
  })
}
async function evaluate(session, expression) {
  const result = await command('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true }, session)
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.exception?.description ?? result.exceptionDetails.text)
  return result.result.value
}
async function waitFor(session, expression, label) {
  const deadline = Date.now() + 10000
  while (Date.now() < deadline) {
    if (await evaluate(session, expression)) return
    await delay(75)
  }
  const page = await evaluate(session, '({path:location.pathname, text:document.body.innerText})')
  throw new Error(`Waiting for ${label}: ${JSON.stringify(page)}`)
}
async function tab(contextId) {
  const { targetId } = await command('Target.createTarget', { url: 'about:blank', browserContextId: contextId })
  const { sessionId } = await command('Target.attachToTarget', { targetId, flatten: true })
  await command('Page.enable', {}, sessionId)
  await command('Runtime.enable', {}, sessionId)
  await command('Emulation.setDeviceMetricsOverride', { width: 480, height: 900, deviceScaleFactor: 1, mobile: false }, sessionId)
  return sessionId
}
async function fill(session, selector, value) {
  await evaluate(session, `(() => {
    const element = document.querySelector(${JSON.stringify(selector)});
    const prototype = element.tagName === 'SELECT' ? HTMLSelectElement.prototype : HTMLInputElement.prototype;
    Object.getOwnPropertyDescriptor(prototype, 'value').set.call(element, ${JSON.stringify(value)});
    element.dispatchEvent(new Event(element.tagName === 'SELECT' ? 'change' : 'input', {bubbles:true}));
  })()`)
}
async function click(session, selector) {
  await evaluate(session, `document.querySelector(${JSON.stringify(selector)}).click()`)
}
async function clickText(session, text) {
  await evaluate(session, `(() => {
    const button = [...document.querySelectorAll('button,a')].find(e => e.textContent.trim() === ${JSON.stringify(text)});
    if (!button || button.disabled) throw new Error('Button not available: ' + ${JSON.stringify(text)});
    button.click();
  })()`)
}
async function login(session) {
  await command('Page.navigate', { url: baseURL + '/login' }, session)
  await waitFor(session, `!!document.querySelector('#email')`, 'login form')
  await fill(session, '#email', 'browser@example.test')
  await fill(session, '#password', 'browser-password')
  await click(session, 'button[type=submit]')
  await waitFor(session, `location.pathname === '/characters' || location.pathname === '/game'`, 'signed-in route')
}
async function screenshot(session, name) {
  if (!process.env.CHARACTER_SCREENSHOT_DIR) return
  const { data } = await command('Page.captureScreenshot', { format: 'png' }, session)
  await mkdir(process.env.CHARACTER_SCREENSHOT_DIR, { recursive: true })
  await writeFile(path.join(process.env.CHARACTER_SCREENSHOT_DIR, name + '.png'), Buffer.from(data, 'base64'))
}

try {
  let port, endpoint
  const deadline = Date.now() + 15000
  while (!endpoint && Date.now() < deadline) {
    if (launchError) throw launchError
    try { [port, endpoint] = (await readFile(path.join(profilePath, 'DevToolsActivePort'), 'utf8')).trim().split('\n') }
    catch { await delay(100) }
  }
  assert(endpoint, 'Browser did not start its debugger')
  socket = new WebSocket(`ws://127.0.0.1:${port}${endpoint}`)
  socket.addEventListener('message', ({ data }) => {
    const message = JSON.parse(data)
    if (message.method === 'Runtime.exceptionThrown') exceptions.push(message.params.exceptionDetails.text)
    const task = pending.get(message.id)
    if (!task) return
    pending.delete(message.id); clearTimeout(task.timer)
    if (message.error) task.reject(new Error(message.error.message))
    else task.resolve(message.result)
  })
  await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once:true }); socket.addEventListener('error', reject, { once:true }) })

  const first = await tab()
  await login(first)
  await waitFor(first, `document.body.innerText.includes('У вас пока нет персонажей')`, 'empty roster')
  await clickText(first, 'Создать персонажа')
  await waitFor(first, `document.querySelector('#faction')?.options.length === 3`, 'catalog')
  const factions = await evaluate(first, `[...document.querySelector('#faction').options].slice(1).map(o => o.value)`)
  await fill(first, '#character-name', 'Эйла')
  await fill(first, '#faction', factions[0])
  assert.equal(await evaluate(first, `document.querySelector('#character-class').options.length`), 5)
  await fill(first, '#character-class', await evaluate(first, `document.querySelector('#character-class').options[1].value`))
  await fill(first, '#faction', factions[1])
  assert.equal(await evaluate(first, `document.querySelector('#character-class').value`), '')
  assert.equal(await evaluate(first, `document.querySelector('#character-class').options.length`), 5)
  await fill(first, '#character-class', await evaluate(first, `document.querySelector('#character-class').options[1].value`))
  await screenshot(first, 'create-character')
  await clickText(first, 'Создать')
  await waitFor(first, `location.pathname === '/characters' && document.querySelector('h2')?.textContent === 'Эйла'`, 'created character in list')
  assert.equal(await evaluate(first, `fetch('/api/characters').then(r => r.json()).then(r => r.active_character_id)`), null)
  await clickText(first, 'Создать персонажа')
  await waitFor(first, `!!document.querySelector('#faction')`, 'second create form')
  await fill(first, '#character-name', 'эйла')
  await fill(first, '#faction', factions[1])
  await fill(first, '#character-class', await evaluate(first, `document.querySelector('#character-class').options[1].value`))
  await clickText(first, 'Создать')
  await waitFor(first, `document.querySelector('[role=alert]')?.textContent.includes('уже занято')`, 'duplicate name message')
  await clickText(first, 'К списку персонажей')
  await clickText(first, 'Играть')
  await waitFor(first, `location.pathname === '/game' && document.querySelector('h1')?.textContent === 'Эйла'`, 'selected character')
  await command('Page.reload', {}, first)
  await waitFor(first, `document.body.innerText.includes('Персонаж готов к игре.')`, 'selection after reload')
  await screenshot(first, 'selected-character')

  const { browserContextId } = await command('Target.createBrowserContext')
  const second = await tab(browserContextId)
  await login(second)
  assert.equal(await evaluate(second, 'location.pathname'), '/game')
  await clickText(first, 'К выбору персонажа')
  await waitFor(first, `location.pathname === '/characters'`, 'explicit clear')
  await evaluate(second, `window.dispatchEvent(new Event('focus'))`)
  await waitFor(second, `location.pathname === '/characters'`, 'clear shared with second session')
  await clickText(second, 'Играть')
  await waitFor(second, `location.pathname === '/game'`, 'second session play')
  await clickText(first, 'Выйти')
  await waitFor(first, `location.pathname === '/login'`, 'logout')
  await evaluate(second, `window.dispatchEvent(new Event('focus'))`)
  await waitFor(second, `location.pathname === '/characters' && document.body.innerText.includes('Эйла')`, 'other session survives logout with selection cleared')
  assert.deepEqual(exceptions, [], 'Uncaught browser exceptions')
  console.log('Browser flow PASS: login, catalog/filter, create/list, conflict, select/reload, cross-session clear/logout')
} catch (error) {
  if (browserLog) console.error(browserLog)
  throw error
} finally {
  if (socket?.readyState === WebSocket.OPEN) {
    await command('Browser.close').catch(() => {})
    socket.close()
  }
  await Promise.race([browserExited, delay(3000)])
  if (browser.exitCode === null) browser.kill()
  for (const task of pending.values()) clearTimeout(task.timer)
}
