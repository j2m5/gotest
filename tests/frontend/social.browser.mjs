import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile, writeFile, mkdir } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

const [baseURL, browserPath, profilePath] = process.argv.slice(2)
const browser = spawn(browserPath, ['--headless=new', '--remote-debugging-port=0', `--user-data-dir=${profilePath}`, '--disable-background-networking', '--no-first-run', '--no-default-browser-check', '--disable-extensions', '--disable-gpu', 'about:blank'], { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] })
let log = '', socket, next = 0
const waiting = new Map(), exceptions = []
browser.stderr.on('data', chunk => { log = (log + chunk).slice(-4000) })
const exited = new Promise(resolve => browser.once('exit', resolve))
function command(method, params = {}, sessionId) {
  return new Promise((resolve, reject) => {
    const id = ++next, timer = setTimeout(() => { waiting.delete(id); reject(new Error(`CDP timeout ${method}`)) }, 10000)
    waiting.set(id, { resolve, reject, timer }); socket.send(JSON.stringify({ id, method, params, sessionId }))
  })
}
async function evaluate(session, expression) {
  const result = await command('Runtime.evaluate', { expression, awaitPromise: true, returnByValue: true }, session)
  if (result.exceptionDetails) throw new Error(result.exceptionDetails.exception?.description ?? result.exceptionDetails.text)
  return result.result.value
}
async function wait(session, expression, label) {
  const end = Date.now() + 10000
  while (Date.now() < end) { if (await evaluate(session, expression)) return; await delay(60) }
  throw new Error(`${label}: ${await evaluate(session, 'document.body.innerText')}`)
}
async function open(session, route) { await command('Page.navigate', { url: baseURL + route }, session) }
async function fill(session, selector, value) {
  await evaluate(session, `(() => { const el=document.querySelector(${JSON.stringify(selector)}); if(!el)throw new Error('Missing field'); const proto=el.tagName==='SELECT'?HTMLSelectElement.prototype:el.tagName==='TEXTAREA'?HTMLTextAreaElement.prototype:HTMLInputElement.prototype; Object.getOwnPropertyDescriptor(proto,'value').set.call(el,${JSON.stringify(value)}); el.dispatchEvent(new Event(el.tagName==='SELECT'?'change':'input',{bubbles:true})); })()`)
}
async function click(session, text) {
  await evaluate(session, `(() => { const el=[...document.querySelectorAll('button,a')].find(e=>e.textContent.trim()===${JSON.stringify(text)}); if(!el||el.disabled)throw new Error('Button unavailable: '+${JSON.stringify(text)});el.click(); })()`)
}
async function request(session, url, method = 'GET', body, accountId) {
  return evaluate(session, `(async()=>{const r=await fetch(${JSON.stringify(url)},{method:${JSON.stringify(method)},headers:{'Content-Type':'application/json',${accountId ? `'X-Account-ID':'${accountId}'` : ''}},${body ? `body:JSON.stringify(${JSON.stringify(body)}),` : ''}});return {status:r.status,body:await r.json()};})()`)
}
async function tab(email) {
  const { browserContextId } = await command('Target.createBrowserContext')
  const { targetId } = await command('Target.createTarget', { url: 'about:blank', browserContextId })
  const { sessionId } = await command('Target.attachToTarget', { targetId, flatten: true })
  await command('Page.enable', {}, sessionId); await command('Runtime.enable', {}, sessionId)
  await command('Emulation.setDeviceMetricsOverride', { width: 480, height: 900, deviceScaleFactor: 1, mobile: false }, sessionId)
  await open(sessionId, '/login'); await wait(sessionId, "!!document.querySelector('#email')", 'login')
  await fill(sessionId, '#email', email); await fill(sessionId, '#password', 'browser-password')
  await evaluate(sessionId, "document.querySelector('button[type=submit]').click()")
  await wait(sessionId, "location.pathname==='/characters'||location.pathname==='/game'", 'login complete')
  await open(sessionId, '/chat'); await wait(sessionId, "!!document.querySelector('#chat-body')&&!document.body.innerText.includes('Загрузка сообщений')", 'chat ready')
  return sessionId
}
async function screenshot(session, name) {
  if (!process.env.SOCIAL_SCREENSHOT_DIR) return
  await mkdir(process.env.SOCIAL_SCREENSHOT_DIR, { recursive: true })
  const { data } = await command('Page.captureScreenshot', { format: 'png' }, session)
  await writeFile(path.join(process.env.SOCIAL_SCREENSHOT_DIR, name + '.png'), Buffer.from(data, 'base64'))
}
try {
  let endpoint, port
  const deadline = Date.now() + 15000
  while (!endpoint && Date.now() < deadline) {
    try { [port, endpoint] = (await readFile(path.join(profilePath, 'DevToolsActivePort'), 'utf8')).trim().split('\n') } catch { await delay(100) }
  }
  assert(endpoint, 'Chrome debugger unavailable: ' + log)
  socket = new WebSocket(`ws://127.0.0.1:${port}${endpoint}`)
  socket.addEventListener('message', ({ data }) => {
    const m = JSON.parse(data)
    if (m.method === 'Runtime.exceptionThrown') exceptions.push(m.params.exceptionDetails.text)
    const task = waiting.get(m.id); if (!task) return
    clearTimeout(task.timer); waiting.delete(m.id)
    if (m.error) task.reject(new Error(m.error.message)); else task.resolve(m.result)
  })
  await new Promise((resolve, reject) => { socket.addEventListener('open', resolve, { once: true }); socket.addEventListener('error', reject, { once: true }) })
  const user = await tab('first@example.test'), mod = await tab('second@example.test'), admin = await tab('admin@example.test')
  await wait(user, "document.body.innerText.includes('Связь установлена')", 'realtime connected')
  const text = '<img src=x onerror="window.chatInjected=true"> Привет галактике'
  await fill(user, '#chat-body', text); await click(user, 'Отправить')
  await wait(mod, `document.querySelector('.chat-messages')?.textContent.includes(${JSON.stringify(text)})`, 'message reached other account')
  const firstMessageObservedAt = Date.now()
  assert.equal(await evaluate(mod, "!!window.chatInjected||document.querySelectorAll('.chat-body img').length>0"), false, 'message is escaped text')
  assert.equal(await evaluate(user, "[...document.querySelectorAll('button')].some(e=>e.textContent==='Удалить')"), false)
  assert.equal((await request(user, '/api/moderation/bans')).status, 403)
  await fill(mod, '#chat-body', 'Дежурный модератор на связи'); await click(mod, 'Отправить')
  await fill(admin, '#chat-body', 'Сообщение администратора'); await click(admin, 'Отправить')
  await wait(user, "document.querySelectorAll('.chat-message .account-nickname--administrator').length===1&&document.querySelectorAll('.chat-message .account-nickname--moderator').length===1", 'colored staff nicknames')
  const colors = await evaluate(user, `['user','moderator','administrator'].map(role=>getComputedStyle(document.querySelector('.chat-message .account-nickname--'+role)).color)`)
  assert.equal(new Set(colors).size, 3, 'three roles must have distinct visible nickname colors')
  await click(user, 'Список модераторов')
  await wait(user, "document.querySelectorAll('.moderator-directory > li').length===2", 'public staff directory')
  const directory = (await request(user, '/api/chat/moderators')).body.moderators
  assert.deepEqual(directory.map(member => member.role), ['administrator', 'moderator'])
  assert(directory.every(member => Object.keys(member).sort().join(',') === 'character_name,login,role,user_id'), 'directory exposes only public fields')
  await screenshot(user, 'moderators')
  await click(user, 'Вернуться в чат')
  await wait(user, "!!document.querySelector('#chat-body')&&!document.body.innerText.includes('Загрузка сообщений')", 'return from directory')
  for (const width of [320, 480, 1024]) {
    await command('Emulation.setDeviceMetricsOverride', { width, height: 900, deviceScaleFactor: 1, mobile: false }, mod)
    assert(await evaluate(mod, 'document.documentElement.scrollWidth<=document.documentElement.clientWidth'), 'chat layout overflow')
    await screenshot(mod, `chat-${width}`)
  }
  await evaluate(mod, `(() => { const row=[...document.querySelectorAll('.chat-message')].find(e=>e.querySelector('.chat-body')?.textContent===${JSON.stringify(text)}); [...row.querySelectorAll('button')].find(e=>e.textContent==='Удалить').click(); })()`)
  await fill(mod, '#delete-reason', 'Тестовое удаление'); await click(mod, 'Удалить сообщение')
  await wait(user, "document.querySelector('.chat-messages')?.textContent.includes('Сообщение удалено модератором')", 'deleted event')
  await click(mod, 'Ограничить First character')
  await wait(mod, "!!document.querySelector('#ban-reason')", 'moderation form')
  assert.equal(await evaluate(mod, "[...document.querySelector('#ban-scope').options].some(o=>o.value==='account')"), false)
  await fill(mod, '#ban-reason', 'Тестовая блокировка чата'); await click(mod, 'Выдать ограничение')
  await wait(user, "document.querySelector('#chat-body').disabled&&document.body.innerText.includes('Тестовая блокировка чата')", 'chat ban reaches active page')
  await screenshot(mod, 'moderation')
  assert.equal((await request(user, '/api/chat/messages', 'POST', { body: 'blocked', character_id: 1 }, 1)).status, 403)
  const bans = (await request(mod, '/api/moderation/bans')).body.bans
  const ban = bans.find(item => item.scope === 'chat' && item.user_id === 1)
  assert(ban)
  await click(mod, `Снять ограничение #${ban.id}`); await fill(mod, '#revoke-reason', 'Тестовая апелляция'); await click(mod, 'Снять ограничение')
  await wait(user, "!document.querySelector('#chat-body').disabled", 'unban enables composer')
  await fill(mod, '#ban-scope', 'forum'); await fill(mod, '#ban-reason', 'Тестовый бан форума'); await click(mod, 'Выдать ограничение')
  await wait(mod, "document.body.innerText.includes('Тестовый бан форума')&&document.querySelector('#ban-reason').value===''", 'forum ban issued')
  // Visiting the directory remounts the composer; the server still enforces
  // its account-wide interval from the first persisted message.
  await delay(Math.max(0, 2100 - (Date.now() - firstMessageObservedAt)))
  await fill(user, '#chat-body', 'Форумный бан не мешает чату')
  await wait(user, "[...document.querySelectorAll('button')].some(e=>e.textContent==='Отправить'&&!e.disabled)", 'send cooldown elapsed')
  await click(user, 'Отправить')
  await wait(admin, "document.querySelector('.chat-messages')?.textContent.includes('Форумный бан не мешает чату')", 'forum ban does not block chat')
  // A draft from one character must be discarded before another character sends.
  await fill(user, '#chat-body', 'Старый черновик')
  assert.equal((await request(user, '/api/characters/active', 'PUT', { character_id: 2 }, 1)).status, 200)
  await command('Page.bringToFront', {}, user)
  await evaluate(user, "window.dispatchEvent(new Event('focus'))")
  await wait(user, "document.querySelector('label[for=chat-body]')?.textContent.includes('Second character')&&document.querySelector('#chat-body').value===''", 'character draft reset')
  await delay(2100)
  await fill(user, '#chat-body', 'Сообщение второго персонажа'); await click(user, 'Отправить')
  await wait(admin, "document.querySelector('.chat-messages')?.textContent.includes('Сообщение второго персонажа')", 'second character message')
  const messages = (await request(admin, '/api/chat/messages')).body.messages
  assert.equal(messages.find(m => m.body === 'Сообщение второго персонажа').author_name, 'Second character')
  assert.equal(messages.find(m => m.body === 'Форумный бан не мешает чату').author_name, 'First character')
  await open(admin, '/moderation?user_id=1'); await wait(admin, "!!document.querySelector('#ban-scope')", 'admin moderation')
  await fill(admin, '#ban-scope', 'account'); await fill(admin, '#ban-duration', 'permanent'); await fill(admin, '#ban-reason', 'Проверка блокировки аккаунта'); await click(admin, 'Выдать ограничение')
  await wait(admin, "document.querySelector('#ban-reason').value===''", 'account ban committed')
  assert.equal((await request(user, '/api/auth/me')).status, 401, 'account ban revokes session')
  const denied = await request(user, '/api/auth/login', 'POST', { email: 'first@example.test', password: 'browser-password' })
  assert.equal(denied.status, 403); assert.equal(denied.body.code, 'account_banned')
  assert.equal(exceptions.length, 0, 'browser exceptions: ' + exceptions.join(';'))
  console.log('PASS social browser: realtime, escaped text, roles, deletion, scoped bans/revocation, character draft reset and permanent account ban')
} finally {
  if (socket?.readyState === WebSocket.OPEN) { await command('Browser.close').catch(() => {}); socket.close() }
  await Promise.race([exited, delay(3000)])
  if (browser.exitCode === null) browser.kill()
  for (const task of waiting.values()) clearTimeout(task.timer)
}
