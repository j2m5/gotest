import { Fragment, type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAccount } from './context'
import Layout from '../../shared/ui/Layout'
export default function AccountGate({ children }: { children: ReactNode }) {
  const { status, error, store, user } = useAccount()
  if (status === 'anonymous') return <Navigate to="/login" replace />
  if (status === 'loading') return <Layout><p role="status">Проверяем сеанс…</p></Layout>
  if (status === 'error') return <Layout><p className="flash-error" role="alert">Не удалось обновить данные. {error}</p><button className="btn-primary" onClick={() => void store.refresh()}>Повторить</button></Layout>
  // A background refresh can replace the account without entering loading.
  // Drafts and other page-local state belong to the account that opened them.
  return <Fragment key={user?.id}>{children}</Fragment>
}
