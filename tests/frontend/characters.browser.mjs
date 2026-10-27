// Optional real-browser smoke test, launched by tests/db/character_browser_test.go.
// Node's built-in WebSocket talks to a disposable headless Chromium profile.
import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile, readdir, writeFile, mkdir, rm } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

const [baseURL, browserPath, profilePath, mailDirectory] = process.argv.slice(2)
assert(baseURL && browserPath && profilePath && mailDirectory, 'Expected test server, browser, temporary profile and local mailbox')
let browser, browserExited, launchError, socket
let browserLog = ''
let nextID = 0
const pending = new Map()
const exceptions = []
const networkFailures = []

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
async function login(session, rememberMe = false, password = 'browser-password') {
  await command('Page.navigate', { url: baseURL + '/login' }, session)
  await waitFor(session, `!!document.querySelector('#email')`, 'login form')
  assert.equal(await evaluate(session, "document.querySelector('#remember-me')?.checked"), false, 'Remember me defaults to off')
  if (rememberMe) await click(session, '#remember-me')
  await fill(session, '#email', 'browser@example.test')
  await fill(session, '#password', password)
  await click(session, 'button[type=submit]')
  await waitFor(session, `location.pathname === '/characters' || location.pathname === '/game'`, 'signed-in route')
}
async function screenshot(session, name) {
  if (!process.env.CHARACTER_SCREENSHOT_DIR) return
  const { data } = await command('Page.captureScreenshot', { format: 'png' }, session)
  await mkdir(process.env.CHARACTER_SCREENSHOT_DIR, { recursive: true })
  await writeFile(path.join(process.env.CHARACTER_SCREENSHOT_DIR, name + '.png'), Buffer.from(data, 'base64'))
}

async function startBrowser() {
  // Chrome writes a fresh endpoint for each launch of this disposable profile.
  await rm(path.join(profilePath, 'DevToolsActivePort'), { force: true })
  launchError = undefined
  browserLog = ''
  browser = spawn(browserPath, [
    '--headless=new', '--remote-debugging-port=0', `--user-data-dir=${profilePath}`,
    '--disable-background-networking', '--no-first-run', '--no-default-browser-check',
    '--disable-extensions', '--disable-gpu', 'about:blank',
  ], { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] })
  browser.stderr.on('data', chunk => { browserLog = (browserLog + chunk).slice(-6000) })
  browserExited = new Promise(resolve => browser.once('exit', resolve))
  browser.on('error', error => { launchError = error })
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
    if (message.method === 'Network.loadingFailed') networkFailures.push({ error: message.params.errorText, cors: message.params.corsErrorStatus })
    const task = pending.get(message.id)
    if (!task) return
    pending.delete(message.id); clearTimeout(task.timer)
    if (message.error) task.reject(new Error(message.error.message))
    else task.resolve(message.result)
  })
  await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once:true }); socket.addEventListener('error', reject, { once:true }) })

}

async function stopBrowser() {
  if (socket?.readyState === WebSocket.OPEN) {
    await command('Browser.close').catch(() => {})
    socket.close()
  }
  await Promise.race([browserExited, delay(3000)])
  if (browser?.exitCode === null) {
    browser.kill()
    await Promise.race([browserExited, delay(3000)])
  }
  for (const task of pending.values()) clearTimeout(task.timer)
  pending.clear()
}

async function sessionCookie(session) {
  const { cookies } = await command('Network.getCookies', { urls: [baseURL] }, session)
  const matching = cookies.filter(cookie => cookie.name === 'session_token')
  assert(matching.length <= 1, 'Cookie modes must share the same name and scope')
  return matching[0]
}

