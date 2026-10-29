import { Navigate } from 'react-router-dom'
import { useAccount } from '../../entities/account/context'
import AccountGate from '../../entities/account/AccountGate'
import Layout from '../../shared/ui/Layout'
import { ButtonLink } from '../../shared/ui/kit'

export default function Home() {
  const { status, user } = useAccount()
  if (status === 'anonymous') return <Layout>
    <section className="welcome">
      <h1 className="visually-hidden">Star Wars: The Old Republic</h1>
      <img className="welcome-image" src="/images/scenes/icewar.jpg" width="340" height="183" alt="Джедаи и ситхи в сражении" />
      <p>После подписания Корускантского договора наступил шаткий мир между Империей Ситов и Галактической Республикой. Обе стороны готовятся к новой схватке…</p>
      <p>Пока Республика заживляет раны, полученные во время Великой войны, Империя укрепляет свою мощь.</p>
      <p>Тебе предстоит выбрать сторону.</p>
      <nav className="welcome-actions" aria-label="Вход в игру">
        <ButtonLink to="/register">Зарегистрироваться</ButtonLink>
        <ButtonLink to="/login" variant="secondary">Войти</ButtonLink>
      </nav>
    </section>
  </Layout>
  return <AccountGate><Navigate to={user?.character_id ? '/game' : '/characters'} replace /></AccountGate>
}
