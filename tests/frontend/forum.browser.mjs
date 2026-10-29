import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile, writeFile, mkdir } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'
import { assertUIControls } from './ui-kit.browser-assertions.mjs'

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
  await command('Runtime.enable', {}, sessionId)
  await command('Emulation.setDeviceMetricsOverride', { width: 480, height: 900, deviceScaleFactor: 1, mobile: false }, sessionId)
  await open(sessionId, '/login'); await wait(sessionId, "!!document.querySelector('#email')", 'login')
  await fill(sessionId, '#email', email); await fill(sessionId, '#password', 'browser-password')
  await evaluate(sessionId, "document.querySelector('button[type=submit]').click()")
  await wait(sessionId, "location.pathname==='/characters'||location.pathname==='/game'", 'login complete')
  await open(sessionId, '/chat'); await wait(sessionId, "!!document.querySelector('#chat-body')&&!document.body.innerText.includes('Загрузка сообщений')", 'chat ready')
  return sessionId
}
async function screenshot(session, name) {
  if (!process.env.FORUM_SCREENSHOT_DIR) return
  await mkdir(process.env.FORUM_SCREENSHOT_DIR, { recursive: true })
  const { data } = await command('Page.captureScreenshot', { format: 'png' }, session)
  await writeFile(path.join(process.env.FORUM_SCREENSHOT_DIR, name + '.png'), Buffer.from(data, 'base64'))
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
  const user = await tab('first@example.test'), mod = await tab('second@example.test')
  await open(user, '/forum')
  await wait(user, "document.querySelectorAll('.forum-categories a').length===4", 'forum categories available')
  await click(user, 'Список модераторов')
  await wait(user, "document.querySelectorAll('.moderator-directory > li').length===1", 'forum moderator directory')
  assert.equal(await evaluate(user, "document.querySelector('.moderator-directory .account-nickname')?.textContent"), 'second')
  assert.equal(await evaluate(user, "document.querySelectorAll('.moderator-directory .account-nickname--administrator').length"), 0)
  await open(user, '/forum')
  await wait(user, "document.querySelectorAll('.forum-categories a').length===4", 'return to forum categories')
  await click(user, 'Общение')
  await wait(user, "!!document.querySelector('#forum-title')", 'new topic form')
  await assertUIControls(evaluate, user, 'forum topic form')
  const text = '<img src=x onerror="window.forumInjected=true"> Привет форуму'
  await fill(user, '#forum-title', 'Первая тема')
  await fill(user, '#forum-body', text)
  await click(user, 'Создать тему')
  await wait(user, "!!document.querySelector('.forum-topic')", 'created topic')
  assert.equal(await evaluate(user, "!!window.forumInjected||document.querySelectorAll('.forum-body img').length>0"), false)
  assert(await evaluate(user, 'document.querySelector(".forum-topic .forum-body").textContent==='+JSON.stringify(text)))
  const topicId = Number(await evaluate(user, 'location.pathname.split("/").at(-1)'))
  const topicURL = `/forum/topics/${topicId}`
  await fill(user, '#forum-body', 'Draft from first character')
  assert.equal((await request(user, '/api/characters/active', 'PUT', { character_id: 2 }, 1)).status, 200)
  await evaluate(user, "window.dispatchEvent(new Event('focus'))")
  await wait(user, "document.querySelector('label[for=forum-body]')?.textContent.includes('Second character')&&document.querySelector('#forum-body').value===''", 'character change discards draft')
  await delay(5100)
  await fill(user, '#forum-body', 'Ответ второго персонажа'); await click(user, 'Отправить ответ')
  await wait(user, "document.querySelector('.forum-posts')?.textContent.includes('Ответ второго персонажа')", 'reply persisted')
  let data = (await request(user, `/api/forum/topic?topic_id=${topicId}`)).body
  assert.equal(data.topic.author_name, 'First character')
  assert.equal(data.posts[0].author_name, 'Second character')
  assert.equal(await evaluate(user, "document.querySelector('.forum-topic .account-nickname').closest('a').getAttribute('href')"), '/characters/1', 'topic author link survives character switch')
  assert.equal(await evaluate(user, "document.querySelector('.forum-posts .account-nickname').closest('a').getAttribute('href')"), '/characters/2', 'reply links to its actual author')
  await evaluate(user, "document.querySelector('.forum-topic .account-nickname').closest('a').click()")
  await wait(user, "location.pathname==='/characters/1'&&document.querySelector('h1')?.textContent==='First character'", 'original topic author profile')
  assert.equal((await request(user, '/api/characters')).body.active_character_id, 2, 'profile leaves viewer selection unchanged')
  await click(user, 'Вернуться в форум')
  await wait(user, "location.pathname==="+JSON.stringify(topicURL)+"&&!!document.querySelector('.forum-topic')", 'return to originating topic')
  const postId = data.posts[0].id
  await click(user, 'Редактировать')
  await fill(user, '#forum-edit-body', 'Мой сохранённый черновик')
  assert.equal((await request(user, '/api/forum/edit-topic', 'POST', { topic_id: topicId, character_id: 2, title: 'Первая тема', body: 'Изменение из другой вкладки', version: data.topic.version }, 1)).status, 200)
  await click(user, 'Сохранить изменения')
  await wait(user, "document.querySelector('[role=alert]')?.textContent.includes('Черновик сохранён')", 'stale version notice')
  assert.equal(await evaluate(user, "document.querySelector('#forum-edit-body').value"), 'Мой сохранённый черновик')
  await click(user, 'Обновить')
  await wait(user, "document.querySelector('.forum-topic .forum-body')?.textContent==='Изменение из другой вкладки'", 'conflict refresh displays current text')
  assert.equal(await evaluate(user, "document.querySelector('#forum-edit-body').value"), 'Мой сохранённый черновик')
  assert.equal(await evaluate(user, "[...document.querySelectorAll('button')].find(b=>b.textContent==='Сохранить изменения').disabled"), true, 'refreshed conflict requires explicit acknowledgement before saving')
  assert.equal((await request(user, `/api/forum/topic?topic_id=${topicId}`)).body.topic.body, 'Изменение из другой вкладки', 'refresh must not publish retained draft')
  await click(user, 'Продолжить с моим текстом')
  assert.equal(await evaluate(user, "document.querySelector('#forum-edit-body').value"), 'Мой сохранённый черновик', 'acknowledgement preserves draft')
  assert.equal((await request(user, `/api/forum/topic?topic_id=${topicId}`)).body.topic.body, 'Изменение из другой вкладки', 'acknowledgement must not publish retained draft')
  await click(user, 'Сохранить изменения')
  await wait(user, "!document.querySelector('#forum-edit-body')&&document.querySelector('.forum-topic .forum-body')?.textContent==='Мой сохранённый черновик'", 'retained conflict draft saved after acknowledgement')
  assert.equal((await request(user, `/api/forum/topic?topic_id=${topicId}`)).body.topic.body, 'Мой сохранённый черновик')
  await click(user, 'Редактировать'); await fill(user, '#forum-edit-body', text); await click(user, 'Сохранить изменения')
  await wait(user, "!document.querySelector('#forum-edit-body')", 'own edit saved')
  data = (await request(user, `/api/forum/topic?topic_id=${topicId}`)).body
  assert.equal(data.topic.author_name, 'First character', 'edit preserves original author after switching')
  await open(mod, topicURL); await wait(mod, "!!document.querySelector('.forum-topic')", 'moderator reads topic')
  async function act(selector, label) {
    await evaluate(mod, `(()=>{const root=document.querySelector(${JSON.stringify(selector)});const b=[...root.querySelectorAll('button')].find(e=>e.textContent===${JSON.stringify(label)});if(!b)throw new Error('Missing action');b.click()})()`)
    assert.equal(await evaluate(mod, "[...document.querySelectorAll('button')].find(b=>b.textContent==='Применить действие').disabled"), true, 'moderation requires a reason')
    await fill(mod, '#forum-reason', 'Проверка модерации форума')
    await click(mod, 'Применить действие')
    await wait(mod, "!document.querySelector('#forum-reason')", 'moderation saved')
  }
  await act(`[data-post-id="${postId}"]`, 'Скрыть')
  await click(user, 'Обновить')
  await wait(user, "document.querySelector('.forum-posts .forum-body')?.textContent==='Ответ скрыт модератором.'", 'hidden reply redacted')
  await act('.forum-topic', 'Скрыть')
  await click(user, 'Обновить'); await wait(user, "!!document.querySelector('[role=alert]')", 'hidden topic refresh denied')
  assert.equal(await evaluate(user, "!!document.querySelector('.forum-topic')"), false, 'refresh removes a now-hidden topic')
  await open(user, topicURL); await wait(user, "!!document.querySelector('[role=alert]')", 'hidden topic direct URL denied')
  assert.equal(await evaluate(user, "!!document.querySelector('.forum-topic')"), false)
  await open(mod, '/forum/categories/2'); await wait(mod, "!!document.querySelector('.forum-hidden-toggle')", 'staff hidden toggle')
  assert.equal(await evaluate(mod, "document.querySelector('.forum-topics').textContent.includes('Первая тема')"), false)
  await evaluate(mod, "document.querySelector('.forum-hidden-toggle input').click()")
  await wait(mod, "document.querySelector('.forum-topics')?.textContent.includes('Первая тема')", 'include hidden topics')
  await open(mod, topicURL); await wait(mod, "!!document.querySelector('.forum-topic')", 'staff reads hidden topic')
  await act('.forum-topic', 'Восстановить')
  await open(user, topicURL); await wait(user, "!!document.querySelector('.forum-topic')", 'topic restored')
  assert.equal(await evaluate(user, "document.querySelector('.forum-posts .forum-body').textContent"), 'Ответ скрыт модератором.', 'restoring parent preserves independent reply hiding')
  await act(`[data-post-id="${postId}"]`, 'Восстановить')
  await act('.forum-topic', 'Закрыть'); await click(user, 'Обновить')
  await wait(user, "document.querySelector('#forum-body')?.disabled&&document.body.innerText.includes('Тема закрыта')", 'closed topic disables replies')
  await act('.forum-topic', 'Открыть'); await act('.forum-topic', 'Закрепить')
  await click(user, 'Обновить'); await wait(user, "!document.querySelector('#forum-body').disabled", 'reopened topic')
  const forumBan = await request(mod, '/api/moderation/bans', 'POST', { user_id: 1, scope: 'forum', duration_hours: 1, reason: 'Форумный тестовый бан' }, 2)
  assert.equal(forumBan.status, 201)
  await click(user, 'Обновить'); await wait(user, "document.querySelector('#forum-body')?.disabled&&document.body.innerText.includes('Форумный тестовый бан')", 'forum ban allows reading but disables writing')
  assert.equal((await request(user, '/api/chat/messages', 'POST', { character_id: 2, body: 'Чат доступен при бане форума' }, 1)).status, 201)
  const bans = (await request(mod, '/api/moderation/bans')).body.bans
  assert.equal((await request(mod, '/api/moderation/revoke', 'POST', { ban_id: bans.find(b=>b.scope==='forum'&&b.user_id===1).id, reason: 'Снятие тестового бана' }, 2)).status, 200)
  assert.equal((await request(mod, '/api/moderation/bans', 'POST', { user_id: 1, scope: 'chat', duration_hours: 1, reason: 'Бан чата отдельно от форума' }, 2)).status, 201)
  await click(user, 'Обновить'); await wait(user, "!document.querySelector('#forum-body').disabled", 'chat ban does not block forum')
  await fill(mod, '#forum-body', 'Ответ модератора'); await click(mod, 'Отправить ответ')
  await wait(mod, "!!document.querySelector('.forum-posts .account-nickname--moderator')", 'shared role color')
  for (const width of [320, 480, 1024]) {
    await command('Emulation.setDeviceMetricsOverride', { width, height: 900, deviceScaleFactor: 1, mobile: false }, mod)
    assert(await evaluate(mod, 'document.documentElement.scrollWidth<=document.documentElement.clientWidth'), `forum layout overflow at ${width}`)
    await screenshot(mod, `forum-${width}`)
  }
  await open(mod, '/moderation?user_id=1&scope=forum'); await wait(mod, "document.querySelector('#ban-scope')?.value==='forum'", 'forum sanction link prefill')
  await wait(mod, "document.body.innerText.includes('Скрытие темы')&&document.body.innerText.includes('Восстановление ответа')", 'forum moderation journal')
  await open(user, '/forum/categories/3'); await wait(user, "document.querySelectorAll('.forum-topics>li').length===30", 'first topic page')
  await click(user, 'Следующие темы'); await wait(user, "document.querySelectorAll('.forum-topics>li').length===1", 'next topic page')
  await click(user, 'Paging topic 1'); await wait(user, "document.querySelectorAll('.forum-posts>li').length===30", 'first replies page')
  await click(user, 'Следующие ответы'); await wait(user, "document.querySelectorAll('.forum-posts>li').length===1", 'next replies page')
  assert.equal(await evaluate(user, "document.querySelector('.forum-posts .forum-body').textContent"), 'Reply 31')
  assert.equal(await evaluate(user, "document.querySelector('.forum-topic .forum-body').textContent"), 'Opening body', 'opening accompanies every reply page')
  assert.equal((await request(user, '/api/characters/active', 'PUT', { character_id: null }, 1)).status, 200)
  await open(user, topicURL); await wait(user, "document.body.innerText.includes('выберите персонажа')", 'read without selected character')
  assert.equal(await evaluate(user, "!!document.querySelector('#forum-body')"), false)
  await evaluate(user, "document.querySelector('.forum-topic .account-nickname').closest('a').click()")
  await wait(user, "location.pathname==='/characters/1'&&document.querySelector('h1')?.textContent==='First character'", 'signed-in viewer without selected character can read profile')
  assert.equal((await request(user, '/api/characters')).body.active_character_id, null)
  await open(user, '/characters/2147483647')
  await wait(user, "document.querySelector('[role=alert]')?.textContent.includes('Персонаж не найден')", 'missing public character guidance')
  assert.equal(await evaluate(user, "!!document.querySelector('.character-identity')"), false, 'missing character does not retain previous profile')
  assert.equal((await request(user, '/api/characters/active', 'PUT', { character_id: 2 }, 1)).status, 200)
  await open(user, topicURL); await wait(user, "!!document.querySelector('#forum-body')", 'selected character restored')
  await fill(user, '#forum-body', 'Old identity must never publish this draft')
  assert.equal((await request(user, '/api/auth/login', 'POST', { email: 'second@example.test', password: 'browser-password' })).status, 200)
  await click(user, 'Отправить ответ')
  await wait(user, "document.querySelector('label[for=forum-body]')?.textContent.includes('Other account character')&&document.querySelector('#forum-body').value===''", 'replacement account resets form after rejected stale write')
  data = (await request(user, `/api/forum/topic?topic_id=${topicId}`)).body
  assert(!data.posts.some(post=>post.body==='Old identity must never publish this draft'))
  assert.equal(exceptions.length, 0, 'browser exceptions: '+exceptions.join(';'))
  console.log('PASS forum browser: character authorship, plain text, editing/conflict drafts, role colors, hide/restore/close/pin, scoped bans, journal, no-character reading and responsive layout')
} finally {
  if (socket?.readyState === WebSocket.OPEN) { await command('Browser.close').catch(() => {}); socket.close() }
  await Promise.race([exited, delay(3000)])
  if (browser.exitCode === null) browser.kill()
  for (const task of waiting.values()) clearTimeout(task.timer)
}
