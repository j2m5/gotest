import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { authError, currentUser, logout } from '../../api'
import type { User } from '../../types'
import Layout from '../../components/layout/Layout'

export default function Home() {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    currentUser(controller.signal).then(({ user }) => {
      if (!controller.signal.aborted) setUser(user)
    }).catch((error: unknown) => {
      if (!controller.signal.aborted && authError(error).status !== 401) {
        setError(`Не удалось проверить сеанс. ${authError(error).message}`)
      }
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false)
    })
    return () => controller.abort()
  }, [])

  const handleLogout = async () => {
    if (pending) return
    setPending(true)
    setError('')
    try {
      await logout()
      setUser(null)
    } catch (error) {
      setError(`Не удалось выйти. ${authError(error).message}`)
    } finally {
      setPending(false)
    }
  }

  return (
    <Layout>
      <h1>Go Test</h1>

      {error && <p className="flash-error" role="alert">{error}</p>}
      {loading ? <p role="status">Проверяем сеанс…</p> : user ? <>
        <p>Добро пожаловать, {user.login}</p>
        <button type="button" className="btn-primary" disabled={pending} onClick={handleLogout}>
          {pending ? 'Выход…' : 'Выйти'}
        </button>
      </> : <>
        {!error && <p>Добро пожаловать, гость</p>}
        <p><Link to="/login">Вход</Link> · <Link to="/register">Регистрация</Link></p>
      </>}
    </Layout>
  )
}
