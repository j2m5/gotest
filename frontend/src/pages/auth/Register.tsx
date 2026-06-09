import { useState } from 'react'
import { register } from '../../api'
import Layout from '../../components/layout/Layout'

export default function Register() {
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')

  const handleRegister = async() => {
    await register(email, login, password)
  }

  return (
    <Layout>
      <h1>Регистрация</h1>

      <div>Lorem ipsum dolor sit amet, consectetur adipisicing elit. Beatae blanditiis consequatur corporis dicta
        dignissimos dolorem doloremque eaque enim eveniet explicabo id, illo ipsam iste laboriosam modi nulla officia
        omnis optio quaerat quam recusandae sed, sunt totam vel, vero voluptas voluptatem! Dolorem eius illum neque
        rem repellendus sed. Aliquam cum cumque error eum eveniet, incidunt laudantium libero nemo possimus quas sed
        suscipit ut veniam. A perferendis placeat, provident quae quasi qui repellendus sequi suscipit! A ab dolorem
        eos officiis, qui reiciendis rerum soluta! Consequatur cupiditate deserunt dolor doloremque dolores earum enim
        eum inventore molestiae numquam, optio, quaerat qui reiciendis vel vitae.
      </div>

      <form>
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
            value={password}
            placeholder="Пароль"
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <button
          type="button"
          className="btn-primary"
          onClick={handleRegister}
        >
          Регистрация
        </button>
      </form>
    </Layout>
  )
}