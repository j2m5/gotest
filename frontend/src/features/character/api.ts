import { apiRequest } from '../../shared/api/client.ts'
import type { CreateCharacterInput } from '../../entities/character/types'

export function listCharacters(accountId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/characters', { headers: { 'X-Account-ID': accountId }, signal })
}

export function characterCatalog(accountId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/characters/catalog', { headers: { 'X-Account-ID': accountId }, signal })
}

export function createCharacter(accountId: number, input: CreateCharacterInput, signal?: AbortSignal) {
  return apiRequest('post', '/api/characters', { headers: { 'X-Account-ID': accountId }, body: input, signal })
}

export function setActiveCharacter(accountId: number, character_id: number | null, signal?: AbortSignal) {
  return apiRequest('put', '/api/characters/active', { headers: { 'X-Account-ID': accountId }, body: { character_id }, signal })
}
