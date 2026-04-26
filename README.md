# Gift Suggestion Web Service

Монорепозиторий MVP-сервиса подбора подарков. В текущем состоянии проект состоит из:

- backend на Go 1.26 + Gin + PostgreSQL в `services/backend`
- frontend на React 19 + Vite в `services/frontend`
- локального окружения через `docker-compose.yml`

## Что реализовано

### Backend

Сервис уже поднимает и использует:

- health endpoints: `GET /health/live`, `GET /health/ready`
- auth/user flow: регистрация, логин, refresh, logout, `GET/PATCH /api/v1/users/me`
- password reset flow: `request` и `confirm`
- email verification confirm flow
- catalog read API: подарки, категории, карточка подарка
- wishlist API
- admin catalog import API для `CSV`, `JSON`, `XLSX`
- recommendation API с fallback-ранжированием без ML
- tracking events ingestion API
- VK integration scaffold с feature flag

### Frontend

В браузере сейчас доступны:

- главная страница
- каталог с поиском и фильтрами
- страница подарка
- мастер рекомендаций
- login/register
- запрос на восстановление пароля

## Структура репозитория

```text
.
├── docker-compose.yml
├── README.md
├── AGENTS.md
└── services
    ├── backend
    │   ├── cmd/api
    │   ├── docs/openapi/backend.yaml
    │   ├── internal
    │   ├── migrations
    │   └── Taskfile.yaml
    └── frontend
        ├── src
        └── design
```

## Быстрый запуск

Требования:

- Docker + Docker Compose

Из корня репозитория:

```bash
docker compose up --build postgres backend frontend
```

Локальные URL:

- frontend: `http://localhost:5173`
- backend: `http://localhost:8080`
- backend readiness: `http://localhost:8080/health/ready`
- PostgreSQL: `localhost:5432`

Что важно по умолчанию:

- миграции запускаются при старте backend, если `DB_MIGRATIONS_ENABLED=true`
- ML gRPC выключен, поэтому рекомендации работают через deterministic fallback
- VK integration выключена через `VK_ENABLED=false`
- email delivery выключен через `EMAIL_ENABLED=false`, а `noop` sender не ломает регистрацию и reset flow

## Локальная разработка

### Backend

```bash
cd services/backend
go test ./...
task lint
go run ./cmd/api
```

Полезные task-команды:

- `task tests`
- `task lint`
- `task build`
- `task format`

### Frontend

```bash
cd services/frontend
npm ci
npm run generate:api
npm run lint
npm run build
npm run dev
```

По умолчанию frontend обращается к `http://localhost:8080`, если `VITE_API_BASE_URL` не задан.

## OpenAPI и контракты

Основной HTTP-контракт хранится в:

```bash
services/backend/docs/openapi/backend.yaml
```

Frontend генерирует типы из этой спецификации:

```bash
cd services/frontend
npm run generate:api
```

Сейчас это критичная точка синхронизации: backend-код уже содержит `POST /api/v1/auth/password-reset/confirm` и `POST /api/v1/auth/email-verification/confirm`, поэтому при изменениях auth-flow нужно обновлять и Go-код, и OpenAPI, и сгенерированные frontend-типы.

## Основные env-переменные

Базовые значения для compose лежат в корневом `.env.example`, для локального backend-запуска без Docker есть `services/backend/.env.example`.

На практике чаще всего меняются:

- `DB_DSN`
- `DB_MIGRATIONS_ENABLED`
- `ML_GRPC_ENABLED`, `ML_GRPC_ADDR`, `ML_GRPC_REQUEST_TIMEOUT`
- `VK_ENABLED`, `VK_TOKEN_ENCRYPTION_KEY`
- `AUTH_JWT_SECRET`
- `EMAIL_ENABLED`, `EMAIL_PROVIDER`, `EMAIL_FROM_EMAIL`, `FRONTEND_BASE_URL`
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_USE_TLS`
- `VITE_API_BASE_URL`

Если включать SMTP (`EMAIL_ENABLED=true` и `EMAIL_PROVIDER=smtp`), нужно как минимум задать:

- `EMAIL_FROM_EMAIL`
- `SMTP_HOST`
- `SMTP_PORT`

Сейчас email используется для:

- подтверждения почты после регистрации
- запроса на восстановление пароля

В dev/test можно оставить `EMAIL_PROVIDER=noop`: backend сохранит verification/reset foundation, но не будет отправлять реальные письма и не будет логировать токены.

## CI/CD

Для репозитория используется GitHub Actions.

- `CI` прогоняет backend tests/lint/build, frontend install/generate/lint/build, проверку `docker compose` и migration smoke-check через подъем `postgres` + `backend`.
- `CD` собирает и публикует Docker-образы `backend` и `frontend`, после чего выкатывает их на параметризованный remote target через `ssh` и `docker compose`.
- Все deployment-specific значения должны приходить из GitHub `vars`/`secrets` и runtime env-файла. В workflow не зашиваются registry coordinates, домены, SSH host/user, DB credentials, JWT secrets и другие environment-specific значения.

Подробная схема, список required variables/secrets и deployment scaffold описаны в [docs/ci-cd.md](docs/ci-cd.md).

## Ограничения текущего состояния

- Реальный ML service не обязателен: при `ML_GRPC_ENABLED=false` или ошибке gRPC backend использует fallback ranking.
- VK integration остаётся scaffold-модулем: storage и endpoint wiring есть, полноценный внешний VK flow ещё не доведён.
- Frontend пока не выводит весь backend-функционал: нет UI для wishlist, import jobs, tracking, VK integration.
- В коде frontend есть профильная страница, но она не подключена в router.
- Frontend хранит access token только в памяти и восстанавливает сессию через refresh-cookie bootstrap.
- Frontend сейчас покрывает только запрос на восстановление пароля; confirm reset и confirm email verification на UI ещё не выведены.

## Что смотреть в первую очередь

- backend composition root: `services/backend/internal/app/app.go`
- backend router: `services/backend/internal/transport/http/router.go`
- frontend router: `services/frontend/src/app/router/router.tsx`
- frontend API layer: `services/frontend/src/shared/api`
