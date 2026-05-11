# Gift Suggestion Web Service
## http://giftsuggestion.ru/
Монорепозиторий веб-сервиса подбора подарков. Состоит из трёх сервисов:

- **backend** — Go 1.26 + Gin + PostgreSQL (`services/backend/`)
- **frontend** — React 19 + Vite + TypeScript (`services/frontend/`)
- **ml-service** — Python 3.13 + FastAPI + gRPC (`services/ml/`)

## Что реализовано

### Backend

- Health: `GET /health/live`, `GET /health/ready` — `503` только если недоступен обязательный Postgres; ML gRPC в readiness необязателен, но при `ML_GRPC_ENABLED=true` доступный ML нужен **на старте** процесса backend
- Auth: регистрация, логин, refresh, logout, `GET/PATCH /api/v1/users/me`
- Password reset: `request` + `confirm`
- Email verification: `confirm`
- Catalog: список подарков с фильтрами, карточка подарка с offers (мультимагазин), категории
- Похожие подарки: `GET /api/v1/catalog/gifts/{gift_id}/similar`
- Wishlist: единый список желаний пользователя
- Catalog import: admin upload CSV / JSON / XLSX с автоматическим созданием `gift_offers`
- Recommendation: мастер подбора с полями повод / бюджет / отношения / пол / возраст / интересы; ранжирование через ML (LightGBM gRPC) или fallback; объяснения; альтернативы
- Tracking: события просмотра, добавления в избранное, переходов по ссылкам
- VK integration: реальный импорт групп из VK API (`groups.get` 5.199) с пагинацией; шифрование токенов AES-256-GCM; обработка ошибок API (невалидный токен, rate limit, закрытый список групп)

### ML сервис

- gRPC `RankingService.Rank` (порт 50051) — при наличии обученной модели использует LightGBM LambdaRank, иначе echo-заглушка
- FastAPI `/healthz` (порт 8081)
- `training/` — скрипты подготовки датасета, обучения, оценки NDCG
- Инструкции по обучению: `services/ml/Guide.md`

### Frontend

- Главная страница, каталог с поиском и фильтрами, карточка подарка
- Мастер рекомендаций (`/recommendation`)
- Wishlist (`/wishlist`)
- Профиль (`/profile`) с редактированием имени и панелью VK-интеграции
- Login / Register / сброс пароля / подтверждение email
- Admin import (`/admin/import`)
- VK OAuth implicit flow: редирект на oauth.vk.com, callback-страница (`/auth/vk-callback`), сохранение токена на бэкенде, синхронизация интересов

### Данные каталога

- `scripts/data/fetch_wb.py` — выгрузка из Wildberries (публичный поиск)
- `scripts/data/fetch_aliexpress.py` — выгрузка из AliExpress Affiliate API
- `scripts/data/normalize_to_catalog.py` — нормализация и слияние CSV
- `services/ml/dataset/dataset_example.csv` — 10 000 синтетических записей (RU + CN)
- Инструкции: `docs/data.md`

#### Перенос датасета на сервер и импорт в БД

`scripts/data/upload_and_import.sh` — переносит датасеты на сервер и заполняет каталог товарами через API импорта.

**Требования:** `ssh`, `rsync`, `curl`, `python3` (локально); `docker` (на сервере).

```bash
# Полный запуск: rsync датасетов + импорт всего каталога
bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154

# С явным SSH-ключом
bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --key ~/.ssh/id_rsa

# Тест: 2000 строк, без rsync (файлы уже на сервере)
bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-upload --max-rows 2000

# Только rsync датасетов, без импорта в БД
bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --max-rows 0

# Повторный импорт (категории и admin уже созданы)
bash scripts/data/upload_and_import.sh --server ilyaytrewq@51.250.117.154 --skip-db-init
```

Что делает скрипт:
1. `rsync` CSV/parquet/моделей в `~/gift-datasets/` на сервере
2. Создаёт категории в БД через `docker exec psql`
3. Регистрирует admin-пользователя и устанавливает роль
4. Конвертирует `catalog_real.csv` в формат импорта (локально)
5. Загружает чанками через `POST /api/v1/admin/import-jobs`

Основные параметры:

