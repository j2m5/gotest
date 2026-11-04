import type { components } from '../../../shared/api/schema'

export default function NPCKindIcon({ kind }: { kind: components['schemas']['NPCInstance']['kind'] }) {
  return <svg className="npc-icon" viewBox="0 0 48 48" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    {kind === 'humanoid' ? <>
      <path d="M12 23v-5a12 12 0 0 1 24 0v5l-3 12-9 5-9-5zM12 23h24M17 23v5l7 4 7-4v-5M24 7v9M20 36h8" />
      <path d="M16 18h16" />
    </> : kind === 'droid' ? <>
      <rect x="11" y="13" width="26" height="24" rx="5" />
      <path d="M24 7v6M8 21v9M40 21v9M19 32h10" />
      <circle cx="18" cy="23" r="3" /><circle cx="30" cy="23" r="3" />
    </> : <>
      <path d="m10 10 10 7h8l10-7-2 18-12 12-12-12zM13 21l6 4M35 21l-6 4M20 31l4 3 4-3M24 34v6" />
      <path d="m16 30 2 6M32 30l-2 6" />
    </>}
  </svg>
}
