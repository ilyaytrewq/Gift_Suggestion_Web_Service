# Backend

`services/backend` — Go 1.26 API на Gin, оркестрирует backend-сценарии сервиса подбора подарков, работает с PostgreSQL, интегрируется с ML-service по gRPC и VK API.

## Реализованные эндпоинты

### Health

- `GET /health/live`
- `GET /health/ready`
- `GET /api/v1/health/live`
- `GET /api/v1/health/ready`

Readiness и ML — **два разных режима**:

1. **Старт процесса** — при `ML_GRPC_ENABLED=true` клиент `mlgrpc.NewClient` выполняет gRPC health check до поднятия HTTP; если ML недоступен, инициализация падает с ошибкой подключения к ML (процесс не слушает порт).
2. **`GET /health/ready` во время работы** — обязателен только PostgreSQL: при его сбое HTTP `503` и суммарный `data.status` = `down`. Зонд `ml_service` в отчёте **необязательный** — при его `down` суммарный статус может остаться `up`.
3. **Fallback recommendation** — при `ML_GRPC_ENABLED=false` или при ошибке/таймауте `Rank` в запросе используется эвристическое ранжирование в Go; это не переводит readiness в `down`, если Postgres доступен.

### Auth / User

- `POST /api/v1/users`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/password-reset/request`
- `POST /api/v1/auth/password-reset/confirm`
- `POST /api/v1/auth/email-verification/confirm`
- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`

Токены сброса пароля и верификации email хранятся только в виде хэша и никогда не логируются.

### Catalog

- `GET /api/v1/catalog/gifts`
- `GET /api/v1/catalog/gifts/{gift_id}`
- `GET /api/v1/catalog/gifts/{gift_id}/similar`
- `GET /api/v1/catalog/categories`

Фильтры списка: `q`, `category_id`, `min_price`, `max_price`, `age_restriction`, `has_image`, `limit`, `offset`, `sort`.

### Wishlist

- `GET /api/v1/wishlist`
- `POST /api/v1/wishlist/items`
- `DELETE /api/v1/wishlist/items/{gift_id}`
- `DELETE /api/v1/wishlist`

Все wishlist-эндпоинты требуют JWT. Обращение к чужому wishlist маскируется ответом `404`.

### Catalog import (admin)

- `POST /api/v1/admin/import-jobs`
- `GET /api/v1/admin/import-jobs/{job_id}`
- `GET /api/v1/admin/import-jobs/{job_id}/errors`

Требуется роль `admin`. Форматы: CSV, JSON, XLSX. Импорт синхронный; результат сохраняется в `import_jobs` / `import_errors`.

### Recommendation

- `POST /api/v1/recommendations`
- `GET /api/v1/recommendations/{request_id}`

`POST` доступен без авторизации; JWT опционален. Pipeline: hard filters → ML ranking (gRPC) → fallback при ошибке ML → explanations → alternatives.

### Tracking

- `POST /api/v1/tracking/events`

Типы событий: `recommendation_request`, `card_view`, `wishlist_add`, `outbound_click`. Требует JWT.

### VK интеграция

- `GET /api/v1/integrations/vk/connection`
- `PUT /api/v1/integrations/vk/connection`
- `DELETE /api/v1/integrations/vk/connection`
- `POST /api/v1/integrations/vk/connection/sync-interests`

Все эндпоинты требуют JWT. При `VK_ENABLED=false` возвращают `503 vk_integration_disabled`.

Реализованный импорт интересов (`sync-interests`):

- вызывает VK API `groups.get` методом POST (токен передаётся только в теле запроса)
- версия API: `5.199`; поля: `name`, `description`, `activity`
- пагинация: `count=1000`, повторные запросы до получения всех групп
- названия групп сохраняются как imported interests с `source_label=vk_group`
- дедупликация по нормализованному значению

Обработка ошибок VK API:

| error_code | apperrors code |
|---|---|
| 5, 1117 | `vk_token_invalid` |
| 6, 29 | `vk_rate_limited` |
| 260 | `vk_groups_access_denied` |

Для хранения access token нужен `VK_TOKEN_ENCRYPTION_KEY` — base64-закодированный 32-байтовый AES-ключ. Без него `PUT /connection` с токеном вернёт `503 vk_token_storage_not_configured`.

## OpenAPI

Спецификация: `services/backend/docs/openapi/backend.yaml`

Покрывает все реализованные эндпоинты. После изменений спецификации обязательно:

```bash
cd services/frontend && npm run generate:api
```

## Email delivery

Infrastructure: `internal/platform/email`. Auth-specific notifier: `internal/modules/auth/infra/email`.

Сценарии, использующие email:
- верификация email при регистрации
- запрос сброса пароля

Переменные для рабочего SMTP-режима:

```
EMAIL_ENABLED=true
EMAIL_PROVIDER=smtp
EMAIL_FROM_EMAIL=...
SMTP_HOST=...
SMTP_PORT=587
SMTP_USERNAME=...
SMTP_PASSWORD=...
FRONTEND_BASE_URL=http://localhost:5173
```

По умолчанию используется noop-отправитель (`EMAIL_ENABLED=false`).

## Конфигурация

Все параметры читаются из окружения. Пример — `.env.example`.

Ключевые переменные:

| Переменная | Дефолт | Описание |
|---|---|---|
| `DB_DSN` | — (обязателен) | PostgreSQL DSN |
| `DB_MIGRATIONS_ENABLED` | `true` | Автомиграция при старте |
| `AUTH_JWT_SECRET` | — (обязателен, ≥16 символов) | JWT секрет |
| `VK_ENABLED` | `false` | Включить VK интеграцию |
| `VK_TOKEN_ENCRYPTION_KEY` | — | Base64 AES-256 ключ (`openssl rand -base64 32`) |
| `VK_REQUEST_TIMEOUT` | `3s` | Таймаут запроса к VK API |
| `ML_GRPC_ENABLED` | `false` | Включить ML gRPC клиент |
| `ML_GRPC_ADDR` | — | Адрес ML сервиса |
| `EMAIL_ENABLED` | `false` | Включить email delivery |

## Локальный запуск

```bash
cd services/backend
cp .env.example .env   # задать DB_DSN, AUTH_JWT_SECRET и другие обязательные переменные
ya tool go run ./cmd/api
```

Если `ya` недоступен, используйте обычный `go run ./cmd/api`.

## Через Docker Compose

Из корня репозитория:

```bash
docker compose up --build backend postgres
```

Поднимает `postgres:16-alpine` на `:5432` и `backend` на `:8080`. В корневом `docker-compose.yml` по умолчанию `ML_GRPC_ENABLED=true` и ожидается `ml-service`; без него (например, только `backend` + `postgres`) задайте `ML_GRPC_ENABLED=false` или поднимите ML. Recommendation при ошибке ранжирования всё равно уходит в fallback в Go.

## Полезные команды

Из `services/backend`:

```bash
task tests   # go test ./... -v
task lint    # golangci-lint
task build   # компиляция бинаря
task run     # запуск локально
```
