import assert from 'node:assert/strict'
import { visitWorlds } from './world-content.browser-assertions.mjs'

export async function assertNeutralWorlds({ session, command, evaluate, waitFor, click, clickText, checkLayout, characterId, homeCode }) {
  const codes = [
    'belsavis', 'bothawui', 'dantooine', 'taris', 'hoth', 'tatooine', 'manaan', 'onderon',
    'makeb', 'nal_hutta', 'kashyyyk', 'trandosha', 'ryloth', 'rodia', 'sullust', 'ithor',
    'ilum', 'rishi', 'mandalore', 'concord_dawn', 'cathar', 'myrkr', 'iridonia', 'ambria',
    'sleheyron', 'toydaria', 'kwenn', 'nar_kreeta', 'klatooine', 'vodran', 'kintan', 'ubrikkia',
    'dorin', 'glee_anselm', 'gand', 'togoria', 'eres_iii', 'taanab', 'zeltros', 'cerea', 'bimmisaari',
    'malastare', 'kessel', 'eriadu', 'sluis_van', 'fondor', 'bestine_iv', 'denon', 'mygeeto', 'muunilinst',
  ]
  const catalog = () => evaluate(session, `fetch('/api/world/planets').then(r=>r.json()).then(r=>r.planets)`)
  async function assertClosed() {
    const planets = await catalog()
    const nav = await evaluate(session, `fetch('/api/world/navigation?character_id=${characterId}').then(r=>r.json())`)
    assert.equal(nav.position.planet.code, homeCode, 'Changing ownership does not move the player')
    for (const code of codes) {
      const planet = planets.find(p => p.code === code)
      assert.equal(planet.core_faction, null, `${code} is capturable`)
      assert.equal(planet.controlling_faction, null, `${code} starts/stays neutral`)
      const destination = nav.destinations.find(d => d.planet_id === planet.id)
      assert.equal(destination.can_travel, false)
      assert.equal(destination.reason, 'invasion_required')
      assert.equal(await evaluate(session, `!!document.querySelector('[data-planet-code="${code}"] button')`), false, `${code} stays closed without permission`)
    }
  }
  async function setControl(action) {
    // Only the Go browser harness recognizes this fixture header; the normal
    // server has no ownership mutation route or access bypass.
    assert.equal(await evaluate(session, `fetch('/api/world/navigation?character_id=${characterId}', {headers: {'X-Browser-Test-Neutral-Control': '${action}'}}).then(r=>r.status)`), 200)
    await command('Page.reload', {}, session)
    await waitFor(session, `document.querySelectorAll('.planet-list > li').length === 100 && document.querySelector('.world-position')?.dataset.planetCode === '${homeCode}' && !document.querySelector('[aria-busy="true"]')`, `neutral control ${action}`)
  }
  await assertClosed()
  await setControl('claim')
  const planets = await catalog()
  const faction = homeCode === 'coruscant' ? 'galactic_republic' : 'sith_empire'
  for (const code of codes) {
    assert.equal(planets.find(p => p.code === code).controlling_faction?.code, faction)
  }
  await visitWorlds({ session, command, evaluate, waitFor, click, clickText, checkLayout, characterId, homeCode, planets, codes })
  await setControl('release')
  await assertClosed()
}
