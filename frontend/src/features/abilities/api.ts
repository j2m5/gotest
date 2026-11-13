import { apiRequest } from '../../shared/api/client.ts'
import { observeCharacter } from '../../entities/character/regeneration.ts'

export function characterAbilities(accountId: number, characterId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/abilities', {
    headers: { 'X-Account-ID': accountId }, query: { character_id: characterId }, signal,
  }).then(loadout => {
    observeCharacter(loadout.character)
    return loadout
  })
}
