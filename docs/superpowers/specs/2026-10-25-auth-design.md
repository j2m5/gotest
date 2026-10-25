# Авторизация: архитектура и поведение

Пользователь поручил исправить авторизацию и переработать её архитектуру после аудита.
Дополнительно явно выбраны два почтовых драйвера: local и smtp.
Работа опирается на текущую ветку (исходный HEAD b959a77); существующие React-формы сохраняются.

## Цель
Рабочая цепочка: регистрация → письмо → подтверждение → вход → восстановление сессии → выход.
Ошибки не раскрывают детали БД. Сценарии проверяются автоматическими тестами.
Новые зависимости не требуются: Go 1.26, pgx, bcrypt, React, Axios уже присутствуют.

## Решение
Серверные сессии в PostgreSQL подходят существующему приложению. JWT добавил бы отдельную
задачу отзыва токенов; сохранение бизнес-логики в обработчиках оставило бы нынешнюю связанность.
Выбраны небольшой сервис auth, его интерфейс Repository, SQL-адаптер db, HTTP-адаптеры,
отдельный mailer с драйверами. auth не импортирует HTTP или db.

## Контракты
auth.Repository:
- CreateRegistration(ctx, email, login, passwordHash, tokenHash string, expiresAt time.Time) error
- FindUserByEmail(ctx, email string) (*models.User, error)
- CreateSession(ctx, userID int, tokenHash string, expiresAt time.Time) error
- FindSession(ctx, tokenHash string) (*auth.Session, error)
- DeleteSession(ctx, tokenHash string) error
- ConsumeVerification(ctx, tokenHash string, now time.Time) error
- ReplaceVerification(ctx, userID int, tokenHash string, expiresAt time.Time) error

В подписях ctx означает context.Context; отдельные группы аргументов записаны сокращённо.
Session содержит User *models.User и ExpiresAt time.Time.
auth.VerificationSender.SendVerification(ctx context.Context, email, token string) error.
auth.Options содержит SessionTTL, VerificationTTL time.Duration.
auth.NewService(repo Repository, sender VerificationSender, options Options) *Service.
Service: Register(ctx, RegisterInput) error; Login(ctx, email, password string) (*LoginResult,error);
CurrentUser(ctx, token string) (*models.User,error); Logout(ctx, token string) error;
VerifyEmail(ctx, token string) error; ResendVerification(ctx, email,password string) error.
RegisterInput содержит Email, Login, Password string.
LoginResult содержит User *models.User, Token string, ExpiresAt time.Time.
Сигнальные ошибки: ErrNotFound, ErrConflict, ErrInvalidCredentials, ErrUnverified,
ErrInvalidToken, ErrUnauthenticated, ErrAlreadyVerified, ErrDelivery.
ValidationError.Messages — []string.

## Данные и безопасность
Email trim + lowercase, не более 254 байт, без display name; login trim, 1–32 Unicode-символа,
без управляющих символов. Новый пароль 8–72 байта; вход сохраняет совместимость со старыми
короткими паролями. bcrypt. Случайные токены 32 байта, hex; в БД только SHA-256 токенов.
Сессия по умолчанию 168h, подтверждение 24h. Истёкшая сессия не даёт доступа.
Подтверждение одноразовое, без сессии. GET страницы подтверждения ничего не меняет;
действие выполняется отдельным POST после нажатия кнопки.
Logout идемпотентен, отзывает серверную сессию; сбой БД не выдаётся за успешный отзыв.
HTTP cookie HttpOnly, SameSite=Lax, Secure при HTTPS, явный срок действия.
Небезопасные cross-origin запросы блокируются; auth-запросы ограничиваются по IP.
Тело JSON ограничено 16 KiB, допускается ровно один объект.

## Атомарность и доставка
Создание пользователя + токена и использование подтверждения — отдельные транзакции.
Повторная отправка требует email и пароль, заменяет старый токен под блокировкой пользователя.
ConsumeVerification и ReplaceVerification используют один порядок блокировок: сначала users.
Почта отправляется после commit: SMTP нельзя включить в транзакцию PostgreSQL.
При сбое почты аккаунт остаётся созданным: регистрация отвечает 201 и verification_sent=false,
понятным сообщением и возможностью повторной отправки. Resend при сбое почты отвечает 503.
Outbox/фоновый worker в этот этап не входят; явная повторная отправка обеспечивает восстановление.

## Почта
mailer.Options: Driver, From, PublicURL, Directory, SMTPAddress, SMTPUsername,
SMTPPassword, SMTPTLSMode string; Timeout time.Duration.
mailer.New(Options) (*Mailer,error), Mailer реализует VerificationSender.
local создаёт отдельный .eml в .local/mail, директория исключена из Git.
smtp поддерживает обязательный STARTTLS или implicit TLS с проверкой сертификата,
контекст/таймаут, передачу ошибок. Настоящие письма во время тестов не отправляются.
Оба драйвера получают одинаковое письмо; ссылка PublicURL + /verify?token=....

## HTTP и React
POST /api/auth/register: email, login, password → 201 {message, verification_sent}.
POST /api/auth/login → 200 {user:{id,email,login}} и cookie.
POST /api/auth/logout → 200 {message}, удаляет cookie.
POST /api/auth/verify: {token}; совместимо также с прежним query token → 200 {message}.
POST /api/auth/resend-verification: {email,password} → 200 {message}.
GET /api/auth/me → 200 {user} либо 401.
Ошибки: 400 JSON/токен; 401 credentials/session; 403 unverified; 409 conflict/already verified;
413 body too large; 415 content type; 422 validation; 429 rate limit; 500 internal; 503 delivery.
Общий формат {error, errors?:string[]}. Неизвестный API — JSON 404, неверный метод — JSON 405.
Каждый экземпляр роутера имеет свой ServeMux. Файлы SPA раздаются отдельно от fallback;
неизвестный asset — 404. React показывает результаты/ошибки, не сохраняет токен в JS,
восстанавливает пользователя через /me, поддерживает повторную отправку и страницу /verify.

## Миграция и границы
Новая миграция расширяет email, добавляет нормализованный уникальный индекс, session expires_at,
переводит verification expiry в timestamptz, хеширует существующие токены.
Дубли email после нормализации останавливают миграцию, не удаляются автоматически.
Up сохраняет существующие аккаунты и токены (старые сессии ограничиваются created_at + 7 дней).
Down не может восстановить исходные токены: удаляет сессии/подтверждения, не пользователей.
Миграции не применяются к рабочей БД автоматически. Seeds и игровые модели вне задачи.
Сервер получает timeout и graceful shutdown; ошибки конфигурации возвращаются в main.

## Проверки
Service unit tests: полный цикл, bcrypt/хеши токенов, expiry, повторный токен, валидация,
конфликт, сбой БД/почты, resend, отменённый context.
HTTP tests: cookie, статусы/JSON, методы, лимит тела, cross-origin, API 404, assets.
Mailer tests: отдельные файлы, корректный MIME, инъекции заголовков, отмена/таймаут,
SMTP TLS/STARTTLS с локальным тестовым сервером.
SQL integration tests используют только явно заданный TEST_DATABASE_URL и временную схему.
Frontend: build, lint, сценарии API и проверка доступным браузерным инструментом.
