import type { User } from './types'

export const roleLabels: Record<User['role'], string> = { user: 'Игрок', moderator: 'Модератор', administrator: 'Администратор' }
