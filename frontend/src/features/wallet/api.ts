import { apiRequest } from '../../shared/api/client'

export function walletBalances(accountId: number, characterId: number, signal?: AbortSignal) {
  return apiRequest('get', '/api/wallet', { headers: { 'X-Account-ID': accountId }, query: { character_id: characterId }, signal })
}
export function walletHistory(accountId: number, characterId: number, before?: string, signal?: AbortSignal) {
  return apiRequest('get', '/api/wallet/history', { headers: { 'X-Account-ID': accountId }, query: { character_id: characterId, ...(before ? { before } : {}) }, signal })
}
