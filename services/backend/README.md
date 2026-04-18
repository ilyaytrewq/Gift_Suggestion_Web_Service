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
- `POST /api/v1/auth/password-reset/request`
- `GET /api/v1/users/me`
- `PATCH /api/v1/users/me`

Password reset пока реализован как foundation: backend сохраняет одноразовый reset-token hash и TTL в БД, но подтверждение reset через отдельный delivery flow еще не выведено наружу.

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

## OpenAPI

HTTP contracts ведутся в OpenAPI-формате в файле:

```bash
services/backend/docs/openapi/backend.yaml
```

Сейчас спецификация покрывает health, auth/user foundation, catalog read и wishlist endpoints.

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

## Полезные команды

Из `services/backend`:

```bash
task run
task tests
task lint
```
