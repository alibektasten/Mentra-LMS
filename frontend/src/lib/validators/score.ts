export type ScoreValidationResult =
  | { valid: true; score: number }
  | { valid: false; message: string }

const WHOLE_DECIMAL_PATTERN = /^\d+$/

/**
 * Validates the client-side grading contract before a score is sent to the API.
 * Exponential and fractional notation are deliberately rejected rather than
 * being coerced or truncated.
 */
export function validateIntegerScore(rawScore: string, maxPoints: number): ScoreValidationResult {
  const candidate = rawScore.trim()

  if (candidate === '') {
    return { valid: false, message: 'Enter a whole-number score.' }
  }

  if (!WHOLE_DECIMAL_PATTERN.test(candidate)) {
    return { valid: false, message: 'Score must be a whole number.' }
  }

  const score = Number(candidate)
  if (!Number.isSafeInteger(score)) {
    return { valid: false, message: 'Score must be a finite whole number.' }
  }

  if (!Number.isSafeInteger(maxPoints) || maxPoints < 0) {
    return { valid: false, message: 'This assignment has an invalid score limit.' }
  }

  if (score < 0 || score > maxPoints) {
    return { valid: false, message: `Score must be between 0 and ${maxPoints}.` }
  }

  return { valid: true, score }
}
