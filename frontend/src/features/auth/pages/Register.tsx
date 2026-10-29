import { Alert, Button, Input } from '../../../shared/ui/kit'
import type { SubmitEvent } from 'react'
import { useState } from 'react'
import { register } from '../api'
import { apiError } from '../../../shared/api/errors'
import { Link } from 'react-router-dom'
import type { FieldErrors } from '../../../shared/api/errors'
import FieldFeedback from '../../../shared/ui/FieldFeedback'
import Layout from '../../../shared/ui/Layout'

export default function Register() {
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const [fields, setFields] = useState<FieldErrors>({})
  const clearField = (name: string) => {
    setError('')
    setFields(current => Object.fromEntries(Object.entries(current).filter(([key]) => key !== name)))
  }
  const [registered, setRegistered] = useState(false)
  const [verificationQueued, setVerificationQueued] = useState(false)

  const handleRegister = async(e: SubmitEvent) => {
    e.preventDefault()
    if (pending || registered) return
    setPending(true)
    setError('')
    setFields({})
    try {
      const result = await register(email, login, password)
      setRegistered(true)
      setVerificationQueued(result.verification_queued)
      setPassword('')
    } catch (error) {
      const failure = apiError(error, ['email', 'login', 'password'])
      setError(failure.message)
      setFields(failure.fields)
    } finally {
      setPending(false)
    }
  }

  return (
    <Layout>
      <img className="scene-image" src="/images/scenes/register.jpg" width="550" height="120" alt="" />
      <h1>Регистрация</h1>

      <p>Создайте аккаунт и подтвердите электронную почту по ссылке из письма.</p>
      {error && <Alert variant="error" role="alert">{error}</Alert>}
      {registered && <Alert variant="success" role="status">
        {verificationQueued
          ? 'Аккаунт создан. Проверьте почту и перейдите по ссылке в письме, чтобы подтвердить адрес.'
          : 'Аккаунт создан, но отправить письмо не удалось. Запросите его повторно.'}
      </Alert>}
      <p className="text-center"><Link to="/login">Войти</Link></p>
      {registered && <p className="text-center sm-text"><Link to="/login?resend=1">Письмо не пришло?</Link></p>}

      {!registered && <form onSubmit={handleRegister} aria-busy={pending}>
        <div className="text-center form-item">
          <label
            className="block text-light-orange"
            htmlFor="email">
            Электронная почта
          </label>
          <Input
            id="email"
            type="email"
            name="email"
            autoComplete="email"
            required
            disabled={pending}
            aria-invalid={Boolean(fields.email?.length)}
            aria-describedby={fields.email?.length ? "email-error" : undefined}
            value={email}
            placeholder="Электронная почта"
            onChange={(e) => { setEmail(e.target.value); clearField("email") }}
          />
          <FieldFeedback fields={fields} name="email" id="email-error" />
        </div>
        <div className="text-center form-item">
          <label
            className="block text-light-orange"
            htmlFor="login">
            Имя пользователя
          </label>
          <Input
            id="login"
            type="text"
            name="login"
            autoComplete="username"
            required
            disabled={pending}
            aria-invalid={Boolean(fields.login?.length)}
            aria-describedby={fields.login?.length ? "login-error" : undefined}
            value={login}
            placeholder="Имя пользователя"
            onChange={(e) => { setLogin(e.target.value); clearField("login") }}
          />
          <FieldFeedback fields={fields} name="login" id="login-error" />
        </div>
        <div className="text-center form-item">
          <label
            className="block text-light-orange"
            htmlFor="password">
            Пароль
          </label>
          <Input
            id="password"
            type="password"
            name="password"
            autoComplete="new-password"
            required
            disabled={pending}
            aria-invalid={Boolean(fields.password?.length)}
            aria-describedby={fields.password?.length ? "password-error" : undefined}
            value={password}
            placeholder="Пароль"
            onChange={(e) => { setPassword(e.target.value); clearField("password") }}
          />
          <FieldFeedback fields={fields} name="password" id="password-error" />
        </div>
        <Button pending={pending}
          type="submit"
          disabled={pending}
        >
          {pending ? 'Регистрация…' : 'Регистрация'}
        </Button>
      </form>}
    </Layout>
  )
}
