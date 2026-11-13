import type { Character } from './types'

const receivedAt = new WeakMap<Character, number>()

/** Always derive from the original baseline so partial points are never lost. */
export function projectResource(maximum: number, baseline: number, elapsedMs: number, periodMs: number, remainder = 0): number {
  const max = Math.max(0, maximum)
  if (max === 0) return 0
  const current = Math.max(0, Math.min(max, baseline))
  if (!Number.isFinite(elapsedMs) || !Number.isFinite(periodMs) || periodMs <= 0) return current
  const elapsed = Math.max(0, Math.min(elapsedMs, periodMs))
  const carry = Number.isFinite(remainder) && remainder >= 0 && remainder < periodMs ? remainder : 0
  return Math.min(max, current + Math.floor((max * elapsed + carry) / periodMs))
}

/** Elapsed time is explicit, permitting deterministic projection without a clock. */
export function projectCharacterResources(character: Character, elapsedSinceReceiptMs: number) {
  const regeneration = character.regeneration
  if (!regeneration) return {
    hp: Math.max(0, Math.min(character.hp, character.hp_current)),
    mp: Math.max(0, Math.min(character.mp, character.mp_current)),
  }
  const observedAt = Date.parse(regeneration.observed_at)
  const elapsed = Math.max(0, elapsedSinceReceiptMs)
  return {
    hp: projectResource(character.hp, regeneration.hp_baseline, observedAt - Date.parse(regeneration.hp_updated_at) + elapsed, regeneration.hp_full_restore_ms, regeneration.hp_remainder),
    mp: projectResource(character.mp, regeneration.mp_baseline, observedAt - Date.parse(regeneration.mp_updated_at) + elapsed, regeneration.mp_full_restore_ms, regeneration.mp_remainder),
  }
}

/** Called at API resolution, before the snapshot is published to the account store. */
export function observeCharacter(character: Character, now = performance.now()): void {
  if (!receivedAt.has(character)) receivedAt.set(character, now)
}

export function characterResources(character: Character, now = performance.now()) {
  const receipt = receivedAt.get(character)
  return projectCharacterResources(character, receipt === undefined ? 0 : Math.max(0, now - receipt))
}
