import type { SubmitEvent } from 'react'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { forgotPassword } from '../api'
import { apiError } from '../../../shared/api/errors'
import type { FieldErrors } from '../../../shared/api/errors'
import FieldFeedback from '../../../shared/ui/FieldFeedback'
import Layout from '../../../shared/ui/Layout'

const genericMessage = 'Если аккаунт с такой почтой существует, мы отправим ссылку для сброса пароля. Повторный запрос возможен через минуту.'

export default function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState('')
  const [fields, setFields] = useState<FieldErrors>({})
  const [sent, setSent] = useState(false)

  const submit = async (event: SubmitEvent) => {
    event.preventDefault()
    if (pending || sent) return
    setPending(true)
    setError('')
    setFields({})
    try {
      await forgotPassword(email)
      setSent(true)
    } catch (failure) {
      const result = apiError(failure, ['email'])
      setError(result.message)
      setFields(result.fields)
    } finally {
      setPending(false)
    }
  }

  return <Layout>
    <h1>Сброс пароля</h1>
    <p>Укажите почту, связанную с аккаунтом.</p>
    {error && <p className="flash-error" role="alert">{error}</p>}
    {sent && <p className="flash-success" role="status">{genericMessage}</p>}
    {!sent && <form onSubmit={submit} aria-busy={pending}>
      <div className="form-item text-center">
        <label className="block" htmlFor="email">Электронная почта</label>
        <input id="email" name="email" type="email" autoComplete="email" required disabled={pending}
          aria-invalid={Boolean(fields.email?.length)} aria-describedby={fields.email?.length ? 'email-error' : undefined}
          value={email} onChange={event => { setEmail(event.target.value); setError(''); setFields(current => ({ ...current, email: [] })) }} />
        <FieldFeedback fields={fields} name="email" id="email-error" />
      </div>
      <button type="submit" className="btn-primary" disabled={pending}>{pending ? 'Отправка…' : 'Отправить ссылку'}</button>
    </form>}
    <p className="text-center"><Link to="/login">Вернуться ко входу</Link></p>
  </Layout>
}
