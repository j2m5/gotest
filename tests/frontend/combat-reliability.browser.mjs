import assert from 'node:assert/strict'
import { setTimeout as delay } from 'node:timers/promises'
import { launchChrome } from './helpers/chrome.mjs'

const [baseURL, binary, profile, planetId, locationId] = process.argv.slice(2)
const chrome = await launchChrome(binary, profile)
const { command, evaluate, waitFor, newTab, nextEvent, press } = chrome
const traffic = [], responseStatus = new Map()
chrome.on(event => {
  if (event.method === 'Network.requestWillBeSent' && /\/api\/(battles\/|abilities)/.test(event.params.request.url)) {
    const { url, method, postData } = event.params.request
    traffic.push({ tab: event.sessionId, id: event.params.requestId, url, method, body: postData ? JSON.parse(postData) : null })
  }
  if (event.method === 'Network.responseReceived') responseStatus.set(event.params.requestId, event.params.response.status)
})
const locationPath = `/galaxy/planets/${planetId}/locations/${locationId}`
const ready = "!!document.querySelector('.ability-bar button[aria-disabled=false]')"
const count = tab => evaluate(tab.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length")
const writes = tab => traffic.filter(r => r.tab === tab.sessionId && r.url.endsWith('/api/battles/use') && r.method === 'POST')
async function open(tab, route) { await command('Page.navigate', { url: baseURL + route }, tab.sessionId) }
async function click(tab, selector) {
  await waitFor(tab.sessionId, `(()=>{const b=document.querySelector(${JSON.stringify(selector)});if(!b||b.disabled||b.getAttribute('aria-disabled')==='true')return false;b.click();return true})()`, 'click ' + selector)
}
async function clickText(tab, text) {
  await waitFor(tab.sessionId, `(()=>{const b=[...document.querySelectorAll('button,a')].find(b=>b.textContent.trim()===${JSON.stringify(text)});if(!b||b.disabled)return false;b.click();return true})()`, 'click ' + text)
}
async function request(tab, path, method = 'GET', body) {
  return evaluate(tab.sessionId, `(async()=>{const r=await fetch(${JSON.stringify(path)},{method:${JSON.stringify(method)},headers:{'Content-Type':'application/json','X-Account-ID':${JSON.stringify(String(tab.accountId))}},${body ? 'body:JSON.stringify('+JSON.stringify(body)+'),' : ''}});return {status:r.status,body:await r.json()}})()`)
}
async function login(accountId, characterId, email) {
  const tab = { ...await newTab(), accountId, characterId }
  // Keep references to native sockets only to inject a real transport close.
  await command('Page.addScriptToEvaluateOnNewDocument', { source: `window.__combatSockets=[];const NativeSocket=WebSocket;window.WebSocket=class extends NativeSocket{constructor(...args){super(...args);if(String(args[0]).includes('/api/battles/events'))window.__combatSockets.push(this)}}` }, tab.sessionId)
  await open(tab, '/login'); await waitFor(tab.sessionId, "!!document.querySelector('#email')", 'login form')
  await evaluate(tab.sessionId, `(()=>{for(const [id,value] of [['email',${JSON.stringify(email)}],['password','browser-password']]){const input=document.getElementById(id);Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set.call(input,value);input.dispatchEvent(new Event('input',{bubbles:true}))}document.querySelector('button[type=submit]').click()})()`)
  await waitFor(tab.sessionId, "location.pathname==='/game'||location.pathname==='/characters'", 'login')
  await open(tab, locationPath)
  await waitFor(tab.sessionId, "document.querySelectorAll('[data-attack-npc]:not(:disabled)').length===4", 'available NPCs')
  return tab
}
async function focusAbility(tab) {
  await waitFor(tab.sessionId, ready, 'ability ready')
  await evaluate(tab.sessionId, "document.querySelector('.ability-bar button').focus()")
}
async function events(tab, battleId) {
  const result = await request(tab, `/api/battles/detail?character_id=${tab.characterId}&battle_id=${battleId}`)
  assert.equal(result.status, 200)
  return result.body.events.filter(event => event.kind === 'player_ability')
}
try {
  const a = await login(1, 2, 'first@example.test'), b = await login(2, 3, 'second@example.test')
  const npcId = await evaluate(a.sessionId, "document.querySelector('[data-attack-npc]').dataset.attackNpc")
  await Promise.all([click(a, `[data-attack-npc="${npcId}"]`), click(b, `[data-attack-npc="${npcId}"]`)])
  for (const tab of [a,b]) await waitFor(tab.sessionId, "!!document.querySelector('[data-battle-status=active]')||!!document.querySelector('main [role=alert]')", 'contested NPC result')
  const winner = await evaluate(a.sessionId, "!!document.querySelector('[data-battle-status=active]')") ? a : b
  const loser = winner === a ? b : a
  const battlePath = await evaluate(winner.sessionId, 'location.pathname'), battleId = battlePath.split('/').at(-1)
  assert.equal(await evaluate(loser.sessionId, "!!document.querySelector('[data-battle-status]')"), false)
  const available = await request(loser, `/api/npcs?character_id=${loser.characterId}&location_id=${locationId}`)
  assert.equal(available.status, 200); assert(!available.body.npcs.some(npc => String(npc.id) === npcId))
  assert.equal((await request(loser, `/api/battles/detail?character_id=${loser.characterId}&battle_id=${battleId}`)).status, 404)
  console.log('PASS contested shared NPC and foreign journal denial')

  const other = { ...await newTab(winner.browserContextId), accountId: winner.accountId, characterId: winner.characterId }
  await command('Page.addScriptToEvaluateOnNewDocument', { source: `window.__combatSockets=[];const NativeSocket=WebSocket;window.WebSocket=class extends NativeSocket{constructor(...args){super(...args);if(String(args[0]).includes('/api/battles/events'))window.__combatSockets.push(this)}}` }, other.sessionId)
  await open(other, battlePath); await waitFor(other.sessionId, ready, 'same account second tab')
  await focusAbility(winner); await press(winner.sessionId, 'Enter')
  await waitFor(winner.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===1", 'Enter applies ability')
  await press(winner.sessionId, 'Enter'); await press(winner.sessionId, ' ')
  assert.equal((await events(winner, battleId)).length, 1, 'cooldown keyboard presses do not spend')
  assert(await evaluate(winner.sessionId, "document.activeElement===document.querySelector('.ability-bar button')"), 'focus retained')
  await waitFor(other.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===1&&document.querySelector('.ability-bar button')?.getAttribute('aria-disabled')==='true'", 'other tab sees CD')
  await focusAbility(other); await press(other.sessionId, ' ')
  await waitFor(winner.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===2", 'Space propagates to both tabs')
  console.log('PASS real Enter/Space, blocked keyboard presses and shared cooldown')

  await Promise.all([focusAbility(winner), focusAbility(other)])
  const held = [winner,other].map(tab => nextEvent(tab.sessionId, 'Fetch.requestPaused'))
  for (const tab of [winner,other]) await command('Fetch.enable', { patterns: [{ urlPattern: '*/api/battles/use', requestStage: 'Request' }] }, tab.sessionId)
  await Promise.all([press(winner.sessionId, 'Enter'), press(other.sessionId, 'Enter')])
  const paused = await Promise.all(held)
  await Promise.all(paused.map((event,index) => command('Fetch.continueRequest', { requestId: event.requestId }, [winner,other][index].sessionId)))
  for (const tab of [winner,other]) await command('Fetch.disable', {}, tab.sessionId)
  await waitFor(winner.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===3", 'simultaneous commands reconcile')
  await waitFor(other.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===3", 'both tabs see one effect')
  assert.equal((await events(winner, battleId)).length, 3)
  const racing = [writes(winner).at(-1), writes(other).at(-1)]
  assert.deepEqual(racing.map(r=>responseStatus.get(r.id)).sort(), [200,409])
  console.log('PASS two distinct simultaneous commands cannot bypass cooldown')

  await focusAbility(winner)
  const lostResponse = nextEvent(winner.sessionId, 'Fetch.requestPaused')
  await command('Fetch.enable', { patterns: [{ urlPattern: '*/api/battles/use', requestStage: 'Response' }] }, winner.sessionId)
  await press(winner.sessionId, 'Enter')
  const lost = await lostResponse
  assert.equal(lost.responseStatusCode, 200, 'server committed before transport loss')
  const lostBody = JSON.parse(lost.request.postData)
  const committed = await request(other, `/api/battles/detail?character_id=${other.characterId}&battle_id=${battleId}`)
  assert.equal(committed.status, 200)
  assert.equal(committed.body.events.filter(event => event.kind === 'player_ability').length, 4)
  await command('Fetch.failRequest', { requestId: lost.requestId, errorReason: 'ConnectionClosed' }, winner.sessionId)
  await command('Fetch.disable', {}, winner.sessionId)
  await waitFor(winner.sessionId, "document.body.innerText.includes('Повторить действие')", 'explicit retry after lost response')
  const sent = writes(winner).length
  await delay(300)
  assert.equal(writes(winner).length, sent, 'no automatic write retry')
  assert(await evaluate(winner.sessionId, "document.querySelector('.ability-bar button').getAttribute('aria-disabled')==='true'"), 'unknown command blocks next ability')
  await clickText(winner, 'Повторить действие')
  await waitFor(winner.sessionId, "!document.body.innerText.includes('Повторить действие')&&document.querySelectorAll('[data-event-kind=player_ability]').length===4", 'replay acknowledged')
  assert.deepEqual(writes(winner).at(-1).body, lostBody, 'explicit replay keeps the complete command')
  assert.equal((await events(winner, battleId)).length, 4, 'replay adds no second effect')
  const recovered = await request(winner, `/api/battles/current?character_id=${winner.characterId}`)
  assert.equal(recovered.status, 200)
  const energy = value => [value.regeneration.mp_baseline, value.regeneration.mp_updated_at, value.regeneration.mp_remainder]
  assert.deepEqual(energy(recovered.body.character), energy(committed.body.character), 'replay preserves the stored energy baseline, even if regeneration masks a debit in current MP')
  console.log('PASS lost committed response and explicit idempotent retry')

  await command('Network.emulateNetworkConditions', { offline: true, latency: 0, downloadThroughput: 0, uploadThroughput: 0 }, other.sessionId)
  await evaluate(other.sessionId, 'window.__combatSockets.forEach(socket=>socket.close())')
  await focusAbility(winner); await press(winner.sessionId, 'Enter')
  await waitFor(winner.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===5", 'online tab continues')
  assert((await count(other)) < 5, 'disconnected tab has not received the effect')
  await command('Network.emulateNetworkConditions', { offline: false, latency: 0, downloadThroughput: -1, uploadThroughput: -1 }, other.sessionId)
  await waitFor(other.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===5&&!document.querySelector('.combat-connection')", 'reconnect catches up without reload')
  console.log('PASS actual socket close/offline/reconnect')

  await command('Emulation.setFocusEmulationEnabled', { enabled: false }, other.sessionId)
  await command('Page.setWebLifecycleState', { state: 'frozen' }, other.sessionId)
  assert.equal(await evaluate(other.sessionId, 'document.visibilityState'), 'hidden')
  await delay(150)
  const readsBefore = traffic.filter(r => r.tab === other.sessionId && r.method === 'GET').length
  await focusAbility(winner); await press(winner.sessionId, 'Enter')
  await waitFor(winner.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===6", 'active tab acts while peer frozen')
  await delay(5200)
  assert.equal(traffic.filter(r => r.tab === other.sessionId && r.method === 'GET').length, readsBefore, 'frozen tab sends no polling reads')
  await command('Page.setWebLifecycleState', { state: 'active' }, other.sessionId)
  await command('Page.bringToFront', {}, winner.sessionId)
  assert.equal(await evaluate(other.sessionId, 'document.visibilityState'), 'hidden', 'resumed background tab remains hidden')
  await delay(150)
  const hiddenReads = traffic.filter(r => r.tab === other.sessionId && r.method === 'GET').length
  await delay(5200)
  assert.equal(traffic.filter(r => r.tab === other.sessionId && r.method === 'GET').length, hiddenReads, 'awake hidden tab suppresses HTTP polling and socket-triggered reads')
  await command('Emulation.setFocusEmulationEnabled', { enabled: true }, other.sessionId)
  await waitFor(other.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===6", 'thaw reconciles missed events')
  console.log('PASS real frozen page, awake hidden tab and foreground reconciliation')

  await focusAbility(winner); await press(winner.sessionId, 'Enter')
  await waitFor(winner.sessionId, "document.querySelectorAll('[data-event-kind=player_ability]').length===7", 'last action before retreat')
  await click(winner, '[data-combat-retreat]')
  await waitFor(winner.sessionId, "!!document.querySelector('[data-battle-status=retreated]')", 'retreat')
  await click(winner, '[data-combat-return]'); await click(winner, '[data-attack-npc]')
  await waitFor(winner.sessionId, "!!document.querySelector('[data-battle-status=active]')&&!!document.querySelector('.ability-bar button')", 'new battle')
  assert(await evaluate(winner.sessionId, "document.querySelector('.ability-bar button').getAttribute('aria-disabled')==='true'"), 'cooldown carries across encounters')
  const resources = (await request(winner, `/api/battles/current?character_id=${winner.characterId}`)).body.character
  assert(resources.mp_current < 100, 'retreat did not refill energy')
  await click(winner, '[data-combat-retreat]')
  await waitFor(winner.sessionId, "!!document.querySelector('[data-battle-status=retreated]')", 'final cleanup retreat')
  assert.equal(chrome.exceptions.length, 0, chrome.exceptions.join('; '))
  console.log('PASS preserved resources and cooldown across encounters')
} finally { await chrome.close() }
