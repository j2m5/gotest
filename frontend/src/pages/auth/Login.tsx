import { useState } from 'react'
import { login } from '../../api'
import { Link } from 'react-router-dom'
import Layout from '../../components/layout/Layout'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleLogin = async() => {
    await login(email, password)
  }

  return (
    <Layout>
      <h1>Вход</h1>
      <form>
        <div className="form-item text-center">
          <label className="block" htmlFor="email">
            Электронная почта
          </label>
          <input
            id="email"
            type="email"
            name="email"
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
              value={password}
              placeholder="Пароль"
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <div className="form-item text-center">
            <Link to="/register">Регистрация</Link>
          </div>
          <button
            type="button"
            className="btn-primary"
            onClick={handleLogin}
          >
            Войти
          </button>
      </form>
    </Layout>
  )
}