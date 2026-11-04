import { apiRequest } from '../../shared/api/client'

export function locationNPCs(accountId: number, characterId: number, locationId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/npcs', { headers: { 'X-Account-ID': accountId }, query: { character_id: characterId, location_id: locationId }, signal })
}
