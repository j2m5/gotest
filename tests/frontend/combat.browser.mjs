import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { readFile, writeFile, mkdir } from 'node:fs/promises'
import path from 'node:path'
import { setTimeout as delay } from 'node:timers/promises'

const [baseURL, browserPath, profilePath, planetId, locationId] = process.argv.slice(2)
const browser = spawn(browserPath, ['--headless=new', '--remote-debugging-port=0', `--user-data-dir=${profilePath}`, '--disable-background-networking', '--no-first-run', '--no-default-browser-check', '--disable-extensions', '--disable-gpu', 'about:blank'], { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] })
let socket, sequence = 0, log = ''
const requests = new Map(), exceptions = []
browser.stderr.on('data', data => { log = (log + data).slice(-4000) })
const exited = new Promise(resolve => browser.once('exit', resolve))
function command(method, params = {}, sessionId) {
  return new Promise((resolve,reject) => {
    const id = ++sequence, timer = setTimeout(() => { requests.delete(id); reject(new Error('CDP timeout ' + method)) },10000)
    requests.set(id,{ resolve,reject,timer }); socket.send(JSON.stringify({ id,method,params,sessionId }))
  })
}
async function evaluate(sessionId,expression) {
  const result = await command('Runtime.evaluate',{expression,awaitPromise:true,returnByValue:true},sessionId)
  if(result.exceptionDetails) throw new Error(result.exceptionDetails.exception?.description ?? result.exceptionDetails.text)
  return result.result.value
}
async function wait(sessionId,expression,label) {
  const deadline=Date.now()+10000
  while(Date.now()<deadline) { if(await evaluate(sessionId,`Boolean(${expression})`)) return; await delay(60) }
  throw new Error(label+': '+await evaluate(sessionId,'document.body.innerText'))
}
async function open(sessionId,route) { await command('Page.navigate',{url:baseURL+route},sessionId) }
async function click(sessionId,selector) {
  await wait(sessionId,`(()=>{const b=document.querySelector(${JSON.stringify(selector)});if(!b||b.disabled||b.getAttribute('aria-disabled')==='true')return false;b.click();return true})()`,'click '+selector)
}
async function request(sessionId,url,method='GET',body) {
  return evaluate(sessionId,`(async()=>{const r=await fetch(${JSON.stringify(url)},{method:${JSON.stringify(method)},headers:{'Content-Type':'application/json','X-Account-ID':'1'},${body ? 'body:JSON.stringify('+JSON.stringify(body)+'),' : ''}});return {status:r.status,body:await r.json()}})()`)
}
async function screenshot(sessionId,name) {
  if(!process.env.COMBAT_SCREENSHOT_DIR) return
  await mkdir(process.env.COMBAT_SCREENSHOT_DIR,{recursive:true})
  const {data}=await command('Page.captureScreenshot',{format:'png'},sessionId)
  await writeFile(path.join(process.env.COMBAT_SCREENSHOT_DIR,name+'.png'),Buffer.from(data,'base64'))
}
try {
  let port,endpoint
  for(let i=0;i<150&&!endpoint;i++) { try{[port,endpoint]=(await readFile(path.join(profilePath,'DevToolsActivePort'),'utf8')).trim().split('\n')}catch{await delay(100)} }
  assert(endpoint,'Chrome unavailable: '+log)
  socket=new WebSocket(`ws://127.0.0.1:${port}${endpoint}`)
  socket.addEventListener('message',({data})=>{
    const message=JSON.parse(data)
    if(message.method==='Runtime.exceptionThrown') exceptions.push(message.params.exceptionDetails.text)
    const task=requests.get(message.id);if(!task)return
    clearTimeout(task.timer);requests.delete(message.id)
    if(message.error)task.reject(new Error(message.error.message));else task.resolve(message.result)
  })
  await new Promise((resolve,reject)=>{socket.addEventListener('open',resolve,{once:true});socket.addEventListener('error',reject,{once:true})})
  const {targetId}=await command('Target.createTarget',{url:'about:blank'})
  const {sessionId:s}=await command('Target.attachToTarget',{targetId,flatten:true})
  await command('Page.enable',{},s);await command('Runtime.enable',{},s)
  await command('Emulation.setDeviceMetricsOverride',{width:550,height:900,deviceScaleFactor:1,mobile:false},s)
  await open(s,'/login')
  await wait(s,"!!document.querySelector('#email')",'login form')
  await evaluate(s,`(()=>{for(const [id,value] of [['email','first@example.test'],['password','browser-password']]) {const input=document.getElementById(id);Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set.call(input,value);input.dispatchEvent(new Event('input',{bubbles:true}))}document.querySelector('button[type=submit]').click()})()`)
  await wait(s,"location.pathname==='/game'||location.pathname==='/characters'",'login')
  await open(s,`/galaxy/planets/${planetId}/locations/${locationId}`)
  await wait(s,"document.querySelectorAll('[data-attack-npc]').length===4",'attack buttons')
  const npcId=await evaluate(s,"document.querySelector('[data-attack-npc]').dataset.attackNpc")
  await wait(s,"(()=>{const b=document.querySelector('[data-attack-npc]');if(!b||b.disabled)return false;b.click();b.click();return true})()",'double click attack')
  await wait(s,"location.pathname.startsWith('/battle/')&&document.querySelector('[data-battle-status=active]')",'battle opened')
  const battlePath=await evaluate(s,'location.pathname'), battleId=battlePath.split('/').at(-1)
  await wait(s,"!!document.querySelector('[data-event-kind=npc_ability]')",'autonomous NPC attack')
  await command('Page.reload',{},s)
  await wait(s,"!!document.querySelector('.ability-bar button[aria-disabled=false]')",'active battle restored')
  assert.equal(await evaluate(s,'location.pathname'),battlePath)
  assert.equal((await request(s,'/api/characters/active','PUT',{character_id:1})).status,200)
  await open(s,battlePath)
  await wait(s,"!!document.querySelector('.combat-page')||!!document.querySelector('main [role=alert]')",'different character view')
  assert.equal(await evaluate(s,"!!document.querySelector('[data-battle-status]')"),false,'another selected character cannot view this battle')
  assert.equal((await request(s,'/api/characters/active','PUT',{character_id:2})).status,200)
  await open(s,'/game')
  await wait(s,"!!document.querySelector('.active-battle-banner a')",'ongoing battle banner after character switch')
  await click(s,'.active-battle-banner a')
  await wait(s,"!!document.querySelector('.ability-bar button[aria-disabled=false]')",'return to the same battle')
  assert.equal(await evaluate(s,'location.pathname'),battlePath)
  await evaluate(s,"{const b=document.querySelector('.ability-bar button');b.focus();b.click();b.click()}")
  await wait(s,"document.querySelectorAll('[data-event-kind=player_ability]').length===1",'one accepted ability')
  assert(await evaluate(s,"document.querySelector('.ability-bar button').getAttribute('aria-disabled')==='true'"),'cooldown visible')
  assert(await evaluate(s,"document.activeElement===document.querySelector('.ability-bar button')"),'keyboard focus survives the command')
  await wait(s,"document.querySelector('.ability-icon img')?.complete",'ability icon loaded')
  assert(await evaluate(s,"!document.querySelector('.ability-icon img').src.endsWith('/neutral.png')"),'original ability icon is available')
  const first=await request(s,`/api/battles/detail?battle_id=${battleId}&character_id=2`)
  assert.equal(first.body.events.filter(e=>e.kind==='player_ability').length,1,'double click spends only once')
  assert(first.body.character.mp_current<100,'energy spent')
  for(const width of [320,375,550]) {
    await command('Emulation.setDeviceMetricsOverride',{width,height:900,deviceScaleFactor:1,mobile:false},s)
    assert(await evaluate(s,'document.documentElement.scrollWidth<=document.documentElement.clientWidth'),'combat overflow '+width)
    await screenshot(s,'battle-active-'+width)
  }
  for(let i=0;i<6;i++){
    if(await evaluate(s,"!!document.querySelector('[data-battle-status=won]')"))break
    await wait(s,"(()=>{if(document.querySelector('[data-battle-status=won]'))return true;const b=document.querySelector('.ability-bar button[aria-disabled=false]');if(!b)return false;b.click();return true})()",'next ability or victory')
    await delay(100)
  }
  await wait(s,"!!document.querySelector('[data-battle-status=won]')&&!!document.querySelector('[data-event-kind=won]')",'victory and log')
  await wait(s,"document.querySelector('.combat-rewards')?.textContent.includes('3')",'reward summary')
  await screenshot(s,'battle-won')
  const wallet=await request(s,'/api/wallet?character_id=2')
  assert.equal(wallet.body.credits,'3')
  await click(s,'[data-combat-return]')
  await wait(s,"document.querySelectorAll('[data-attack-npc]').length===3",'defeated NPC waits for respawn')
  assert.equal(await evaluate(s,`!!document.querySelector('[data-attack-npc="${npcId}"]')`),false)
  await click(s,'[data-attack-npc]')
  await wait(s,"!!document.querySelector('[data-battle-status=active]')",'second encounter')
  await click(s,'.ability-bar button')
  await wait(s,"!!document.querySelector('[data-event-kind=player_ability]')",'damage before retreat')
  await click(s,'[data-combat-retreat]')
  await wait(s,"!!document.querySelector('[data-battle-status=retreated]')",'retreat')
  const secondId=(await evaluate(s,'location.pathname')).split('/').at(-1)
  const ended=await request(s,`/api/battles/detail?battle_id=${secondId}&character_id=2`)
  assert.equal(ended.body.battle.npc.hp_current,ended.body.battle.npc.hp,'retreat restores NPC')
  await screenshot(s,'battle-retreated')
  assert.equal((await request(s,'/api/wallet?character_id=2')).body.credits,'3')
  await open(s,'/battle/01')
  await wait(s,"document.body.innerText.includes('Бой не найден')",'invalid bigint route')
  assert.equal((await request(s,'/api/characters/active','PUT',{character_id:1})).status,200)
  const navigation=await request(s,'/api/world/navigation?character_id=1')
  assert.equal(navigation.status,200)
  await open(s,`/galaxy/planets/${navigation.body.position.planet.id}/locations/${navigation.body.position.location.id}`)
  await click(s,'[data-attack-npc]')
  await wait(s,"document.body.innerText.includes('пока нет изученных умений')",'empty ability bar')
  assert.equal(await evaluate(s,"document.querySelectorAll('.ability-bar button').length"),0)
  await click(s,'[data-combat-retreat]')
  await wait(s,"!!document.querySelector('[data-battle-status=retreated]')",'empty loadout can retreat')
  assert.equal(exceptions.length,0,exceptions.join('; '))
  console.log('PASS combat browser: NPC entry, double click, autonomous attack, reload, character switch, keyboard focus, icons, CD/energy, victory/rewards, respawn, retreat, empty loadout and 320/375/550px')
} finally {
  if(socket?.readyState===WebSocket.OPEN){await command('Browser.close').catch(()=>{});socket.close()}
  await Promise.race([exited,delay(3000)])
  if(browser.exitCode===null)browser.kill()
  for(const task of requests.values())clearTimeout(task.timer)
}