| Флаг | По умолчанию | Описание |
|---|---|---|
| `--server ilyaytrewq@51.250.117.154` | — | SSH-цель (обязателен) |
| `--api URL` | `http://HOST:8080` | URL бэкенда (автоопределяется) |
| `--catalog PATH` | `services/ml/dataset/catalog_real.csv` | Локальный CSV |
| `--chunk N` | `5000` | Строк на чанк |
| `--compose-project NAME` | `gift-suggestion` | Имя Docker Compose проекта на сервере |
| `--skip-upload` | — | Пропустить rsync |
| `--skip-db-init` | — | Пропустить создание категорий/admin |

#### Перевод пользователя в роль admin

`scripts/data/make_admin.sh` — устанавливает роль `admin` для существующего пользователя по email.

```bash
# Локально
bash scripts/data/make_admin.sh --email user@example.com

# На сервере через SSH
bash scripts/data/make_admin.sh --email user@example.com --server ilyaytrewq@51.250.117.154
bash scripts/data/make_admin.sh --email user@example.com --server ilyaytrewq@51.250.117.154 --key ~/.ssh/id_rsa
```

Скрипт проверяет, что пользователь существует в БД, выводит его текущую запись и выполняет `UPDATE users SET role = 'admin'`. Для локального запуска использует `psql` напрямую, для сервера — `ssh` + `docker exec` в postgres-контейнер.

## Структура репозитория

```
.
├── docker-compose.yml
├── README.md
├── docs/
│   ├── ci-cd.md
│   └── data.md
├── scripts/data/
└── services/
    ├── backend/
    │   ├── api/proto/ranking/v1/
    │   ├── cmd/api/
    │   ├── docs/openapi/backend.yaml
    │   ├── internal/
    │   ├── migrations/              ← 16 миграций
    │   └── Taskfile.yaml
    ├── frontend/
    │   └── src/
    └── ml/
        ├── Guide.md
        ├── ml_service/
        ├── training/
        ├── tests/
        ├── dataset/
        └── models/
```

## Быстрый запуск

Требования: Docker + Docker Compose.

```bash
docker compose up --build postgres backend ml-service frontend
```

Локальные URL (фронт на стандартном порту 80, API проксируется через nginx):
- Frontend: `http://localhost` (или `http://giftsuggestion.ru`, если домен указывает на машину)
- Backend API через фронт: `http://localhost/api/v1/...`
- Backend напрямую (отладка): `http://localhost:8080`
- ML healthz: `http://localhost:8081/healthz`
- ML gRPC: `localhost:50051`
- PostgreSQL: `localhost:5432`

По умолчанию:
- Миграции запускаются при старте backend (`DB_MIGRATIONS_ENABLED=true`)
- ML gRPC **включён** (`ML_GRPC_ENABLED=true`); если модель отсутствует — используется echo-заглушка
- VK отключён (`VK_ENABLED=false`) — см. раздел ниже
- Email delivery отключён (`EMAIL_ENABLED=false`) — использует noop-отправитель

## Локальная разработка

### Backend

```bash
cd services/backend
GOTOOLCHAIN=auto ya tool go test ./...
task lint
task build
ya tool go run ./cmd/api
```

Если `ya` недоступен (например, в GitHub Actions), используйте обычный `go` — скрипт `scripts/ci/backend-checks.sh` и задачи Task сами подставят `go`, когда `ya` не в `PATH`.

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
pytest

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
| `ML_GRPC_ENABLED` | `true` (compose) / `false` (standalone) | Включить ML gRPC |
| `ML_GRPC_ADDR` | `ml-service:50051` | Адрес ML сервиса |
| `ML_MODEL_PATH` | `models/lightgbm_v2_1_0.pkl` | Путь к модели |
| `VK_ENABLED` | `false` | Включить VK интеграцию |
| `VK_TOKEN_ENCRYPTION_KEY` | — | Base64-ключ AES-256 для хранения токенов (`openssl rand -base64 32`) |
| `EMAIL_ENABLED` | `false` | Email delivery |
| `AUTH_JWT_SECRET` | `change-me-please` | JWT секрет |
| `FRONTEND_PORT` | `80` | Порт хоста для фронта (80 — без `:порта` в URL) |
| `FRONTEND_PUBLIC_URL` | `http://localhost` | Публичный URL фронта (отображается в логах nginx) |
| `FRONTEND_BASE_URL` | `http://localhost` | URL фронта для ссылок в email (на проде: `http://giftsuggestion.ru`) |
| `VITE_API_BASE_URL` | — | Backend URL для фронта |
| `VITE_VK_APP_ID` | — | ID VK-приложения (для OAuth кнопки в профиле) |
| `VITE_VK_REDIRECT_URI` | — | Redirect URI OAuth (`https://giftsuggestion.ru/auth/vk-callback` на проде) |
| `ALIEXPRESS_APP_KEY` | — | Ключ AliExpress Affiliate API |
| `ALIEXPRESS_APP_SECRET` | — | Секрет AliExpress Affiliate API |

