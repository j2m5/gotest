import { apiRequest } from '../../shared/api/client'
import type { components } from '../../shared/api/schema'
export type ForumTopic = components['schemas']['ForumTopic']
export type ForumPost = components['schemas']['ForumPost']
export type ForumTopicPage = components['schemas']['ForumTopicPage']
export type ForumTopicList = components['schemas']['ForumTopicList']
export type ForumCategoryList = components['schemas']['ForumCategoryList']
export type ModerateInput = components['schemas']['ForumModerateRequest']
const options = (id: number, signal: AbortSignal) => ({ headers: { 'X-Account-ID': id }, signal })
export const forumAPI = {
  categories: (id: number, signal: AbortSignal) => apiRequest('get', '/api/forum/categories', options(id, signal)),
  topics: (id: number, category: number, cursor: string | null, hidden: boolean, signal: AbortSignal) => apiRequest('get', '/api/forum/topics', { ...options(id, signal), query: { category_id: category, cursor: cursor ?? undefined, include_hidden: hidden || undefined } }),
  topic: (id: number, topic: number, after: number | null, signal: AbortSignal) => apiRequest('get', '/api/forum/topic', { ...options(id, signal), query: { topic_id: topic, after_id: after ?? undefined } }),
  create: (id: number, body: components['schemas']['ForumCreateRequest'], signal: AbortSignal) => apiRequest('post', '/api/forum/topics', { ...options(id, signal), body }),
  reply: (id: number, body: components['schemas']['ForumReplyRequest'], signal: AbortSignal) => apiRequest('post', '/api/forum/replies', { ...options(id, signal), body }),
  editTopic: (id: number, body: components['schemas']['ForumEditTopicRequest'], signal: AbortSignal) => apiRequest('post', '/api/forum/edit-topic', { ...options(id, signal), body }),
  editPost: (id: number, body: components['schemas']['ForumEditPostRequest'], signal: AbortSignal) => apiRequest('post', '/api/forum/edit-reply', { ...options(id, signal), body }),
  moderate: (id: number, body: ModerateInput, signal: AbortSignal) => apiRequest('post', '/api/forum/moderate', { ...options(id, signal), body }),
}
