import { apiRequest } from '../../shared/api/client'
import type { components } from '../../shared/api/schema'
export type ChatMessage = components['schemas']['ChatMessage']
export type ChatPage = components['schemas']['ChatPage']
export type UserBan = components['schemas']['UserBan']
export type BanPage = components['schemas']['BanPage']
export type ActionPage = components['schemas']['ActionPage']
export type BanInput = components['schemas']['BanRequest']
export type ModeratorList = components['schemas']['ModeratorList']
const options = (id: number, signal: AbortSignal) => ({ headers: { 'X-Account-ID': id }, signal })
export const socialAPI = {
  moderators: (id: number, signal: AbortSignal) => apiRequest('get', '/api/chat/moderators', options(id, signal)),
  messages: (id: number, before: number | null, signal: AbortSignal) => apiRequest('get', '/api/chat/messages', { ...options(id, signal), query: { before_id: before ?? undefined } }),
  send: (id: number, character: number, body: string, signal: AbortSignal) => apiRequest('post', '/api/chat/messages', { ...options(id, signal), body: { character_id: character, body } }),
  delete: (id: number, message: number, reason: string, signal: AbortSignal) => apiRequest('post', '/api/chat/delete', { ...options(id, signal), body: { message_id: message, reason } }),
  bans: (id: number, before: number | null, user: number | undefined, signal: AbortSignal) => apiRequest('get', '/api/moderation/bans', { ...options(id, signal), query: { before_id: before ?? undefined, user_id: user } }),
  ban: (id: number, body: BanInput, signal: AbortSignal) => apiRequest('post', '/api/moderation/bans', { ...options(id, signal), body }),
  revoke: (id: number, ban: number, reason: string, signal: AbortSignal) => apiRequest('post', '/api/moderation/revoke', { ...options(id, signal), body: { ban_id: ban, reason } }),
  actions: (id: number, before: number | null, signal: AbortSignal) => apiRequest('get', '/api/moderation/actions', { ...options(id, signal), query: { before_id: before ?? undefined } }),
}
