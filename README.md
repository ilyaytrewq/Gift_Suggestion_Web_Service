# Gift Suggestion Web Service

Монорепозиторий веб-сервиса подбора подарков. Состоит из трёх сервисов:

- **backend** — Go 1.26 + Gin + PostgreSQL (`services/backend/`)
- **frontend** — React 19 + Vite + TypeScript (`services/frontend/`)
- **ml-service** — Python 3.13 + FastAPI + gRPC (`services/ml/`)

## Что реализовано

### Backend

- Health: `GET /health/live`, `GET /health/ready`
- Auth: регистрация, логин, refresh, logout, `GET/PATCH /api/v1/users/me`
- Password reset: `request` + `confirm`
- Email verification: `confirm`
- Catalog: список подарков с фильтрами, карточка подарка с **offers (мультимагазин)**, категории
- Похожие подарки: `GET /api/v1/catalog/gifts/{gift_id}/similar`
- Wishlist: единый список желаний пользователя
- Catalog import: admin upload CSV / JSON / XLSX с автоматическим созданием `gift_offers`; алиасы заголовков (`title`, `store_url`, `currency`); JSON поддерживает массив `offers`
- Recommendation: мастер подбора с полями повод / бюджет / отношения / **пол** / возраст / интересы; ранжирование через ML (LightGBM gRPC) или fallback; объяснения; альтернативы
- Tracking: события просмотра, добавления в избранное, переходов по ссылкам
- VK integration: scaffold (хранение соединений, шифрование токена)

### ML сервис

- gRPC `RankingService.Rank` (порт 50051) — при наличии обученной модели использует LightGBM LambdaRank, иначе echo-заглушка
- FastAPI `/healthz` (порт 8081)
- `training/` — скрипты подготовки датасета, обучения, оценки NDCG
- Инструкции по обучению: `services/ml/Guide.md`

### Frontend

- Главная страница, каталог с поиском и фильтрами, карточка подарка
- Мастер рекомендаций (`/recommendation`)
- Wishlist (`/wishlist`)
- Login / Register / запрос сброса пароля

### Данные каталога

- `scripts/data/fetch_wb.py` — выгрузка из Wildberries (публичный поиск)
- `scripts/data/fetch_aliexpress.py` — выгрузка из AliExpress Affiliate API
- `scripts/data/normalize_to_catalog.py` — нормализация и слияние CSV
- `services/ml/dataset/dataset_example.csv` — 10 000 синтетических записей (RU + CN)
- Инструкции: `docs/data.md`

## Структура репозитория

```
.
├── docker-compose.yml
├── README.md
├── AGENTS.md
├── CLAUDE.md
├── Plan.md                          ← план реализации
├── docs/
│   ├── ci-cd.md
│   └── data.md                      ← источники данных + инструкции
├── scripts/data/                    ← скрипты выгрузки RU/CN каталога
└── services/
    ├── backend/
    │   ├── api/proto/ranking/v1/    ← ranking.proto (gRPC контракт)
    │   ├── cmd/api/
    │   ├── docs/openapi/backend.yaml
    │   ├── internal/
    │   ├── migrations/              ← 15 миграций
    │   └── Taskfile.yaml
    ├── frontend/
    │   └── src/
    └── ml/
        ├── Guide.md                 ← инструкции по обучению
        ├── ml_service/              ← gRPC сервер, ranker, feature_builder
        ├── training/                ← скрипты обучения
        ├── tests/
        ├── dataset/
        └── models/                  ← сюда кладётся .pkl модель
```

## Быстрый запуск

Требования: Docker + Docker Compose.

```bash
docker compose up --build postgres backend ml-service frontend
```

Локальные URL:
- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`
- ML healthz: `http://localhost:8081/healthz`
- ML gRPC: `localhost:50051`
- PostgreSQL: `localhost:5432`

По умолчанию:
- Миграции запускаются при старте backend (`DB_MIGRATIONS_ENABLED=true`)
- ML gRPC **включён** (`ML_GRPC_ENABLED=true`); если `services/ml/models/lightgbm_v0_1_0.pkl` отсутствует — использует echo-заглушку
- VK отключён (`VK_ENABLED=false`)
- Email delivery отключён (`EMAIL_ENABLED=false`)

## Локальная разработка

### Backend

```bash
cd services/backend
GOTOOLCHAIN=auto ya tool go test ./...
task lint
task build
ya tool go run ./cmd/api
```

Если `ya` недоступен (например, в типовом GitHub Actions), используйте обычный `go` — скрипт `scripts/ci/backend-checks.sh` и задачи Task сами подставят `go`, когда `ya` не в `PATH`.

Генерация proto:
```bash
task gen:proto
```

### Frontend

```bash
cd services/frontend
npm ci
npm run generate:api   # после изменений OpenAPI
npm run lint
npm run build
npm run dev
```

### ML сервис

```bash
cd services/ml
pytest                 # тесты (9 штук)

# Запуск без Docker
ML_MODEL_PATH=models/lightgbm_v0_1_0.pkl python3 -m ml_service.main
```

Обучение модели — см. `services/ml/Guide.md`.

## OpenAPI и контракты

Спецификация: `services/backend/docs/openapi/backend.yaml`

После изменений OpenAPI обязательно:
```bash
cd services/frontend && npm run generate:api
```

## Основные env-переменные

| Переменная | Дефолт | Описание |
|---|---|---|
| `DB_DSN` | `postgres://gift:gift@postgres:5432/gift_suggestion?sslmode=disable` | PostgreSQL DSN |
| `DB_MIGRATIONS_ENABLED` | `true` | Автомиграция при старте |
| `ML_GRPC_ENABLED` | `true` | Включить ML gRPC |
| `ML_GRPC_ADDR` | `ml-service:50051` | Адрес ML сервиса |
| `ML_MODEL_PATH` | `models/lightgbm_v0_1_0.pkl` | Путь к модели |
| `VK_ENABLED` | `false` | VK интеграция |
| `EMAIL_ENABLED` | `false` | Email delivery |
| `AUTH_JWT_SECRET` | `change-me-please` | JWT секрет |
| `VITE_API_BASE_URL` | `` | Backend URL для фронта |
| `ALIEXPRESS_APP_KEY` | — | Ключ AliExpress Affiliate API |
| `ALIEXPRESS_APP_SECRET` | — | Секрет AliExpress Affiliate API |

## CI/CD

GitHub Actions:
- `CI` — backend tests/lint/build, frontend install/generate/lint/build, ML tests (`pytest`), docker compose smoke
- `CD` — сборка и публикация образов backend / frontend / ml-service, deploy через SSH

Подробнее: `docs/ci-cd.md`

## Ограничения текущего состояния

- ML модель не обучена — при старте сервис использует echo-stub (кандидаты в исходном порядке). Обучи через `services/ml/Guide.md`
- VK OAuth flow scaffold: хранение / шифрование есть, реальный вызов VK API не реализован
- Frontend: нет UI для admin import, tracking событий, VK интеграции, confirm email/password-reset
- `ProfilePage` существует, но не подключена в router
- Данные каталога: текущий `dataset_example.csv` синтетический. Для реальных данных используй `scripts/data/`