async function waitForMail(previousNames, subject) {
  const deadline = Date.now() + 10000
  while (Date.now() < deadline) {
    for (const name of await readdir(mailDirectory)) {
      if (!name.endsWith('.eml') || previousNames.has(name)) continue
      const message = await readFile(path.join(mailDirectory, name), 'utf8')
      const [headers, ...body] = message.split(/\r?\n\r?\n/)
      if (!headers.split(/\r?\n/).includes(`Subject: ${subject}`)) continue
      assert(headers.split(/\r?\n/).includes('To: browser@example.test'), 'Mail recipient must match the requesting account')
      return body.join('\r\n\r\n').replace(/=\r?\n/g, '').replace(/=([0-9A-F]{2})/gi, (_, hex) => String.fromCharCode(parseInt(hex, 16)))
    }
    await delay(50)
  }
  throw new Error(`Waiting for local email: ${subject}`)
}

async function submitReset(session, password, confirmation = password) {
  await fill(session, '#password', password)
  await fill(session, '#confirmation', confirmation)
  await click(session, 'button[type=submit]')
}

try {
  await startBrowser()
  const first = await tab()

  await command('Page.navigate', { url: baseURL + '/register' }, first)
  await waitFor(first, `!!document.querySelector('#login')`, 'registration form')
  await fill(first, '#email', 'fields@example.test')
  await fill(first, '#login', 'x'.repeat(33))
  await fill(first, '#password', 'short')
  await click(first, 'button[type=submit]')
  await waitFor(first, `!!document.querySelector('#login-error') && !!document.querySelector('#password-error')`, 'registration field errors')
  assert.equal(await evaluate(first, `document.querySelector('#login').getAttribute('aria-invalid')`), 'true')
  assert.equal(await evaluate(first, `document.querySelector('#password').getAttribute('aria-describedby')`), 'password-error')
  assert.equal(await evaluate(first, `document.querySelector('#email').getAttribute('aria-invalid')`), 'false')
  await screenshot(first, 'registration-field-errors')
  await fill(first, '#login', 'valid-login')
  assert.equal(await evaluate(first, `document.querySelector('#login-error') === null && !!document.querySelector('#password-error')`), true)
  await fill(first, '#password', 'browser-password')
  assert.equal(await evaluate(first, `document.querySelector('#password-error') === null`), true)

  await command('Page.navigate', { url: baseURL + '/login' }, first)
  await waitFor(first, `!!document.querySelector('#email')`, 'login form')
  await fill(first, '#email', 'unverified@example.test')
  await fill(first, '#password', 'browser-password')
  await click(first, 'button[type=submit]')
  await waitFor(first, `document.body.innerText.includes('Подтвердите почту, чтобы войти.')`, 'unverified email guidance')

  // Ask the test server to inject cross-site metadata before the real origin
  // guard. A second 403 must replace the previous verification guidance.
  await command('Network.enable', {}, first)
  await command('Network.setExtraHTTPHeaders', { headers: { 'X-Browser-Test-Cross-Site': '1' } }, first)
  await click(first, 'button[type=submit]')
  await waitFor(first, `document.querySelector('[role=alert]')?.textContent.includes('Запрос с другого сайта запрещён')`, 'origin rejection')
  assert.equal(await evaluate(first, `document.body.innerText.includes('Подтвердите почту, чтобы войти.')`), false)
  await screenshot(first, 'origin-forbidden')
  await command('Network.setExtraHTTPHeaders', { headers: {} }, first)

  await login(first)
  assert.equal((await sessionCookie(first)).session, true, 'Ordinary login uses a browser-session cookie')
  assert.equal((await sessionCookie(first)).httpOnly, true)
  assert.equal(await evaluate(first, "document.cookie.includes('session_token')"), false)
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
  await fill(first, '#character-name', 'Э')
  await clickText(first, 'Создать')
  await waitFor(first, `!!document.querySelector('#name-error')`, 'character name validation')
  assert.equal(await evaluate(first, `document.querySelector('#character-name').getAttribute('aria-describedby')`), 'name-help name-error')
  await fill(first, '#character-name', 'Эйла')
  assert.equal(await evaluate(first, `document.querySelector('#name-error') === null`), true)
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
  await waitFor(first, `document.querySelector('#name-error')?.textContent.includes('уже занято')`, 'duplicate name message')
  assert.equal(await evaluate(first, `document.querySelector('#character-name').getAttribute('aria-invalid')`), 'true')
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
  // Use the default profile: incognito contexts cannot exercise disk persistence.
  await login(first, true)
  const remembered = await sessionCookie(first)
  assert.equal(remembered.session, false)
  assert.equal(remembered.httpOnly, true)
  assert.equal(remembered.sameSite, 'Lax')
  assert(Math.abs(remembered.expires - Date.now() / 1000 - 30 * 24 * 3600) < 10)
  await stopBrowser()
  await startBrowser()
  const restored = await tab()
  await command('Page.navigate', { url: baseURL + '/characters' }, restored)
  await waitFor(restored, `location.pathname === '/characters' && document.body.innerText.includes('Эйла')`, 'remembered login after browser restart')
  assert.equal((await sessionCookie(restored)).value, remembered.value)
  await clickText(restored, 'Выйти')
  await waitFor(restored, `location.pathname === '/login'`, 'remembered logout')
  assert.equal(await sessionCookie(restored), undefined)

  await login(restored, true)
  // Re-authentication without the flag replaces the persistent cookie in place.
  assert.equal(await evaluate(restored, `fetch('/api/auth/login', {
    method: 'POST', headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({email:'browser@example.test', password:'browser-password', remember_me:false})
  }).then(response => response.status)`), 200)
  assert.equal((await sessionCookie(restored)).session, true)
  const preResetCookie = await sessionCookie(restored)
  await stopBrowser()
  await startBrowser()
  const ordinary = await tab()
  await command('Page.navigate', { url: baseURL + '/characters' }, ordinary)
  await waitFor(ordinary, `location.pathname === '/login' && !!document.querySelector('#email')`, 'ordinary cookie absent after clean restart')
  assert.equal(await sessionCookie(ordinary), undefined)

  // The ordinary browser cookie is gone, but its server session still exists.
  const savedSessionStatus = () => fetch(baseURL + '/api/auth/me', { headers: { Cookie: `${preResetCookie.name}=${preResetCookie.value}` } }).then(response => response.status)
  assert.equal(await savedSessionStatus(), 200, 'Saved server session must exist before reset')
  await command('Page.navigate', { url: baseURL + '/reset-password' }, ordinary)
  await waitFor(ordinary, `!!document.querySelector('[role=alert]') && !!document.querySelector('a[href="/forgot-password"]')`, 'missing token recovery')
  assert.equal(await evaluate(ordinary, `document.querySelector('form') === null`), true)
  await click(ordinary, 'a[href="/login"]')
  await waitFor(ordinary, `!!document.querySelector('#email')`, 'login before recovery request')
  await click(ordinary, 'a[href="/forgot-password"]')
  await waitFor(ordinary, `location.pathname === '/forgot-password' && !!document.querySelector('#email')`, 'forgot password form from login')
  const beforeRequest = new Set(await readdir(mailDirectory))
  await fill(ordinary, '#email', 'browser@example.test')
  await click(ordinary, 'button[type=submit]')
  await waitFor(ordinary, `!!document.querySelector('[role=status]') && document.querySelector('form') === null`, 'generic reset request confirmation')
  const resetMessage = await waitForMail(beforeRequest, 'Reset your password')
  const resetLink = resetMessage.match(/https?:\/\/[^\s]+/)?.[0]
  assert(resetLink, 'Reset email must contain a link')
  const resetURL = new URL(resetLink)
  assert.equal(resetURL.origin, baseURL)
  assert.equal(resetURL.pathname, '/reset-password')
  assert(/^[0-9a-f]{64}$/.test(resetURL.searchParams.get('token') ?? ''), 'Reset email must carry the raw one-use token')

  // A live browser session must also be cleared by the successful reset.
  await login(ordinary)
  assert(await sessionCookie(ordinary), 'Recovery must start with an authenticated browser')
  await command('Page.navigate', { url: resetLink }, ordinary)
  await waitFor(ordinary, `!!document.querySelector('#confirmation')`, 'reset form from real email link')
  const newPassword = 'changed-browser-password'
  await submitReset(ordinary, newPassword, 'different-password')
  await waitFor(ordinary, `!!document.querySelector('#confirmation-error')`, 'password confirmation mismatch')
  assert.equal(await evaluate(ordinary, `document.querySelector('#confirmation').getAttribute('aria-invalid')`), 'true')
  // This is 74 UTF-8 bytes and only 37 characters, so the real API must reject it.
  await submitReset(ordinary, '\u044f'.repeat(37))
  await waitFor(ordinary, `!!document.querySelector('#password-error')`, 'server password byte-length validation')
  assert.equal(await evaluate(ordinary, `document.querySelector('#password').getAttribute('aria-invalid')`), 'true')
  assert.equal(await evaluate(ordinary, `document.querySelector('#confirmation-error') === null`), true)
  const beforeConsumption = new Set(await readdir(mailDirectory))
  await submitReset(ordinary, newPassword)
  await waitFor(ordinary, `location.search === '' && !!document.querySelector('[role=status]') && document.querySelector('form') === null`, 'successful reset and token removed from URL')
  assert.equal(await sessionCookie(ordinary), undefined, 'Reset must clear the current browser cookie')
  assert.equal(await evaluate(ordinary, `document.querySelector('a[href="/characters"]') === null && document.querySelector('a[href="/game"]') === null`), true, 'Reset must clear authenticated navigation')
  assert.equal(await savedSessionStatus(), 401, 'Reset must revoke the saved session from before the browser restart')
  const changedMessage = await waitForMail(beforeConsumption, 'Your password was changed')
  assert(!changedMessage.includes(newPassword) && !changedMessage.includes(resetURL.searchParams.get('token')), 'Changed notification must exclude credentials and reset tokens')
  await screenshot(ordinary, 'password-reset-complete')

  await command('Page.navigate', { url: resetLink }, ordinary)
  await waitFor(ordinary, `!!document.querySelector('#confirmation')`, 'replayed reset link form')
  await submitReset(ordinary, newPassword)
  await waitFor(ordinary, `!!document.querySelector('[role=alert]') && !!document.querySelector('a[href="/forgot-password"]')`, 'reused reset link recovery guidance')
  await command('Page.navigate', { url: baseURL + '/login' }, ordinary)
  await waitFor(ordinary, `!!document.querySelector('#email')`, 'login after reset')
  await fill(ordinary, '#email', 'browser@example.test')
  await fill(ordinary, '#password', 'browser-password')
  await click(ordinary, 'button[type=submit]')
  await waitFor(ordinary, `!!document.querySelector('[role=alert]') && !document.querySelector('button[type=submit]').disabled`, 'old password rejection')
  assert.equal(await sessionCookie(ordinary), undefined)
  await login(ordinary, false, newPassword)
  assert.equal(await evaluate(ordinary, 'location.pathname'), '/characters', 'Reset clears character selection')
  assert.equal(await evaluate(ordinary, `fetch('/api/characters').then(r => r.json()).then(r => r.active_character_id)`), null)
  assert.deepEqual(exceptions, [], 'Uncaught browser exceptions')
  console.log('Browser flow PASS: field validation/editing, distinct email/origin 403, login, catalog/filter, create/list, conflict, select/reload, cross-session clear/logout, remember-me cookie/restart/replacement, queued local-mail password recovery, validation/replay, cookie/state/session revocation and new-password login')
} catch (error) {
  if (browserLog) console.error(browserLog)
  if (networkFailures.length) console.error('Network failures:', networkFailures)
  throw error
} finally {
  await stopBrowser()
}
