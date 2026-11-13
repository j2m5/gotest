import assert from 'node:assert/strict'
import { launchChrome } from './helpers/chrome.mjs'

const [baseURL, browser, profile, battleId, cookieName, token, experience, credits] = process.argv.slice(2)
const chrome = await launchChrome(browser, profile)
try {
  const { sessionId: tab } = await chrome.newTab()
  await chrome.command('Network.setCookie', { name: cookieName, value: token, url: baseURL, httpOnly: true, sameSite: 'Lax' }, tab)
  await chrome.command('Emulation.setDeviceMetricsOverride', { width: 375, height: 850, deviceScaleFactor: 1, mobile: false }, tab)
  await chrome.command('Page.navigate', { url: `${baseURL}/battle/${battleId}` }, tab)
  for (let pass = 0; pass < 2; pass++) {
    await chrome.waitFor(tab, "!!document.querySelector('[data-log-cleared]')", 'cleared history notice')
    assert.equal(await chrome.evaluate(tab, "document.querySelector('.combat-page').dataset.battleStatus"), 'won')
    assert.deepEqual(await chrome.evaluate(tab, "[...document.querySelectorAll('[data-event-kind]')].map(e=>e.dataset.eventKind)"), ['won'])
    const reward = await chrome.evaluate(tab, "document.querySelector('.combat-rewards').textContent")
    assert(reward.includes(`+${experience}`) && reward.includes(`+${credits}`), reward)
    assert.equal(await chrome.evaluate(tab, "document.querySelectorAll('.ability-bar button[aria-disabled=false]').length"), 0)
    assert(await chrome.evaluate(tab, 'document.documentElement.scrollWidth<=document.documentElement.clientWidth'), 'mobile overflow')
    if (pass === 0) {
      await chrome.command('Page.reload', {}, tab)
      await chrome.waitFor(tab, "document.readyState==='complete'", 'reload')
    }
  }
  assert.deepEqual(chrome.exceptions, [])
  console.log('PASS archived combat: notice, retained reward, no stale events, disabled abilities, reload and mobile layout')
} finally {
  await chrome.close()
}
