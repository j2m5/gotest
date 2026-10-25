import type { SubmitEvent } from 'react'
import { useState } from 'react'
import { authError, register } from '../../api'
import { Link } from 'react-router-dom'
import Layout from '../../components/layout/Layout'

export default function Register() {
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const [registered, setRegistered] = useState(false)
  const [verificationSent, setVerificationSent] = useState(false)

  const handleRegister = async(e: SubmitEvent) => {
    e.preventDefault()
    if (pending || registered) return
    setPending(true)
    setError('')
    try {
      const result = await register(email, login, password)
      setRegistered(true)
      setVerificationSent(result.verification_sent)
      setPassword('')
    } catch (error) {
      setError(authError(error).message)
    } finally {
      setPending(false)
    }
  }

  return (
    <Layout>
      <h1>Регистрация</h1>

      <p>Создайте аккаунт и подтвердите электронную почту по ссылке из письма.</p>
      {error && <p className="flash-error" role="alert">{error}</p>}
      {registered && <div className="flash-success" role="status">
        {verificationSent
          ? 'Аккаунт создан. Проверьте почту и подтвердите адрес перед входом.'
          : 'Аккаунт создан, но письмо отправить не удалось. Повторите отправку на странице входа, указав почту и пароль.'}
      </div>}
      <p className="text-center"><Link to="/login">{registered && !verificationSent ? 'Войти / повторить отправку письма' : 'Войти'}</Link></p>

      {!registered && <form onSubmit={handleRegister} aria-busy={pending}>
        <div className="text-center form-item">
          <label
            className="block text-light-orange"
            htmlFor="email">
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
        <div className="text-center form-item">
          <label
            className="block text-light-orange"
            htmlFor="login">
            Имя пользователя
          </label>
          <input
            id="login"
            type="text"
            name="login"
            autoComplete="username"
            required
            disabled={pending}
            value={login}
            placeholder="Имя пользователя"
            onChange={(e) => setLogin(e.target.value)}
          />
        </div>
        <div className="text-center form-item">
          <label
            className="block text-light-orange"
            htmlFor="password">
            Пароль
          </label>
          <input
            id="password"
            type="password"
            name="password"
            autoComplete="new-password"
            required
            disabled={pending}
            value={password}
            placeholder="Пароль"
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <button
          type="submit"
          className="btn-primary"
          disabled={pending}
        >
          {pending ? 'Регистрация…' : 'Регистрация'}
        </button>
      </form>}
    </Layout>
  )
}
