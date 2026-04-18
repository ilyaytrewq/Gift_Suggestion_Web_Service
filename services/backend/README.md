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