## VK интеграция

Чтобы VK-интеграция заработала, нужно:

1. Создать VK-приложение на [vk.com/apps?act=manage](https://vk.com/apps?act=manage) (тип: Веб-сайт), добавить redirect URI.
2. Задать переменные бэкенда:
   ```
   VK_ENABLED=true
   VK_TOKEN_ENCRYPTION_KEY=<openssl rand -base64 32>
   ```
3. Задать переменные фронтенда:
   ```
   VITE_VK_APP_ID=<id приложения>
   VITE_VK_REDIRECT_URI=https://giftsuggestion.ru/auth/vk-callback
   ```

После этого в профиле появится кнопка «Войти через VK», а синхронизация интересов будет вызывать реальный `groups.get` VK API.

## HTTPS на проде (Caddy + Let's Encrypt)

Продакшен-деплой поднимает **Caddy** перед `frontend` (TLS автоматически):

- [`deploy/Caddyfile`](deploy/Caddyfile) — домен из `CADDY_DOMAIN`
- [`deploy/docker-compose.https.yml`](deploy/docker-compose.https.yml) — порты 80/443, `frontend` только внутри сети Docker

На сервере (после `git pull`):

```bash
docker compose -p gift-suggestion --env-file .deploy.env \
  -f docker-compose.yml -f deploy/docker-compose.https.yml up -d
```

В [`deploy/runtime.env`](deploy/runtime.env) / секрете `DEPLOY_RUNTIME_ENV_FILE`:

- `CADDY_DOMAIN=giftsuggestion.ru`
- `ACME_EMAIL=` email для Let's Encrypt
- `FRONTEND_BASE_URL=https://giftsuggestion.ru`
- `AUTH_REFRESH_COOKIE_SECURE=true`
- `VITE_VK_REDIRECT_URI=https://giftsuggestion.ru/auth/vk-callback`

DNS: A-запись `giftsuggestion.ru` → IP сервера; в Yandex Cloud открыты порты **80** и **443**.

Локально без TLS: `docker compose up` (без `docker-compose.https.yml`) — фронт на `:80`.

## CI/CD

GitHub Actions:
- `CI` — backend tests/lint/build, frontend install/generate/lint/build, ML tests (`pytest`), docker compose smoke
- `CD` — сборка и публикация образов backend / frontend / ml-service, deploy через SSH + Caddy HTTPS

В GitHub Environment **production**:

| Тип | Имя | Пример |
|-----|-----|--------|
| Variable | `DEPLOY_HOST` | `giftsuggestion.ru` |
| Variable | `VK_APP_ID` | ID приложения VK |
| Variable | `VK_REDIRECT_URI` | `https://giftsuggestion.ru/auth/vk-callback` |
| Variable | `ACME_EMAIL` | email для Let's Encrypt |
| Secret | `DEPLOY_SSH_PRIVATE_KEY` | SSH-ключ |
| Secret | `DEPLOY_RUNTIME_ENV_FILE` | содержимое `deploy/runtime.env` |

## Текущие ограничения

- **ML модель не обучена** — сервис использует echo-stub (кандидаты в исходном порядке). Обучи через `services/ml/Guide.md`.
- **Email** отключён по умолчанию (`EMAIL_ENABLED=false`). Регистрация и сброс пароля работают через noop-отправитель; письма не доходят до получателя без настройки SMTP.
- **VK** отключён по умолчанию (`VK_ENABLED=false`). Требует ручной настройки переменных — см. раздел выше.
- **Данные каталога** — `dataset_example.csv` синтетический. Для реальных данных используй `scripts/data/`.
