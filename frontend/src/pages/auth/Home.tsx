import { useState } from 'react'
import Layout from '../../components/layout/Layout'

interface AuthUser {
  login: string
}

export default function Home() {
  // заглушка имитирующая пользователя в системе
  const [user, _] = useState<AuthUser | null>(null)

  return (
    <Layout>
      <h1>Go Test</h1>

      {user ? <p>Welcome, {user.login}</p> : <p>Welcome, guest</p>}
    </Layout>
  )
}