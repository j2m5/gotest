import { useState } from 'react'
import { Link } from 'react-router-dom'
import Layout from '../../shared/ui/Layout'
import { Alert, Button, ButtonLink, Checkbox, Divider, Input, Select, Textarea } from '../../shared/ui/kit'

export default function UiKit() {
  const [actions, setActions] = useState(0)
  return <Layout title="Компоненты интерфейса">
    <h1>UI-кит</h1>
    <p className="muted">Локальные образцы общих компонентов. Поля можно заполнить, а состояния — проверить мышью и клавиатурой.</p>

    <section aria-labelledby="kit-buttons">
      <h2 id="kit-buttons">Кнопки и ссылки</h2>
      <div className="social-actions">
        <Button onClick={() => setActions(value => value + 1)}>Основная</Button>
        <Button variant="secondary">Вторичная</Button>
        <Button variant="danger">Опасное действие</Button>
        <Button variant="ghost">Текстовое действие</Button>
        <Button disabled>Недоступно</Button>
        <Button pending>Сохранение…</Button>
        <Button variant="ghost" disabled>Недоступное действие</Button>
      </div>
      <p className="muted" role="status">Нажатий: {actions}</p>
      <p><Link to="/">Обычная ссылка</Link></p>
      <ButtonLink variant="secondary" to="/login">Ссылка-кнопка: вход</ButtonLink>
    </section>
    <Divider />

    <section aria-labelledby="kit-fields">
      <h2 id="kit-fields">Поля</h2>
      <form className="social-form" onSubmit={event => { event.preventDefault(); setActions(value => value + 1) }}>
        <label htmlFor="kit-name">Текст</label>
        <Input id="kit-name" name="name" placeholder="Имя персонажа" />
        <label htmlFor="kit-email">Электронная почта</label>
        <Input id="kit-email" name="email" type="email" placeholder="pilot@example.test" autoComplete="email" />
        <label htmlFor="kit-password">Пароль</label>
        <Input id="kit-password" name="password" type="password" defaultValue="example" autoComplete="new-password" />
        <label htmlFor="kit-number">Число</label>
        <Input id="kit-number" type="number" defaultValue={1} min={1} />
        <label htmlFor="kit-select">Список</label>
        <Select id="kit-select" defaultValue="republic"><option value="republic">Галактическая Республика</option><option value="empire">Империя Ситов</option></Select>
        <label htmlFor="kit-text">Сообщение</label>
        <Textarea id="kit-text" rows={3} placeholder="Текст сообщения…" />
        <label htmlFor="kit-readonly">Только чтение</label>
        <Input id="kit-readonly" readOnly value="Смотритель" />
        <label htmlFor="kit-disabled">Недоступное поле</label>
        <Input id="kit-disabled" disabled value="Недоступно" />
        <label htmlFor="kit-invalid">Поле с ошибкой</label>
        <Input id="kit-invalid" aria-invalid="true" aria-describedby="kit-field-error" defaultValue="Имя" />
        <Alert id="kit-field-error" variant="error">Пример ошибки заполнения поля.</Alert>
        <label><Checkbox defaultChecked /> Выбранный флажок</label>
        <label><Checkbox /> Обычный флажок</label>
        <label><Checkbox disabled /> Недоступный флажок</label>
        <div className="social-actions"><Button type="submit">Проверить форму</Button><Button variant="secondary" type="reset">Сбросить</Button></div>
      </form>
    </section>
    <Divider />

    <section aria-labelledby="kit-feedback">
      <h2 id="kit-feedback">Сообщения и роли</h2>
      <Alert>Информация о текущем действии.</Alert>
      <Alert variant="success">Изменения сохранены.</Alert>
      <Alert variant="warning">Действие временно недоступно.</Alert>
      <Alert variant="error">Не удалось выполнить действие. Попробуйте ещё раз.</Alert>
      <p className="social-actions"><strong className="account-nickname">Игрок</strong><strong className="account-nickname account-nickname--moderator">Модератор</strong><strong className="account-nickname account-nickname--administrator">Администратор</strong></p>
    </section>
  </Layout>
}
