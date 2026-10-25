import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAccount } from '../account/context'
import Layout from './layout/Layout'
export default function AccountGate({ children }: { children: ReactNode }) {
  const { status, error, store } = useAccount()
  if (status === 'anonymous') return <Navigate to="/login" replace />
  if (status === 'loading') return <Layout><p role="status">Проверяем сеанс…</p></Layout>
  if (status === 'error') return <Layout><p className="flash-error" role="alert">Не удалось обновить данные. {error}</p><button className="btn-primary" onClick={() => void store.refresh()}>Повторить</button></Layout>
  return children
}
