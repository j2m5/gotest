import assert from 'node:assert/strict'
import test from 'node:test'
import { apiError, needsEmailVerification } from '../../frontend/src/api/errors.ts'

test('preserves validation details from an API response', () => {
  assert.deepEqual(apiError({ response: { status: 422, data: {
    error: 'Validation failed', errors: ['Email is invalid', 'Password is too short'],
  } } }), { status: 422, code: undefined, fields: {}, message: 'Validation failed: Email is invalid; Password is too short' })
})

test('ignores malformed validation details and retains the server message', () => {
  assert.deepEqual(apiError({ response: { status: 403, data: {
    error: 'Verify your email', errors: [null, 2, ''],
  } } }), { status: 403, code: undefined, fields: {}, message: 'Verify your email' })
})

test('gives a useful fallback for network failures and unknown thrown values', () => {
  for (const error of [new Error('Network Error'), null, 'oops', undefined]) {
    assert.match(apiError(error).message, /соединение/)
  }
})

test('provides a generic fallback for malformed server responses', () => {
  assert.match(apiError({ response: { status: 500, data: '<html>error</html>' } }).message, /сервер/)
})

test('exposes stable codes and sanitized field messages without duplicating legacy details', () => {
  const failure = apiError({ response: { status: 422, data: {
    error: 'Validation failed', code: 'validation_failed',
    fields: { email: ['Invalid email', null, '', 4], password: ['Too short'], bad: 'ignored' },
    errors: ['Invalid email', 'Too short'],
  } } })
  assert.equal(failure.code, 'validation_failed')
  assert.deepEqual(Object.entries(failure.fields), [['email', ['Invalid email']], ['password', ['Too short']]])
  assert.equal(failure.message, 'Validation failed: Invalid email; Too short')
})

test('inline fields leave only general and unhandled details in the banner', () => {
  const failure = apiError({ response: { status: 422, data: {
    error: 'Validation failed', fields: { email: ['Invalid email'], other: ['Other issue'] },
    errors: ['Invalid email', 'Other issue', 'Legacy detail'],
  } } }, ['email'])
  assert.equal(failure.message, 'Validation failed: Other issue; Legacy detail')
})

test('rejects prototype field keys and malformed field containers safely', () => {
  const failure = apiError({ response: { data: JSON.parse('{"code":2,"fields":{"__proto__":["bad"],"constructor":["bad"],"prototype":["bad"],"email":["valid"]}}') } })
  assert.equal(failure.code, undefined)
  assert.deepEqual(Object.keys(failure.fields), ['email'])
  for (const fields of [null, [], 'bad', 4]) {
    assert.deepEqual(Object.keys(apiError({ response: { data: { fields } } }).fields), [])
  }
})

test('verification guidance requires both the specific code and forbidden status', () => {
  for (const [status, code, expected] of [[403, 'email_unverified', true], [403, 'origin_forbidden', false], [403, undefined, false], [401, 'email_unverified', false]]) {
    assert.equal(needsEmailVerification(apiError({ response: { status, data: { code } } })), expected)
  }
})
