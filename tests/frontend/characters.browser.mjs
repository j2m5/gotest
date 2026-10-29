// Optional real-browser smoke test, launched by tests/db/character_browser_test.go.
// Node's built-in WebSocket talks to a disposable headless Chromium profile.
import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile, readdir, writeFile, mkdir, rm } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { assertUIControls } from './ui-kit.browser-assertions.mjs'

const [baseURL, browserPath, profilePath, mailDirectory] = process.argv.slice(2)
assert(baseURL && browserPath && profilePath && mailDirectory, 'Expected test server, browser, temporary profile and local mailbox')
let browser, browserExited, launchError, socket
let browserLog = ''
let nextID = 0
const pending = new Map()
const exceptions = []
const networkFailures = []
const verificationRequests = []

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
async function login(session, rememberMe = false, password = 'browser-password', email = 'browser@example.test') {
  await command('Page.navigate', { url: baseURL + '/login' }, session)
  await waitFor(session, `!!document.querySelector('#email')`, 'login form')
  assert.equal(await evaluate(session, "document.querySelector('#remember-me')?.checked"), false, 'Remember me defaults to off')
  if (rememberMe) await click(session, '#remember-me')
  await fill(session, '#email', email)
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

async function checkLayout(session, name, widths = [320, 375, 480, 1024]) {
  await waitFor(session, `[...document.images].length > 0 && [...document.images].every(img => img.complete && img.naturalWidth > 0)`, name + ' images loaded')
  for (const width of widths) {
    await command('Emulation.setDeviceMetricsOverride', { width, height: 900, deviceScaleFactor: 1, mobile: false }, session)
    const layout = await evaluate(session, `new Promise(resolve => requestAnimationFrame(() => {
      const root = document.querySelector('#root').getBoundingClientRect();
      const footer = document.querySelector('.site-footer');
      const naturalTail = parseFloat(getComputedStyle(footer.lastElementChild).marginBottom) + parseFloat(getComputedStyle(document.querySelector('#root')).borderBottomWidth);
      resolve({viewport: document.documentElement.clientWidth, scroll: document.documentElement.scrollWidth, rootWidth: root.width, left: root.left,
        trailingSpace: root.bottom - footer.getBoundingClientRect().bottom, naturalTail});
    }))`)
    assert(layout.scroll <= layout.viewport, `${name} must not scroll horizontally at ${width}px: ${JSON.stringify(layout)}`)
    assert.equal(layout.rootWidth, Math.min(layout.viewport, 480), `${name} keeps the original maximum width`)
    assert(Math.abs(layout.left - (layout.viewport - layout.rootWidth) / 2) < 1, `${name} stays centered`)
    assert(layout.trailingSpace <= layout.naturalTail + .1, `${name} ends after its footer without stretching to the viewport: ${JSON.stringify(layout)}`)
    await screenshot(session, `${name}-${width}`)
  }
  await command('Emulation.setDeviceMetricsOverride', { width: 480, height: 900, deviceScaleFactor: 1, mobile: false }, session)
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
    if (message.method === 'Network.requestWillBeSent' && new URL(message.params.request.url).pathname === '/api/auth/verify') {
      verificationRequests.push({ session: message.sessionId, method: message.params.request.method })
    }
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

async function waitForMail(previousNames, subject, recipient = 'browser@example.test') {
  const deadline = Date.now() + 10000
  while (Date.now() < deadline) {
    for (const name of await readdir(mailDirectory)) {
      if (!name.endsWith('.txt') || previousNames.has(name)) continue
      const message = await readFile(path.join(mailDirectory, name), 'utf8')
      const [headers, ...body] = message.split(/\r?\n\r?\n/)
      if (!headers.split(/\r?\n/).includes(`Subject: ${subject}`)) continue
      if (!headers.split(/\r?\n/).includes(`To: ${recipient}`)) continue
      // This is the same readable file a developer opens in a text editor.
      // No MIME decoding or repairing of wrapped links is allowed here.
      if (body.length) return body.join('\r\n\r\n')
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

  await command('Page.navigate', { url: baseURL + '/missing-browser-route' }, first)
  await waitFor(first, `document.querySelector('h1')?.textContent === '404' && !!document.querySelector('main a[href="/"]')`, 'not found with home link')
  await click(first, 'main a[href="/"]')
  await waitFor(first, "location.pathname === '/'", 'home from not found')
  await checkLayout(first, 'guest')

  await command('Page.navigate', { url: baseURL + '/profile' }, first)
  await waitFor(first, `location.pathname === '/login' && !!document.querySelector('#email')`, 'profile requires a session')
  await command('Page.navigate', { url: baseURL + '/characters/1' }, first)
  await waitFor(first, `location.pathname === '/login' && !!document.querySelector('#email')`, 'public character profile requires a session')
  await checkLayout(first, 'login')
  const fieldBackground = await assertUIControls(evaluate, first, 'login')
  for (const width of [320, 480]) {
    await command('Emulation.setDeviceMetricsOverride', { width, height: 900, deviceScaleFactor: 1, mobile: false }, first)
    const checkboxLayout = () => evaluate(first, `new Promise(resolve => requestAnimationFrame(() => {
      const checkbox = document.querySelector('#remember-me');
      const row = checkbox.closest('.form-item');
      resolve({checkboxTop: checkbox.getBoundingClientRect().top, rowHeight: row.getBoundingClientRect().height,
        followingTop: row.nextElementSibling.getBoundingClientRect().top, footerTop: document.querySelector('.site-footer').getBoundingClientRect().top});
    }))`)
    const before = await checkboxLayout()
    await click(first, '#remember-me')
    assert.equal(await evaluate(first, `document.querySelector('#remember-me').checked`), true)
    assert.deepEqual(await checkboxLayout(), before, `Checking the checkbox must not shift the row or following content at ${width}px`)
    await click(first, '#remember-me')
    assert.equal(await evaluate(first, `document.querySelector('#remember-me').checked`), false)
    assert.deepEqual(await checkboxLayout(), before, `Unchecking the checkbox must preserve the layout at ${width}px`)
  }

  await command('Page.navigate', { url: baseURL + '/register' }, first)
  await waitFor(first, `!!document.querySelector('#login')`, 'registration form')
  await checkLayout(first, 'register', [320, 480])
  await assertUIControls(evaluate, first, 'registration', fieldBackground)
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

  const beforeRegistration = new Set(await readdir(mailDirectory))
  await click(first, 'button[type=submit]')
  await waitFor(first, `!!document.querySelector('[role=status]') && !document.querySelector('form')`, 'registered account awaits email confirmation')
  const verificationMessage = await waitForMail(beforeRegistration, 'Verify your email address', 'fields@example.test')
  const verificationLink = verificationMessage.match(/https?:\/\/[^\s]+/)?.[0]
  assert(verificationLink, 'Readable verification email must contain a copyable link')
  const verificationURL = new URL(verificationLink)
  assert.equal(verificationURL.origin, baseURL)
  assert.equal(verificationURL.pathname, '/verify')
  assert(/^[0-9a-f]{64}$/.test(verificationURL.searchParams.get('token') ?? ''), 'Copying the text-preview link must preserve the complete token')
  await command('Network.enable', {}, first)
  await command('Page.navigate', { url: verificationLink }, first)
  await waitFor(first, `location.search === '' && document.querySelector('[role=status]')?.textContent.includes('Почта подтверждена')`, 'email confirms automatically when its link opens')
  assert.deepEqual(verificationRequests, [{ session: first, method: 'POST' }], 'Opening the link sends exactly one confirmation request')
  assert.equal(await evaluate(first, 'document.querySelector("main button")'), null, 'Successful verification needs no confirmation button')
  await checkLayout(first, 'email-verified', [320, 480])
  await command('Page.reload', {}, first)
  await waitFor(first, `document.querySelector('[role=status]')?.textContent.includes('Почта подтверждена')`, 'refresh preserves the completed result without a token')
  assert.equal(verificationRequests.length, 1, 'Refresh does not consume the token again')
  await command('Page.navigate', { url: verificationLink }, first)
  await waitFor(first, `document.querySelector('[role=alert]')?.textContent.includes('уже использована')`, 'reopened single-use link explains that it may already be used')
  assert.equal(await evaluate(first, `!!document.querySelector('main a[href="/login"]') && !!document.querySelector('main a[href="/login?resend=1"]')`), true, 'Used or expired links offer login and a new email')
  await click(first, 'main a[href="/login?resend=1"]')
  await waitFor(first, '!!document.querySelector("#resend-verification")', 'invalid link opens resend help')
  await clickText(first, 'Письмо не пришло?')
  assert.equal(await evaluate(first, 'document.querySelector("#resend-verification")'), null, 'Automatically opened help can still be collapsed')
  await command('Page.navigate', { url: baseURL + '/verify' }, first)
  await waitFor(first, `document.querySelector('[role=alert]')?.textContent.includes('нет кода')`, 'missing token has a useful error')
  assert.equal(verificationRequests.length, 2, 'A missing token never sends a confirmation request')
  await login(first, false, 'browser-password', 'fields@example.test')
  await waitFor(first, `location.pathname === '/characters' && document.body.innerText.includes('valid-login')`, 'newly verified account can sign in')
  await clickText(first, 'Выйти')
  await waitFor(first, `location.pathname === '/login'`, 'verified registration fixture signed out')

  await command('Page.navigate', { url: baseURL + '/login' }, first)
  await waitFor(first, `!!document.querySelector('#email')`, 'login form')
  assert.equal(await evaluate(first, 'document.querySelector("#resend-verification")'), null, 'Ordinary login keeps resend controls collapsed')
  await clickText(first, 'Письмо не пришло?')
  await waitFor(first, '!!document.querySelector("#resend-verification")', 'resend help can be opened without a failed login')
  await clickText(first, 'Письмо не пришло?')
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

  await clickText(first, 'Письмо не пришло?')
  const beforeResend = new Set(await readdir(mailDirectory))
  await click(first, '#resend-verification')
  await waitFor(first, '!!document.querySelector("[role=status]")', 'resend email accepted')
  const resentMessage = await waitForMail(beforeResend, 'Verify your email address', 'unverified@example.test')
  const resentLink = resentMessage.match(/https?:\/\/[^\s]+/)?.[0]
  assert(resentLink, 'Resending produces a usable verification email')
  await command('Network.setExtraHTTPHeaders', { headers: { 'X-Browser-Test-Verification-Failure': '1' } }, first)
  await command('Page.navigate', { url: resentLink }, first)
  await waitFor(first, '!!document.querySelector("[role=alert]") && !!document.querySelector("main button")', 'temporary verification error offers retry')
  assert.equal(verificationRequests.length, 3, 'An unsuccessful automatic request is not retried in a loop')
  await command('Network.setExtraHTTPHeaders', { headers: {} }, first)
  await clickText(first, 'Повторить попытку')
  await waitFor(first, `location.search === '' && document.querySelector('[role=status]')?.textContent.includes('Почта подтверждена')`, 'retry completes verification after the server recovers')
  assert.equal(verificationRequests.length, 4, 'Explicit retry sends one new request')

  await login(first)
  assert.equal((await sessionCookie(first)).session, true, 'Ordinary login uses a browser-session cookie')
  assert.equal((await sessionCookie(first)).httpOnly, true)
  assert.equal(await evaluate(first, "document.cookie.includes('session_token')"), false)
  await waitFor(first, `document.body.innerText.includes('У вас пока нет персонажей')`, 'empty roster')
  await command('Page.navigate', { url: baseURL + '/profile' }, first)
  await waitFor(first, `location.pathname === '/characters' && document.body.innerText.includes('У вас пока нет персонажей')`, 'profile requires an active character')
  await clickText(first, 'Создать персонажа')
  await waitFor(first, `document.querySelector('#faction')?.options.length === 3`, 'catalog')
  const factions = await evaluate(first, `[...document.querySelector('#faction').options].slice(1).map(o => o.value)`)
  const creationCatalog = await evaluate(first, `fetch('/api/characters/catalog').then(r => r.json())`)
  let illustratedClasses = 0
  for (const faction of factions) {
    await fill(first, '#faction', faction)
    const factionImage = '/' + creationCatalog.factions.find(item => item.id === Number(faction)).icon
    await waitFor(first, `document.querySelector('.faction-preview img')?.getAttribute('src') === ${JSON.stringify(factionImage)} && document.querySelector('.faction-preview img')?.naturalWidth === 90`, 'faction illustration')
    const classes = await evaluate(first, `[...document.querySelector('#character-class').options].slice(1).map(o => o.value)`)
    assert.equal(classes.length, 4, 'Each faction offers four base classes')
    for (const characterClass of classes) {
      await fill(first, '#character-class', characterClass)
      const classImage = '/' + creationCatalog.classes.find(item => item.id === Number(characterClass)).icon
      await waitFor(first, `document.querySelector('.class-preview img')?.getAttribute('src') === ${JSON.stringify(classImage)} && document.querySelector('.class-preview img')?.naturalWidth === 195 && document.querySelector('.class-preview img')?.complete`, 'class illustration')
      illustratedClasses++
    }
  }
  assert.equal(illustratedClasses, 8, 'Both factions and all base classes have illustrations')
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
  await checkLayout(first, 'create-character')
  await assertUIControls(evaluate, first, 'character creation', fieldBackground)
  const unfocusedBorder = await evaluate(first, `getComputedStyle(document.querySelector('#character-class')).borderColor`)
  await evaluate(first, `document.querySelector('#faction').focus()`)
  await command('Input.dispatchKeyEvent', { type: 'keyDown', key: 'Tab', code: 'Tab', windowsVirtualKeyCode: 9 }, first)
  await command('Input.dispatchKeyEvent', { type: 'keyUp', key: 'Tab', code: 'Tab', windowsVirtualKeyCode: 9 }, first)
  assert.equal(await evaluate(first, `document.activeElement.id`), 'character-class', 'Native class selection remains keyboard accessible')
  await waitFor(first, `getComputedStyle(document.activeElement).borderColor !== ${JSON.stringify(unfocusedBorder)} && getComputedStyle(document.activeElement).boxShadow !== 'none'`, 'Keyboard focus is visible through a soft border and highlight')
  assert.equal(await evaluate(first, `getComputedStyle(document.activeElement).outlineStyle`), 'none', 'Field focus does not add the old harsh outer outline')
  await clickText(first, 'Создать')
  await waitFor(first, `location.pathname === '/characters' && document.querySelector('h2')?.textContent === 'Эйла'`, 'created character in list')
  await checkLayout(first, 'roster', [320, 480])
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
  await waitFor(first, `location.pathname === '/game' && document.querySelector('.character-strip a')?.textContent === '[1] Эйла'`, 'selected character')
  await command('Page.reload', {}, first)
  await waitFor(first, `document.querySelector('.character-strip a')?.textContent === '[1] Эйла' && !!document.querySelector('[role="meter"]')`, 'selection after reload')
  await screenshot(first, 'selected-character')
  await waitFor(first, `!!document.querySelector('main .scene-image')`, 'faction landscape')
  await checkLayout(first, 'game')

  await evaluate(first, `document.querySelector('.character-strip a').click()`)
  await waitFor(first, `location.pathname === '/profile' && document.querySelector('h1')?.textContent === 'Эйла'`, 'own character profile')
  const selected = await evaluate(first, `fetch('/api/characters').then(r => r.json()).then(r => r.characters.find(c => c.id === r.active_character_id))`)
  const catalog = await evaluate(first, `fetch('/api/characters/catalog').then(r => r.json())`)
  const selectedClass = catalog.classes.find(item => item.id === selected.class_id)
  const selectedFaction = catalog.factions.find(item => item.id === selected.faction_id)
  await waitFor(first, `document.querySelector('.character-identity')?.textContent.includes(${JSON.stringify(selectedClass.name)})`, 'profile catalog identity')
  const profileText = await evaluate(first, `document.querySelector('main').innerText`)
  assert(profileText.includes(selectedFaction.name), 'Profile must resolve the character faction from the catalog')
  assert(profileText.includes(`Опыт: ${selected.experience}`), 'Profile must show actual accumulated experience')
  for (const [label, value, maximum] of [['Здоровье', selected.hp_current, selected.hp], ['Энергия', selected.mp_current, selected.mp]]) {
    const meter = await evaluate(first, `(() => { const el = document.querySelector('[role="meter"][aria-label="${label}"]'); return {value: Number(el?.getAttribute('aria-valuenow')), max: Number(el?.getAttribute('aria-valuemax'))} })()`)
    assert.deepEqual(meter, { value, max: maximum }, 'Resource panels must use current character data')
  }
  await command('Page.reload', {}, first)
  await waitFor(first, `location.pathname === '/profile' && document.querySelector('h1')?.textContent === 'Эйла'`, 'profile after reload')
  assert.equal(await evaluate(first, `fetch('/api/characters').then(r => r.json()).then(r => r.active_character_id)`), selected.id, 'Opening the profile must preserve selection')
  await screenshot(first, 'character-profile')
  await checkLayout(first, 'profile')

  // Presence text belongs only to public profiles; the directory lists selected characters.
  assert.equal(await evaluate(first, `document.querySelector('.presence-indicator')`), null, 'Private profile and footer do not show presence indicators')
  await evaluate(first, `document.querySelector('.presence-footer a').click()`)
  await waitFor(first, `location.pathname === '/online' && document.querySelector('h1')?.textContent === 'Кто онлайн' && document.querySelector('.online-directory')?.textContent.includes('Эйла')`, 'online directory contains selected character')
  assert.equal(await evaluate(first, `document.querySelectorAll('.online-directory li').length`), 1, 'Unselected characters must not appear online')
  assert.equal(await evaluate(first, `document.querySelector('.presence-indicator')`), null, 'Directory nicknames do not show presence indicators')
  await screenshot(first, 'online-directory')
  await checkLayout(first, 'online-directory', [320, 480])
  await evaluate(first, `document.querySelector('.online-directory a').click()`)
  await waitFor(first, `location.pathname === '/characters/${selected.id}' && document.querySelector('.character-identity .presence-indicator--online')?.textContent === 'В игре'`, 'online directory opens public profile with live status text')
  assert.equal(await evaluate(first, `!!document.querySelector('.game-navigation a[href="/online"]')`), true, 'Public profile preserves return to online list')
  await screenshot(first, 'public-profile-online')
  await checkLayout(first, 'public-profile-online', [320, 480])
  await command('Page.navigate', { url: baseURL + '/characters/1' }, first)
  await waitFor(first, `document.querySelector('h1')?.textContent === 'First character' && document.querySelector('.character-identity .presence-indicator--offline')?.textContent === 'Не в игре'`, 'unselected other character is offline')
  await screenshot(first, 'public-profile-offline')
  await command('Page.navigate', { url: baseURL + '/profile' }, first)
  await waitFor(first, `location.pathname === '/profile' && document.querySelector('h1')?.textContent === 'Эйла'`, 'own profile after presence directory')

  // A transient catalog failure must not hide resources or strand the player.
  await command('Network.setExtraHTTPHeaders', { headers: { 'X-Browser-Test-Catalog-Failure': '1' } }, first)
  await command('Page.reload', {}, first)
  await waitFor(first, `!!document.querySelector('[role="alert"]') && document.querySelector('h1')?.textContent === 'Эйла'`, 'profile survives catalog failure')
  assert.equal(await evaluate(first, `document.querySelectorAll('[role="meter"]').length`), 2)
  assert.equal(await evaluate(first, `[...document.querySelectorAll('button')].some(button => button.textContent === 'К выбору персонажа' && !button.disabled)`), true)
  await command('Network.setExtraHTTPHeaders', { headers: {} }, first)
  await clickText(first, 'Повторить')
  await waitFor(first, `document.querySelector('.character-identity')?.textContent.includes(${JSON.stringify(selectedClass.name)}) && !document.querySelector('[role="alert"]')`, 'catalog retry restores profile identity')
  await evaluate(first, `document.querySelector('.site-header a').click()`)
  await waitFor(first, `location.pathname === '/game'`, 'game from profile')

  // Tabs in the default context share cookies and auth notifications.
  const peer = await tab()
  await command('Page.navigate', { url: baseURL + '/characters' }, peer)
  await waitFor(peer, "document.body.innerText.includes('browser') && !!document.querySelector('h2')", 'same-cookie peer roster')
  // A background refresh can discover a different account without a broadcast.
  // Local page state must be discarded before that account can submit a form.
  await command('Page.navigate', { url: baseURL + '/characters/new' }, first)
  await waitFor(first, "!!document.querySelector('#character-name')", 'original account draft')
  await fill(first, '#character-name', 'Old Draft')
  await fill(first, '#faction', factions[1])
  await fill(first, '#character-class', await evaluate(first, "document.querySelector('#character-class').options[1].value"))
  assert.equal(await evaluate(peer, `fetch('/api/auth/login', {
    method: 'POST', headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({email:'other-browser@example.test', password:'browser-password'})
  }).then(r => r.status)`), 200)
  await command('Page.bringToFront', {}, first)
  await evaluate(first, "window.dispatchEvent(new Event('focus'))")
  await waitFor(first, "document.querySelector('#character-name')?.value === '' && !!document.querySelector('#faction')", 'new account discards previous draft')
  assert.equal(await evaluate(first, "document.querySelector('#character-name').value"), '', 'New account must not inherit the previous account draft')
  assert.equal(await evaluate(first, "document.querySelector('#faction').value"), '')
  assert.equal(await evaluate(first, "document.querySelector('#character-class').value"), '')
  assert.equal(await evaluate(first, "fetch('/api/characters').then(r => r.json()).then(r => r.characters.length)"), 0)
  await login(peer)
  await command('Page.navigate', { url: baseURL + '/game' }, first)
  await waitFor(first, "location.pathname === '/game' && document.querySelector('.character-strip a')?.textContent === '[1] Эйла'", 'original account restored after draft check')
  await login(peer, false, 'browser-password', 'other-browser@example.test')
  await waitFor(first, "document.body.innerText.includes('other-browser') && document.querySelector('h2') === null", 'cross-tab login invalidates old account')
  // Change the shared cookie without emitting an auth event. The stale UI must
  // send its displayed B ID; logout must reject it and preserve the A cookie.
  assert.equal(await evaluate(peer, `fetch('/api/auth/login', {
    method: 'POST', headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({email:'browser@example.test', password:'browser-password'})
  }).then(r => r.status)`), 200)
  await clickText(first, 'Выйти')
  await waitFor(first, "document.body.innerText.includes('browser') && !document.body.innerText.includes('other-browser') && !!document.querySelector('h2')", 'stale logout rejected and account refreshed')
  assert.equal(await evaluate(first, "fetch('/api/auth/me').then(r => r.json()).then(r => r.user.email)"), 'browser@example.test')
  await command('Page.close', {}, peer)
  await command('Page.navigate', { url: baseURL + '/game' }, first)
  await waitFor(first, "location.pathname === '/game' && document.querySelector('.character-strip a')?.textContent === '[1] Эйла'", 'original selection preserved after rejected logout')

  const { browserContextId } = await command('Target.createBrowserContext')
  const second = await tab(browserContextId)
  await login(second)
  assert.equal(await evaluate(second, 'location.pathname'), '/game')
  await evaluate(first, `document.querySelector('.character-strip a').click()`)
  await waitFor(first, `location.pathname === '/profile'`, 'profile before clearing selection')
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

  const longName = 'А'.repeat(32)
  await clickText(ordinary, 'Создать персонажа')
  await waitFor(ordinary, `!!document.querySelector('#faction')`, 'long-name character form')
  await fill(ordinary, '#character-name', longName)
  await fill(ordinary, '#faction', factions[0])
  await fill(ordinary, '#character-class', await evaluate(ordinary, `document.querySelector('#character-class').options[1].value`))
  await clickText(ordinary, 'Создать')
  await waitFor(ordinary, `location.pathname === '/characters' && [...document.querySelectorAll('h2')].some(el => el.textContent === ${JSON.stringify(longName)})`, 'long-name roster entry')
  await checkLayout(ordinary, 'long-name-roster', [320])
  await evaluate(ordinary, `[...document.querySelectorAll('.character-card')].find(card => card.querySelector('h2').textContent === ${JSON.stringify(longName)}).querySelector('button').click()`)
  await waitFor(ordinary, `location.pathname === '/game' && document.querySelector('.character-strip a')?.textContent === ${JSON.stringify('[1] ' + longName)}`, 'newly selected character replaces old identity')
  await checkLayout(ordinary, 'long-name-game', [320])
  await evaluate(ordinary, `document.querySelector('.character-strip a').click()`)
  await waitFor(ordinary, `location.pathname === '/profile' && document.querySelector('h1')?.textContent === ${JSON.stringify(longName)}`, 'long-name profile')
  await checkLayout(ordinary, 'long-name-profile', [320, 480])
  await clickText(ordinary, 'Выйти')
  await waitFor(ordinary, `location.pathname === '/login' && !!document.querySelector('#email')`, 'logout from profile')
  assert.equal(await sessionCookie(ordinary), undefined)

  await login(ordinary, false, 'browser-password', 'profile-fixture@example.test')
  await clickText(ordinary, 'Играть')
  await waitFor(ordinary, `location.pathname === '/game' && document.querySelector('.character-strip a')?.textContent === '[1] Проверка'`, 'resource boundary fixture')
  await evaluate(ordinary, `document.querySelector('.character-strip a').click()`)
  await waitFor(ordinary, `location.pathname === '/profile' && !!document.querySelector('.character-portrait')`, 'resource boundary profile')
  await waitFor(ordinary, `document.querySelector('.character-portrait')?.getAttribute('src') === '/images/icons/neutral.png' && document.querySelector('.character-portrait')?.naturalWidth > 0`, 'missing stored avatar falls back')
  const initialHP = await evaluate(ordinary, `Number(document.querySelector('[role="meter"][aria-label="Здоровье"]').getAttribute('aria-valuenow'))`)
  assert(initialHP >= 37 && initialHP < 200, 'Health includes recovery since the stored fixture baseline')
  await waitFor(ordinary, `Number(document.querySelector('[role="meter"][aria-label="Здоровье"]').getAttribute('aria-valuenow')) > ${initialHP}`, 'health regenerates without reloading the page')
  assert.equal(await evaluate(ordinary, `(() => { const current = document.querySelector('[role="meter"][aria-label="Здоровье"]').getAttribute('aria-valuenow'); return document.querySelector('.character-stats').textContent.includes(current + ' / 200'); })()`), true, 'Profile numbers and live bar agree')
  assert.equal(await evaluate(ordinary, `document.querySelector('[role="meter"][aria-label="Энергия"]').getAttribute('aria-valuetext')`), '0 из 0', 'A character without energy must not gain a fictitious maximum')
  assert.equal(await evaluate(ordinary, `document.querySelector('.resource-energy .resource-caption').textContent`), 'Энергия: 0 / 00%')
  assert.equal(await evaluate(ordinary, `document.querySelector('.resource-energy .resource-fill').style.width`), '0%')
  await checkLayout(ordinary, 'profile-zero-energy', [320])
  assert.deepEqual(exceptions, [], 'Uncaught browser exceptions')
  console.log('Browser flow PASS: registration and automatic email confirmation, one POST per attempt, refresh/reuse/missing-token handling, secondary resend help and recovery after 503; legacy shell at 320/375/480/1024, all 8 class images, keyboard focus, profile guards/data/reload/catalog retry, long-name selection and portrait fallback; 404/home, two-tab auth sync and stale logout guard, field validation/editing, distinct email/origin 403, login, catalog/filter, create/list, conflict, select/reload, cross-session clear/logout, remember-me cookie/restart/replacement, queued local-mail password recovery, validation/replay, cookie/state/session revocation and new-password login')
} catch (error) {
  if (browserLog) console.error(browserLog)
  if (networkFailures.length) console.error('Network failures:', networkFailures)
  throw error
} finally {
  await stopBrowser()
}
