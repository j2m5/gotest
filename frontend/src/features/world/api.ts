import { apiRequest } from '../../shared/api/client.ts'
import type { components } from '../../shared/api/schema'

export function planetCatalog(accountId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/world/planets', { headers: { 'X-Account-ID': accountId }, signal })
}

export function worldNavigation(accountId: number, characterId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/world/navigation', { headers: { 'X-Account-ID': accountId }, query: { character_id: characterId }, signal })
}
export function travelToPlanet(accountId: number, body: components['schemas']['WorldTravelRequest'], signal?: AbortSignal) {
  return apiRequest('post', '/api/world/travel', { headers: { 'X-Account-ID': accountId }, body, signal })
}

export function planetLocations(accountId: number, characterId: number, planetId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/world/locations', { headers: { 'X-Account-ID': accountId }, query: { character_id: characterId, planet_id: planetId }, signal })
}

export function moveToLocation(accountId: number, body: components['schemas']['WorldMoveRequest'], signal?: AbortSignal) {
  return apiRequest('post', '/api/world/move', { headers: { 'X-Account-ID': accountId }, body, signal })
}
