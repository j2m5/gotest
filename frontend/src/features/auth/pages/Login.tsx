import type { SubmitEvent } from 'react'
import { useState } from 'react'
import { resendVerification } from '../api'
import { apiError } from '../../../shared/api/errors'
import { useAccount } from '../../../entities/account/context'
import { Link, useNavigate } from 'react-router-dom'
import { needsEmailVerification } from '../errors'
import Layout from '../../../shared/ui/Layout'

export default function Login() {
  const { store } = useAccount()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(false)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [unverified, setUnverified] = useState(false)
  const navigate = useNavigate()

  const handleLogin = async(e: SubmitEvent) => {
    e.preventDefault()
    if (pending) return
    setPending(true)
    setError('')
    setMessage('')
    setUnverified(false)
    try {
      await store.signIn(email, password, rememberMe)
      setPassword('')
      navigate('/', { replace: true })
    } catch (error) {
      const failure = apiError(error)
      setError(failure.message)
      setUnverified(needsEmailVerification(failure))
    } finally {
      setPending(false)
    }
  }

  const handleResend = async () => {
    if (pending) return
    if (!email || !password) {
      setError('Для повторной отправки введите электронную почту и пароль.')
      return
    }
    setPending(true)
    setError('')
    setMessage('')
    try {
      await resendVerification(email, password)
      setMessage('Письмо подтверждения поставлено в очередь. Когда оно придёт, перейдите по ссылке в письме.')
      setPassword('')
    } catch (error) {
      setError(apiError(error).message)
    } finally {
      setPending(false)
    }
  }

  return (
    <Layout>
      <img className="scene-image" src="/images/scenes/auth.jpg" width="550" height="120" alt="" />
      <h1>Вход</h1>
      {error && <p className="flash-error" role="alert">{error}</p>}
      {message && <p className="flash-success" role="status">{message}</p>}
      <form onSubmit={handleLogin} aria-busy={pending}>
        <div className="form-item text-center">
          <label className="block" htmlFor="email">
            Электронная почта
          </label>
          <input
            id="email"
            type="email"
            name="email"
            autoComplete="email"
            required
            disabled={pending}
            value={email}
            placeholder="Электронная почта"
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
          <div className="form-item text-center">
            <label className="block" htmlFor="password">
              Пароль
            </label>
            <input
              id="password"
              type="password"
              name="password"
              autoComplete="current-password"
              required
              disabled={pending}
              value={password}
              placeholder="Пароль"
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <div className="form-item text-center">
            <label htmlFor="remember-me">
              <input
                id="remember-me"
                name="remember_me"
                type="checkbox"
                checked={rememberMe}
                disabled={pending}
                onChange={(e) => setRememberMe(e.target.checked)}
              />{' '}
              Запомнить меня
            </label>
          </div>
          <div className="form-item text-center">
            <Link to="/forgot-password">Забыли пароль?</Link>
          </div>
          <div className="form-item text-center">
            <Link to="/register">Регистрация</Link>
          </div>
          <button
            type="submit"
            className="btn-primary"
            disabled={pending}
          >
            {pending ? 'Подождите…' : 'Войти'}
          </button>
          <div className="form-item text-center">
            {unverified && <p>Подтвердите почту, чтобы войти.</p>}
            <button type="button" className="btn-primary" disabled={pending} onClick={handleResend}>
              Отправить письмо подтверждения повторно
            </button>
            <p className="sm-text">Для отправки нужны электронная почта и пароль.</p>
          </div>
      </form>
    </Layout>
  )
}
