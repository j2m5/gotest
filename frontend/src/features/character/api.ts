import { apiRequest } from '../../shared/api/client.ts'
import type { CreateCharacterInput } from '../../entities/character/types'
import { observeCharacter } from '../../entities/character/regeneration.ts'

export function listCharacters(accountId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/characters', { headers: { 'X-Account-ID': accountId }, signal }).then(response => {
    const receivedAt = performance.now()
    response.characters.forEach(character => observeCharacter(character, receivedAt))
    return response
  })
}

export function publicCharacterProfile(accountId: number, characterId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/characters/profile', { headers: { 'X-Account-ID': accountId }, query: { character_id: characterId }, signal })
}

export function characterCatalog(accountId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/characters/catalog', { headers: { 'X-Account-ID': accountId }, signal })
}

export function createCharacter(accountId: number, input: CreateCharacterInput, signal?: AbortSignal) {
  return apiRequest('post', '/api/characters', { headers: { 'X-Account-ID': accountId }, body: input, signal }).then(response => {
    observeCharacter(response.character)
    return response
  })
}

export function setActiveCharacter(accountId: number, character_id: number | null, signal?: AbortSignal) {
  return apiRequest('put', '/api/characters/active', { headers: { 'X-Account-ID': accountId }, body: { character_id }, signal }).then(response => {
    if (response.character) observeCharacter(response.character)
    return response
  })
}
