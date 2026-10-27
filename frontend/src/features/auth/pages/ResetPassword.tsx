import type { SubmitEvent } from 'react'
import { useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { apiError } from '../../../shared/api/errors'
import type { FieldErrors } from '../../../shared/api/errors'
import FieldFeedback from '../../../shared/ui/FieldFeedback'
import { useAccount } from '../../../entities/account/context'
import Layout from '../../../shared/ui/Layout'

export default function ResetPassword() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const { store } = useAccount()
  const navigate = useNavigate()
  const [password, setPassword] = useState('')
  const [confirmation, setConfirmation] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const [fields, setFields] = useState<FieldErrors>({})
  const [complete, setComplete] = useState(false)
  const [canRequestLink, setCanRequestLink] = useState(false)

  const submit = async (event: SubmitEvent) => {
    event.preventDefault()
    if (pending || complete || !token) return
    setError('')
    setFields({})
    setCanRequestLink(false)
    if (password !== confirmation) {
      setFields({ confirmation: ['Пароли не совпадают.'] })
      return
    }
    setPending(true)
    try {
      await store.resetPassword(token, password)
      setPassword('')
      setConfirmation('')
      setComplete(true)
      navigate('/reset-password', { replace: true })
    } catch (failure) {
      const result = apiError(failure, ['password'])
      setFields(result.fields)
      const invalidToken = result.code === 'password_reset_token_invalid'
      setCanRequestLink(invalidToken)
      setError(invalidToken
        ? 'Ссылка недействительна или срок её действия истёк. Запросите новую ссылку.'
        : result.message)
    } finally {
      setPending(false)
    }
  }

  return <Layout>
    <h1>Новый пароль</h1>
    {!token && !complete && <p className="flash-error" role="alert">В ссылке нет токена. Запросите новую ссылку для сброса пароля.</p>}
    {error && <p className="flash-error" role="alert">{error}</p>}
    {complete && <p className="flash-success" role="status">Пароль изменён. Теперь можно войти.</p>}
    {!complete && token && <form onSubmit={submit} aria-busy={pending}>
      <div className="form-item text-center">
        <label className="block" htmlFor="password">Новый пароль</label>
        <input id="password" name="password" type="password" autoComplete="new-password" required disabled={pending}
          aria-invalid={Boolean(fields.password?.length)} aria-describedby={fields.password?.length ? 'password-error' : undefined}
          value={password} onChange={event => { setPassword(event.target.value); setError(''); setFields(current => ({ ...current, password: [] })) }} />
        <FieldFeedback fields={fields} name="password" id="password-error" />
      </div>
      <div className="form-item text-center">
        <label className="block" htmlFor="confirmation">Повторите новый пароль</label>
        <input id="confirmation" name="confirmation" type="password" autoComplete="new-password" required disabled={pending}
          aria-invalid={Boolean(fields.confirmation?.length)} aria-describedby={fields.confirmation?.length ? 'confirmation-error' : undefined}
          value={confirmation} onChange={event => { setConfirmation(event.target.value); setFields(current => ({ ...current, confirmation: [] })) }} />
        <FieldFeedback fields={fields} name="confirmation" id="confirmation-error" />
      </div>
      <button type="submit" className="btn-primary" disabled={pending}>{pending ? 'Сохранение…' : 'Сохранить пароль'}</button>
    </form>}
    {(complete || !token) && <p className="text-center"><Link to="/login">Перейти ко входу</Link></p>}
    {(canRequestLink || !token) && <p className="text-center"><Link to="/forgot-password">Запросить новую ссылку</Link></p>}
  </Layout>
}
