import assert from 'node:assert/strict'
import test from 'node:test'
import { validateIntegerScore } from '../../src/lib/validators/score.ts'

test('accepts inclusive integer score boundaries', () => {
  assert.deepEqual(validateIntegerScore('0', 100), { valid: true, score: 0 })
  assert.deepEqual(validateIntegerScore('100', 100), { valid: true, score: 100 })
  assert.deepEqual(validateIntegerScore(' 42 ', 100), { valid: true, score: 42 })
})

test('rejects empty, fractional, exponential, and non-finite scores without coercion', () => {
  for (const rawScore of ['', '   ', '1.5', '1e2', 'Infinity', 'NaN', 'not-a-score']) {
    assert.equal(validateIntegerScore(rawScore, 100).valid, false, rawScore)
  }
})

test('rejects negative, out-of-range, and unsafe integer scores', () => {
  for (const rawScore of ['-1', '101', '9007199254740992']) {
    assert.equal(validateIntegerScore(rawScore, 100).valid, false, rawScore)
  }
})
