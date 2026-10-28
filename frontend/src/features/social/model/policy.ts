import type { components } from '../../../shared/api/schema'
export type Role = components['schemas']['AccountRole']
export const canModerate = (role: Role) => role === 'moderator' || role === 'administrator'
export const canSanction = (role: Role, actor: number, target: number, targetRole: Role = 'user') =>
  actor !== target && (role === 'administrator' || (role === 'moderator' && targetRole === 'user'))
export const messageLength = (text: string) => Array.from(text.trim()).length
export const validReason = (text: string) => messageLength(text) >= 3 && messageLength(text) <= 500
export const roleLabels: Record<Role, string> = { user: 'Игрок', moderator: 'Модератор', administrator: 'Администратор' }
export const scopeLabels = { chat: 'Чат', forum: 'Форум', account: 'Аккаунт' }
export const formatTime = (date: string) => new Date(date).toLocaleString('ru-RU')
