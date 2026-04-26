# Backend Bootstrap

`services/backend` это Go API gateway на `Gin`, который оркестрирует backend-сценарии сервиса подбора подарков, работает с PostgreSQL и готовится к интеграции с ML-service по gRPC.

## Что уже входит в bootstrap

- composition root через `cmd/api` + `internal/app`;
- конфиг из env и структурированный `slog`;
- подключение к PostgreSQL и запуск embedded migrations;
- базовый `health` модуль с `domain/usecase/delivery/http/infra` слоями;
- единый JSON envelope для success/error ответов;
- Gin router, request-id middleware и readiness/liveness endpoints;
- scaffold gRPC client для ML-service;
- Dockerfile и `docker-compose.yml` для локального окружения.

## User/Auth minimum

После `feature/user-postgres-http` backend поддерживает:

- `POST /api/v1/users`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/password-reset/request`
- `POST /api/v1/auth/password-reset/confirm`
- `POST /api/v1/auth/email-verification/confirm`
- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`

Текущий auth/email scope:

- регистрация создаёт пользователя в статусе `email_verified=false`, сохраняет одноразовый verification token hash и инициирует отправку verification email;
- password reset request сохраняет одноразовый reset-token hash и инициирует отправку reset email;
- verification/reset tokens никогда не возвращаются в API response и не логируются;
- `confirm` endpoints принимают raw token, хэшируют его в backend и атомарно завершают verification/reset flow.

## Catalog read API

После `feature/catalog-postgres-http` backend также поддерживает:

- `GET /api/v1/catalog/gifts`
- `GET /api/v1/catalog/gifts/{gift_id}`
- `GET /api/v1/catalog/categories`

List endpoints используют общий envelope и возвращают `data.items` + `data.page`.
Для списка подарков сейчас доступны фильтры `q`, `category_id`, `min_price`, `max_price`, `age_restriction`, `has_image`, `limit`, `offset`, `sort`.

## Wishlist API

После `feature/wishlist-postgres-http` backend также поддерживает:

- `POST /api/v1/wishlists`
- `GET /api/v1/wishlists`
- `GET /api/v1/wishlists/{wishlist_id}`
- `POST /api/v1/wishlists/{wishlist_id}/items`
- `DELETE /api/v1/wishlists/{wishlist_id}/items/{gift_id}`
- `DELETE /api/v1/wishlists/{wishlist_id}`

Все wishlist endpoints требуют JWT access token и работают только с wishlist текущего пользователя.
Попытка обратиться к чужому wishlist маскируется ответом `404`, чтобы не раскрывать существование чужих списков.

## Catalog import API

После `feature/import-jobs` backend также поддерживает admin-oriented import flow:

- `POST /api/v1/admin/import-jobs`
- `GET /api/v1/admin/import-jobs/{job_id}`
- `GET /api/v1/admin/import-jobs/{job_id}/errors`

Все import endpoints требуют JWT access token с ролью `admin`.
Импорт выполняется синхронно в рамках `POST`, но результат всегда сохраняется в `import_jobs` и `import_errors`.

Поддерживаемые форматы:

- `CSV`
- `JSON`
- `XLSX`

Обязательные поля записи:

- `name`
- `category`
- `price`
- `description`
- `store_link`

Опциональные поля:

- `image`
- `age_restriction`
- `source`

Текущий импортный контракт:

- `category` резолвится по уже существующим `categories` case-insensitive; неизвестная категория уходит в `import_errors`;
- дубликаты внутри файла и дубликаты относительно каталога (`normalized(name) + normalized(store_link)`) не импортируются и попадают в `import_errors`;
- частично валидный файл допустим: валидные записи сохраняются, невалидные фиксируются в отчёте;
- максимальный размер файла задаётся через `IMPORT_MAX_FILE_SIZE_BYTES`.

Для `CSV` и `XLSX` первая строка должна содержать headers.
Для `JSON` поддерживается массив объектов, а также объект вида `{ "items": [...] }`.

## Recommendation API

После `feature/recommendation-ml-gateway` backend также поддерживает recommendation flow:

- `POST /api/v1/recommendations`
- `GET /api/v1/recommendations/{request_id}`

`POST /api/v1/recommendations` доступен без авторизации.
JWT access token опционален: при его наличии backend может использовать user/wishlist контекст для дополнительной персонализации.
`GET /api/v1/recommendations/{request_id}` остаётся auth-only endpoint.

Поддерживаемый request payload:

- `budget_max` — обязательный верхний предел бюджета;
- `recipient_age` — optional hard filter по возрастному ограничению;
- `occasion`, `relationship` — optional questionnaire context;
- `preferred_category_ids` — optional hard filter по категориям;
- `interests` — optional ranking context;
- `top_n` — optional limit, по умолчанию `5`, максимум `10`;
- `use_wishlist_context` — optional flag, по умолчанию `true`.

Онлайн pipeline recommendation:

- backend читает кандидатов из уже нормализованного каталога;
- применяет hard filters по `budget_max`, `recipient_age` и `preferred_category_ids`;
- подготавливает candidate pool и вызывает ML gateway;
- если ML недоступен, таймаутится или возвращает невалидный ranking, backend переключается на deterministic fallback;
- explanations синтезируются backend-ом, если ML их не прислал;
- alternatives достраиваются из remaining candidate pool, если ML их не прислал или прислал частично.

Текущие execution guarantees:

