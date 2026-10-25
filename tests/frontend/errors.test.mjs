import assert from 'node:assert/strict'
import test from 'node:test'
import { authError } from '../../frontend/src/api/errors.ts'

test('preserves validation details from an API response', () => {
  assert.deepEqual(authError({ response: { status: 422, data: {
    error: 'Validation failed', errors: ['Email is invalid', 'Password is too short'],
  } } }), { status: 422, message: 'Validation failed: Email is invalid; Password is too short' })
})

test('ignores malformed validation details and retains the server message', () => {
  assert.deepEqual(authError({ response: { status: 403, data: {
    error: 'Verify your email', errors: [null, 2, ''],
  } } }), { status: 403, message: 'Verify your email' })
})

test('gives a useful fallback for network failures and unknown thrown values', () => {
  for (const error of [new Error('Network Error'), null, 'oops', undefined]) {
    assert.match(authError(error).message, /соединение/)
  }
})

test('provides a generic fallback for malformed server responses', () => {
  assert.match(authError({ response: { status: 500, data: '<html>error</html>' } }).message, /сервер/)
})
