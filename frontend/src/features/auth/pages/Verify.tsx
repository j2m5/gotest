import { Alert, Button } from '../../../shared/ui/kit'
import { useEffect, useRef, useState } from 'react'
import { Link, useLocation, useNavigate, useSearchParams } from 'react-router-dom'
import { verifyEmail } from '../api'
import { apiError } from '../../../shared/api/errors'
import type { APIError } from '../../../shared/api/errors'
import Layout from '../../../shared/ui/Layout'

function VerifyToken({ token }: { token: string }) {
  const navigate = useNavigate()
  const request = useRef<ReturnType<typeof verifyEmail> | null>(null)
  const [attempt, setAttempt] = useState(0)
  const [failure, setFailure] = useState<APIError | null>(null)

  useEffect(() => {
    let active = true
    // StrictMode reconnects the effect. Reuse its request: aborting or posting
    // again could lose the successful result of a single-use token.
    request.current ??= verifyEmail(token)
    void request.current.then(() => {
      if (active) {
        // History state preserves this presentation after refresh; it never
        // authenticates an account or replaces server-side verification.
        void navigate('/verify', { replace: true, state: { emailVerified: true } })
      }
    }, error => {
      if (active) setFailure(apiError(error))
    })
    return () => { active = false }
  }, [token, navigate, attempt])

  const retry = () => {
    if (!failure) return
    request.current = null
    setFailure(null)
    setAttempt(value => value + 1)
  }

  if (!failure) return <p className="text-center" role="status">Подтверждаем почту…</p>

  const invalidToken = failure.code === 'verification_token_invalid'
  return <>
    <Alert variant="error" role="alert">{invalidToken
      ? 'Ссылка уже использована, недействительна или срок её действия истёк. Если вы уже подтверждали почту, перейдите ко входу.'
      : failure.message}</Alert>
    <p className="text-center">{invalidToken
      ? <Link to="/login?resend=1">Запросить новое письмо</Link>
      : <Button type="button" onClick={retry}>Повторить попытку</Button>}
    </p>
  </>
}

export default function Verify() {
  const [params] = useSearchParams()
  const location = useLocation()
  const token = params.get('token') || ''
  const verified = !token && location.state?.emailVerified === true

  return (
    <Layout>
      <h1>Подтверждение почты</h1>
      {token ? <VerifyToken key={token} token={token} />
        : verified ? <Alert variant="success" role="status">Почта подтверждена. Теперь можно войти.</Alert>
          : <>
            <Alert variant="error" role="alert">В ссылке нет кода подтверждения. Откройте ссылку из письма или запросите новое письмо.</Alert>
            <p className="text-center"><Link to="/login?resend=1">Запросить новое письмо</Link></p>
          </>}
      <p className="text-center"><Link to="/login">Перейти ко входу</Link></p>
    </Layout>
  )
}