- per-call timeout на ML ranking задаётся через `ML_GRPC_REQUEST_TIMEOUT`;
- число retry ограничивается `ML_GRPC_MAX_RETRIES`;
- backend хранит `recommendation_requests` и `recommendation_results` для traceability и следующей tracking-ветки.

## Tracking API

После `feature/tracking-events` backend поддерживает auth-only ingestion path для user interaction events:

- `POST /api/v1/tracking/events`

Минимально поддерживаемые типы:

- `recommendation_request`
- `card_view`
- `wishlist_add`
- `outbound_click`

Контракт события:

- `type` обязателен;
- `gift_id` обязателен для `card_view`, `wishlist_add`, `outbound_click`;
- `wishlist_id` обязателен для `wishlist_add`;
- `recommendation_request_id` обязателен для `recommendation_request` и optional для остальных;
- `client_event_id` optional и используется как idempotency key для повторной отправки одного и того же события;
- `metadata.surface` и `metadata.position` optional и валидируются как компактный аналитический контекст.

Текущая реализация хранит события в append-only `tracking_events` и валидирует ссылки на:

- текущего пользователя из JWT access token;
- `recommendation_request_id` с ownership check;
- `wishlist_id` с ownership check;
- `gift_id` через catalog existence check.

Автоматическая серверная эмиссия tracking-событий из `recommendation` и `wishlist` use-case в этой ветке не добавлялась сознательно.
Сейчас ветка даёт единый ingestion endpoint и storage foundation без расширения чужих модулей.

## VK integration scaffold

После `feature/vk-connections` backend поддерживает auth-only scaffold для consent-aware VK linkage:

- `GET /api/v1/integrations/vk/connection`
- `PUT /api/v1/integrations/vk/connection`
- `DELETE /api/v1/integrations/vk/connection`
- `POST /api/v1/integrations/vk/connection/sync-interests`

VK integration deliberately остаётся supporting-module scaffold:

- backend хранит один `vk_connection` на пользователя;
- consent и connection state валидируются в use-case, не в handler;
- access token никогда не возвращается в API и, если передан, хранится только в зашифрованном виде;
- imported interests сохраняются snapshot-моделью в `vk_imported_interests`;
- если `VK_ENABLED=false`, endpoints возвращают `503 vk_integration_disabled`.

Текущий safe scope этой ветки:

- connect/disconnect текущего пользователя;
- безопасное хранение non-secret metadata (`screen_name`, `profile_url`, scopes, expires_at`);
- foundation для `sync-interests` с pluggable VK importer;
- feature-flag friendly wiring через `VK_ENABLED`, `VK_REQUEST_TIMEOUT`, `VK_TOKEN_ENCRYPTION_KEY`.

Текущая граница реализованного:

- реальный VK API/OAuth flow намеренно не реализован;
- встроенный importer возвращает controlled scaffold error, пока внешний VK client не добавлен;
- `sync-interests` готов к рабочему провайдеру, но не делает опасных предположений о внешнем API.

## OpenAPI

HTTP contracts ведутся в OpenAPI-формате в файле:

```bash
services/backend/docs/openapi/backend.yaml
```

Сейчас спецификация покрывает health, auth/user foundation, email verification, password reset confirm, catalog read, wishlist, admin catalog import, recommendation, tracking и VK integration scaffold endpoints.

## Email delivery

Reusable email infrastructure:

- generic sender infrastructure: `internal/platform/email`
- auth-specific notifier/templates: `internal/modules/auth/infra/email`
- SMTP details остаются в config/env и не протаскиваются в use-case слой
- dev/test режим использует `noop` sender и не выводит токены в логи

Сценарии, которые используют email уже сейчас:

- registration email verification
- password reset request

Обязательные env для рабочего SMTP режима:

- `EMAIL_ENABLED=true`
- `EMAIL_PROVIDER=smtp`
- `EMAIL_FROM_EMAIL`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `FRONTEND_BASE_URL`

Дополнительные env:

- `EMAIL_FROM_NAME`
- `EMAIL_SEND_TIMEOUT`
- `SMTP_USE_TLS`
- `AUTH_EMAIL_VERIFICATION_TTL`
- `AUTH_PASSWORD_RESET_TTL`

Сценарии, оставленные на будущее:

- уведомление о смене пароля
- уведомление о входе с нового устройства
- подтверждение смены email
- admin/system import notifications
- onboarding/welcome email

## Локальный запуск backend

1. Перейти в `services/backend`.
2. Подготовить env на основе `.env.example`.
3. Запустить PostgreSQL через корневой `docker-compose.yml` или использовать локальную БД.
4. Запустить backend:

```bash
go run ./cmd/api
```

Локальные health endpoints:

- `GET /health/live`
- `GET /health/ready`
- `GET /api/v1/health/live`
- `GET /api/v1/health/ready`

## Локальный запуск через Docker Compose

Из корня репозитория:

```bash
docker compose up --build backend postgres
```

Compose поднимает:

- `postgres:16-alpine` на `localhost:5432`;
- `backend` на `localhost:8080`.

По умолчанию в compose отключен ML gRPC (`ML_GRPC_ENABLED=false`), поэтому bootstrap можно поднять без ML-service.
В этом режиме recommendation flow продолжает работать через backend fallback ranking.
VK scaffold по умолчанию тоже отключён (`VK_ENABLED=false`). Для сохранения access token в зашифрованном виде нужно задать `VK_TOKEN_ENCRYPTION_KEY` как base64-encoded 32-byte AES key.

## Полезные команды

Из `services/backend`:

```bash
task run
task tests
task lint
```
