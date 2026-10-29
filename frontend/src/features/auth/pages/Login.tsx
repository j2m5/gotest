import { Alert, Button, Checkbox, Input } from '../../../shared/ui/kit'
import type { SubmitEvent } from 'react'
import { useState } from 'react'
import { resendVerification } from '../api'
import { apiError } from '../../../shared/api/errors'
import { useAccount } from '../../../entities/account/context'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
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
  const [params] = useSearchParams()
  const [resendOpen, setResendOpen] = useState<boolean | null>(null)
  const showResend = resendOpen ?? (unverified || params.get('resend') === '1')
  const navigate = useNavigate()

  const handleLogin = async(e: SubmitEvent) => {
    e.preventDefault()
    if (pending) return
    setPending(true)
    setError('')
    setMessage('')
    setUnverified(false)
    setResendOpen(null)
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
      setMessage('Проверьте почту: отправляем письмо с новой ссылкой подтверждения.')
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
      {error && <Alert variant="error" role="alert">{error}</Alert>}
      {message && <Alert variant="success" role="status">{message}</Alert>}
      <form onSubmit={handleLogin} aria-busy={pending}>
        <div className="form-item text-center">
          <label className="block" htmlFor="email">
            Электронная почта
          </label>
          <Input
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
            <Input
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
              <Checkbox
                id="remember-me"
                name="remember_me"
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
          <Button pending={pending}
            type="submit"
            disabled={pending}
          >
            {pending ? 'Подождите…' : 'Войти'}
          </Button>
          <div className="form-item text-center">
            {unverified && <p>Подтвердите почту, чтобы войти.</p>}
            <Button variant="ghost" pending={pending} type="button" className="sm-text" disabled={pending}
              aria-expanded={showResend} aria-controls="resend-help" onClick={() => setResendOpen(!showResend)}>
              Письмо не пришло?
            </Button>
            {showResend && <div id="resend-help">
              <p className="sm-text">Проверьте папку «Спам». Для нового письма введите почту и пароль в форме выше.</p>
              <Button variant="ghost" pending={pending} id="resend-verification" type="button" disabled={pending} onClick={handleResend}>
                Отправить письмо повторно
              </Button>
            </div>}
          </div>
      </form>
    </Layout>
  )
}
