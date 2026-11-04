import assert from 'node:assert/strict'
import test from 'node:test'
import { formatAmount } from '../../frontend/src/features/wallet/format.ts'

test('currency formatting preserves all signed int64 digits without floating point rounding', () => {
  assert.equal(formatAmount('9223372036854775807').replace(/\s/g, ''), '9223372036854775807')
  assert.equal(formatAmount('-9223372036854775807').replace(/\s/g, ''), '-9223372036854775807')
  assert.equal(formatAmount('9007199254740993', true).replace(/\s/g, ''), '+9007199254740993')
  assert.equal(formatAmount('-123', true), '-123')
  assert.equal(formatAmount('0'), '0')
})
