import { useEffect } from 'react'
import { usePresence } from './context'

export default function PresenceIndicator({ characterId }: { characterId?: number | null }) {
  const { store, statuses } = usePresence()
  useEffect(() => characterId ? store?.watch([characterId]) : undefined, [store, characterId])
  if (!characterId) return null
  const status = statuses[characterId]
  const label = status === true ? 'В игре' : status === false ? 'Не в игре' : 'Статус неизвестен'
  return <span className={`presence-indicator presence-indicator--${status === true ? 'online' : status === false ? 'offline' : 'unknown'}`} role="status">{label}</span>
}
