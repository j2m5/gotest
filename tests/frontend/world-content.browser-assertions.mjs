import assert from 'node:assert/strict'

export async function assertPopulatedWorlds({ session, command, evaluate, waitFor, click, clickText, checkLayout, characterId, homeCode, planets }) {
  const republic = [
    'arkania', 'balmorra', 'ord_mantell', 'carida', 'cato_neimoidia', 'shili', 'commenor', 'ralltiir', 'humbarine', 'neimoidia',
    'corellia', 'alderaan', 'kuat', 'chandrila', 'duro', 'brentaal_iv',
    'anaxes', 'alsakan', 'rendili',
    'caamas', 'esseles', 'rhinnal', 'corulag',
  ]
  const empire = [
    'athiss', 'thule', 'ossus', 'ord_radama', 'mirial', 'telos_iv', 'dathomir', 'saleucami', 'sriluur', 'dac',
    'ziost', 'begeren', 'bosthirda', 'jaguada', 'rhelg', 'khar_delba',
    'dromund_fels', 'krayiss_ii', 'chhodos',
    'ashas_ree', 'kalsunor', 'korriz', 'upekzar',
  ]
  const protectedCores = new Set([
    'corellia', 'alderaan', 'kuat', 'ziost', 'begeren', 'bosthirda',
    'chandrila', 'duro', 'brentaal_iv', 'jaguada', 'rhelg', 'khar_delba',
    'anaxes', 'alsakan', 'rendili', 'dromund_fels', 'krayiss_ii', 'chhodos',
    'caamas', 'esseles', 'rhinnal', 'corulag', 'ashas_ree', 'kalsunor', 'korriz', 'upekzar',
  ])
  const faction = homeCode === 'coruscant' ? 'galactic_republic' : 'sith_empire'
  const owned = homeCode === 'coruscant' ? republic : empire
  const opposing = homeCode === 'coruscant' ? empire : republic
  for (const code of owned) {
    assert.equal(planets.find(planet => planet.code === code).core_faction?.code ?? null, protectedCores.has(code) ? faction : null, `Protected core membership for ${code}`)
    assert.equal(await evaluate(session, `!!document.querySelector('[data-planet-code="${code}"] button:not(:disabled)')`), true, `Own populated world ${code} is available`)
  }
  for (const code of opposing) {
    assert.equal(await evaluate(session, `!!document.querySelector('[data-planet-code="${code}"] button')`), false, `Enemy world ${code} stays closed`)
  }
  await visitWorlds({ session, command, evaluate, waitFor, click, clickText, checkLayout, characterId, homeCode, planets, codes: owned })
}

export async function visitWorlds({ session, command, evaluate, waitFor, click, clickText, checkLayout, characterId, homeCode, planets, codes }) {
  const visitedNPCs = new Set()
  for (const code of codes) {
    await click(session, `[data-planet-code="${code}"] button`)
    await waitFor(session, `document.querySelector('.world-position')?.dataset.planetCode === '${code}' && !document.querySelector('[aria-busy="true"]')`, `world arrival ${code}`)
    const planet = planets.find(planet => planet.code === code)
    const catalog = await evaluate(session, `fetch('/api/world/locations?character_id=${characterId}&planet_id=${planet.id}').then(r=>r.json())`)
    assert.equal(catalog.locations.length, 10)
    assert.equal(catalog.position.location.id, catalog.locations[0].id, 'Flight arrives in the lowest-level location')
    assert.equal(catalog.position.location.recommended_level, 1)
    await clickText(session, 'Локации планеты')
    await waitFor(session, `document.querySelectorAll('.location-list > li').length === 10`, `world locations ${code}`)
    await checkLayout(session, `${code}-locations`)
    const last = catalog.locations[9]
    assert.equal(last.recommended_level, 46)
    await click(session, `[data-location-id="${last.id}"] button`)
    await waitFor(session, `!!document.querySelector('[data-current-location-id="${last.id}"]') && document.querySelectorAll('.npc-list > li').length === 4`, `world NPCs ${code}`)
    const ids = await evaluate(session, `[...document.querySelectorAll('.npc-list > li')].map(item=>Number(item.dataset.npcId))`)
    assert.equal(new Set(ids).size, 4)
    assert(ids.every(id => !visitedNPCs.has(id)), 'Different worlds have distinct shared NPCs')
    ids.forEach(id => visitedNPCs.add(id))
    assert.deepEqual(await evaluate(session, `[...document.querySelectorAll('.npc-meta span:first-child')].map(item=>item.textContent)`), ['Уровень 46', 'Уровень 46', 'Уровень 48', 'Уровень 50'])
    await checkLayout(session, `${code}-npcs`)
    await clickText(session, 'К галактике')
    await waitFor(session, `document.querySelector('.world-position')?.dataset.planetCode === '${code}'`, `world return to galaxy ${code}`)
  }
  await command('Page.reload', {}, session)
  await waitFor(session, `document.querySelector('.world-position')?.dataset.planetCode === '${codes.at(-1)}' && !document.querySelector('[aria-busy="true"]')`, 'world position persists')
  await click(session, `[data-planet-code="${homeCode}"] button`)
  await waitFor(session, `document.querySelector('.world-position')?.dataset.planetCode === '${homeCode}' && !document.querySelector('[aria-busy="true"]')`, 'return home from populated worlds')
}
