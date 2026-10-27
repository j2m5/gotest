import { apiRequest } from '../../shared/api/client'
import type { CreateCharacterInput } from '../../entities/character/types'

export function listCharacters() {
  return apiRequest('get', '/api/characters')
}

export function characterCatalog() {
  return apiRequest('get', '/api/characters/catalog')
}

export function createCharacter(input: CreateCharacterInput) {
  return apiRequest('post', '/api/characters', { body: input })
}

export function setActiveCharacter(character_id: number | null) {
  return apiRequest('put', '/api/characters/active', { body: { character_id } })
}
