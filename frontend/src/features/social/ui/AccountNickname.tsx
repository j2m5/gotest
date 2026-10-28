import type { ReactNode } from 'react'
import type { Role } from '../model/policy'

export default function AccountNickname({ role, children }: { role: Role; children: ReactNode }) {
  return <strong className={`account-nickname account-nickname--${role}`}>{children}</strong>
}
