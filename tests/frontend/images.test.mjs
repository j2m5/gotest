import assert from 'node:assert/strict'
import test from 'node:test'
import { localImagePath } from '../../frontend/src/shared/ui/localImagePath.ts'

test('catalog image paths resolve from the site root on nested routes', () => {
  assert.equal(localImagePath('images/classes/sith_war.png'), '/images/classes/sith_war.png')
  assert.equal(localImagePath('/images/factions/sith_empire.gif'), '/images/factions/sith_empire.gif')
})

test('database image values cannot make external requests or escape the image directory', () => {
  for (const value of [null, undefined, '', 'https://example.test/a.png', '//example.test/a.png',
    'data:image/png;base64,AA==', 'images/../api/logout.png', '/images/%2e%2e/a.png',
    'images\\a.png', '/images/a.png?redirect=x', '/api/a.png', 'images/a.svg']) {
    assert.equal(localImagePath(value), undefined, String(value))
  }
})
