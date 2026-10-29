import type { ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'
import type { Role } from '../model/policy'

export default function AccountNickname({ role, children, characterId }: { role: Role; children: ReactNode; characterId?: number | null }) {
  const location = useLocation()
  const nickname = <strong className={`account-nickname account-nickname--${role}`}>{children}</strong>
  return characterId ? <Link to={`/characters/${characterId}`} state={{ profileOrigin: location.pathname }}>{nickname}</Link> : nickname
}
