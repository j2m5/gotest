import { apiRequest } from '../../shared/api/client'
const options = (id: number, signal: AbortSignal) => ({ headers: { 'X-Account-ID': id }, signal })
export const presenceAPI = {
  heartbeat: (id: number, character: number | null, signal: AbortSignal) => apiRequest('post', '/api/presence/heartbeat', { ...options(id, signal), body: { character_id: character } }),
  statuses: (id: number, ids: number[], signal: AbortSignal) => apiRequest('get', '/api/presence', { ...options(id, signal), query: { character_ids: ids.length ? ids.join(',') : undefined } }),
  online: (id: number, after: number | null, signal: AbortSignal) => apiRequest('get', '/api/presence/online', { ...options(id, signal), query: { after_id: after ?? undefined } }),
}
