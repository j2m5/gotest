# UI kit

Общие нативные React-компоненты для всех экранов. Образцы доступны при `npm run dev`
на `/_ui-kit`; в production маршрут и код галереи не включаются в сборку.

```tsx
import { Alert, Button, ButtonLink, Checkbox, Divider, Input, Select, Textarea } from './shared/ui/kit'

<form onSubmit={save}>
  <label htmlFor="name">Имя</label>
  <Input id="name" value={name} onChange={event => setName(event.target.value)}
    disabled={pending} aria-invalid={!!error} aria-describedby={error ? 'name-error' : undefined} />
  {error && <Alert variant="error" id="name-error">{error}</Alert>}
  <Button type="submit" pending={pending}>{pending ? 'Сохранение…' : 'Сохранить'}</Button>
  <Button variant="secondary" onClick={cancel}>Отмена</Button>
</form>
```

- `Button`: `primary` (по умолчанию), `secondary`, `danger`, `ghost`.
  По умолчанию `type="button"`; для отправки формы **явно** задавать `type="submit"`.
  `pending` блокирует действие и выставляет `aria-busy`; текст ожидания задаёт экран.
  `disabled` — недоступность по другой причине. Компонент не делает запросов.
- `ButtonLink`: те же варианты, но настоящий Router Link с `to`, а не кнопка.
  Обычные текстовые ссылки — `Link`/`NavLink` или `a`; общий стиль убирает underline.
- `Input`, `Select`, `Textarea`, `Checkbox` передают нативные props, события и `ref`.
  Для флажков использовать `Checkbox`; у `Input` по умолчанию `type="text"`.
  Label, id, подсказки, валидацию и управление состоянием задаёт вызывающий экран.
- `Alert`: `info`, `success`, `warning`, `error`. Ошибка имеет `role="alert"`,
  остальные варианты — `role="status"`. Роль можно переопределить.
- `Divider`: семантический `hr`; рамки разделов используют тот же `--divider`.

`tokens.css` — цвета, интервалы, скругление и переходы; `kit.css` — состояния
компонентов. Оба файла подключены из `styles/index.css`, где остаётся компоновка
экранов. Новые страницы используют компоненты; локальные правила не переопределяют
их цвета, фокус и hover. `className` нужен для расположения, а не отдельной темы.

Тёмный фон полей одинаков для всех текстовых типов, списков и многострочного ввода.
Фокус поля выделяется рамкой и мягкой подсветкой, ссылки/кнопки имеют явный Tab-фокус.
Учтены reduced-motion, forced-colors и autofill. Нативные controls сохраняют работу
клавиатуры и браузерную валидацию. Ширина приложения — до 480 px, без внешних шрифтов.

Поведенческие и визуальные проверки находятся в `tests/frontend/*.browser.mjs`
и запускаются Go-сценариями из `tests/db`. Общие проверки оформления —
`tests/frontend/ui-kit.browser-assertions.mjs`.
