import { Link, Navigate } from 'react-router-dom'
import { useAccount } from '../../entities/account/context'
import AccountGate from '../../entities/account/AccountGate'
import Layout from '../../shared/ui/Layout'

export default function Home() {
  const { status, user } = useAccount()
  if (status === 'anonymous') return <Layout><h1>Go Test</h1><p>Добро пожаловать, гость</p><p><Link to="/login">Вход</Link> · <Link to="/register">Регистрация</Link></p></Layout>
  return <AccountGate><Navigate to={user?.character_id ? '/game' : '/characters'} replace /></AccountGate>
}
