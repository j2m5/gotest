import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { verifyEmail } from '../api'
import { apiError } from '../../../shared/api/errors'
import Layout from '../../../shared/ui/Layout'

export default function Verify() {
  const [params, setParams] = useSearchParams()
  const token = params.get('token') || ''
  const [pending, setPending] = useState(false)
  const [verified, setVerified] = useState(false)
  const [error, setError] = useState('')

  // Confirmation is deliberately an explicit action: GET and StrictMode never consume a token.
  const handleVerify = async () => {
    if (pending || !token || verified) return
    setPending(true)
    setError('')
    try {
      await verifyEmail(token)
      setVerified(true)
      setParams({}, { replace: true })
    } catch (error) {
      setError(apiError(error).message)
    } finally {
      setPending(false)
    }
  }

  return (
    <Layout>
      <h1>Подтверждение почты</h1>
      {error && <p className="flash-error" role="alert">{error}</p>}
      {verified ? <p className="flash-success" role="status">Почта подтверждена. Теперь можно войти.</p>
        : token ? <div className="text-center">
          <p>Нажмите кнопку, чтобы подтвердить электронную почту.</p>
          <button type="button" className="btn-primary" disabled={pending} onClick={handleVerify}>
            {pending ? 'Подтверждение…' : 'Подтвердить почту'}
          </button>
        </div> : <p className="flash-error" role="alert">В ссылке нет кода подтверждения. Откройте ссылку из письма или запросите новое письмо на странице входа.</p>}
      <p className="text-center"><Link to="/login">Войти / повторить отправку письма</Link></p>
    </Layout>
  )
}
