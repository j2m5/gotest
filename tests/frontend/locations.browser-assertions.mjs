import assert from 'node:assert/strict'
import { assertNPCs } from './npcs.browser-assertions.mjs'

export async function assertLocations({ session, command, evaluate, waitFor, click, clickText, checkLayout, nextEvent, characterId, planetId, otherPlanetId }) {
  const listPath = `/galaxy/planets/${planetId}/locations`
  const readPosition = () => evaluate(session, `fetch('/api/world/navigation?character_id=${characterId}').then(r=>r.json()).then(r=>r.position)`)
  const initial = await readPosition()
  const catalog = await evaluate(session, `fetch('/api/world/locations?character_id=${characterId}&planet_id=${planetId}').then(r=>r.json())`)
  const target = catalog.locations[9]
  await clickText(session, 'Локации планеты')
  await waitFor(session, `location.pathname === '${listPath}' && document.querySelectorAll('.location-list > li').length === 10`, 'ten capital locations')
  await checkLayout(session, 'locations')
  await click(session, `[data-location-id="${target.id}"] .location-title`)
  await waitFor(session, `document.querySelector('h1')?.textContent === ${JSON.stringify(target.name)}`, 'location preview')
  assert.equal((await readPosition()).location.id, initial.location.id, 'Preview never moves the character')
  assert.equal(await evaluate(session, `!!document.querySelector('.npc-list')`), false, 'Preview does not reveal the live NPC pool')
  assert.equal(await evaluate(session, `!!document.querySelector('.location-enter:not(:disabled)')`), true, 'Level46 recommendation allows a level1 player')
  await click(session, '.location-enter')
  await waitFor(session, `document.querySelector('[data-current-location-id="${target.id}"]') && !document.querySelector('[aria-busy="true"]')`, 'enter location')
  await command('Page.reload', {}, session)
  await waitFor(session, `!!document.querySelector('[data-current-location-id="${target.id}"]')`, 'location survives reload')
  await checkLayout(session, 'location-detail')
  await evaluate(session, 'history.back()')
  await waitFor(session, `location.pathname === '${listPath}' && document.querySelectorAll('.location-list > li').length === 10`, 'back to location list')
  assert.equal((await readPosition()).location.id, target.id, 'Back navigation does not move the character')

  await command('Network.setExtraHTTPHeaders', { headers: { 'X-Browser-Test-Locations-Failure': '1' } }, session)
  await command('Page.reload', {}, session)
  await waitFor(session, `document.querySelector('main [role="alert"]') && !document.querySelector('.location-list')`, 'locations failure')
  await command('Network.setExtraHTTPHeaders', { headers: {} }, session)
  await clickText(session, 'Повторить')
  await waitFor(session, `document.querySelectorAll('.location-list > li').length === 10`, 'locations retry')

  const travel = await evaluate(session, `fetch('/api/world/travel',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({character_id:${characterId},planet_id:${otherPlanetId},from_location_id:${target.id}})}).then(r=>({status:r.status}))`)
  assert.equal(travel.status, 200)
  await command('Page.reload', {}, session)
  await waitFor(session, `document.querySelector('main [role="alert"]')?.textContent.includes('текущей планете')`, 'stale planet URL')
  const elsewhere = await readPosition()
  assert.equal(elsewhere.planet.id, otherPlanetId, 'Stale URL does not teleport back')
  await clickText(session, 'К галактике')
  await waitFor(session, `location.pathname === '/galaxy' && document.querySelector('.world-position')?.dataset.planetCode === '${elsewhere.planet.code}'`, 'galaxy after stale URL')
  await click(session, `[data-planet-code="${initial.planet.code}"] button`)
  await waitFor(session, `document.querySelector('.world-position')?.dataset.planetCode === '${initial.planet.code}' && !document.querySelector('[aria-busy="true"]')`, 'return from location tests')
  await assertNPCs({ session, command, evaluate, waitFor, click, clickText, checkLayout, nextEvent, characterId, planetId, otherPlanetId, locations: catalog.locations, homeCode: initial.planet.code })
}
