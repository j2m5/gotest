import { apiRequest } from '../../shared/api/client.ts'
import { observeCharacter } from '../../entities/character/regeneration.ts'
import type { components } from '../../shared/api/schema'

function observeState<T extends { character: components['schemas']['Character'] }>(state: T): T {
  observeCharacter(state.character)
  return state
}

export function currentBattle(accountId: number, characterId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/battles/current', {
    headers: { 'X-Account-ID': accountId }, query: { character_id: characterId }, signal,
  }).then(observeState)
}

export function startBattle(accountId: number, body: components['schemas']['BattleStartRequest'], signal?: AbortSignal) {
  return apiRequest('post', '/api/battles/start', {
    headers: { 'X-Account-ID': accountId }, body, signal,
  }).then(observeState)
}

export function useBattleAbility(accountId: number, body: components['schemas']['BattleUseRequest'], signal?: AbortSignal) {
  return apiRequest('post', '/api/battles/use', {
    headers: { 'X-Account-ID': accountId }, body, signal,
  }).then(observeState)
}

export function retreatBattle(accountId: number, battleId: string, characterId: number, signal?: AbortSignal) {
  return apiRequest('post', '/api/battles/retreat', {
    headers: { 'X-Account-ID': accountId }, body: { battle_id: battleId, character_id: characterId }, signal,
  }).then(observeState)
}

export function battleDetail(accountId: number, battleId: string, characterId: number, afterSequence = 0, signal?: AbortSignal) {
  return apiRequest('get', '/api/battles/detail', {
    headers: { 'X-Account-ID': accountId },
    query: { battle_id: battleId, character_id: characterId, after_sequence: afterSequence }, signal,
  }).then(observeState)
}
