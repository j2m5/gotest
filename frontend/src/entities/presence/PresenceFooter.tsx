import { Link } from 'react-router-dom'
import { useAccount } from '../account/context'
import { usePresence } from './context'

export default function PresenceFooter() {
  const { status } = useAccount()
  const { total } = usePresence()
  return status === 'ready' ? <p className="presence-footer"><Link to="/online">Кто онлайн: {total ?? '—'}</Link></p> : null
}
