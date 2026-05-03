# План реализации Gift Suggestion Web Service

Документ описывает работы, необходимые для приведения текущего состояния репозитория в соответствие с ТЗ (`TP/ТЗ_Тихонов_Илья-2.pdf`). План разбит на фазы с приоритизацией; каждая задача содержит список затрагиваемых файлов и критерии приёмки.

Ссылки в формате `path:line` указывают на текущее место в коде, где нужно внести изменения или которое следует использовать как точку отсчёта.

---

## 0. Сводка статуса

| Блок | Состояние | Что осталось |
|---|---|---|
| Auth (register/login/refresh/logout/reset) | Бэкенд готов, UI частично | Confirm reset/verify в UI |
| Каталог | Бэкенд + UI | Мультимагазинность |
| Wishlist | Бэкенд + UI | — |
| Импорт каталога | Бэкенд готов | Админ-UI; реальные источники данных |
| Recommendation бэкенд | Скелет + fallback | Реальный ML, недостающие поля анкеты, "альтернативы" эндпоинт |
| **ML сервис** | **Не существует** | **Создать с нуля (Python/FastAPI/gRPC + LightGBM/CatBoost)** |
| **Данные подарков (RU/CN маркетплейсы)** | **Только синтетический CSV** | **Парсеры/выгрузки + ETL** |
| Tracking | Бэкенд готов | Шлёт ли фронт события |
| VK интеграция | Scaffold | Реальный VK OAuth + API + миграция в Wizard |
| Frontend Wizard | Одна форма | Пошаговый Wizard, фильтры результатов, alternatives, multistore card, multiselect интересов, ProfilePage в роутере |
| Tests | Юнит-тесты есть, coverage не измеряется | Покрытие ≥ 50%, e2e |
| Документация ГОСТ 19 | Только README | ПЗ, ПиМИ, Текст программы, Руководство оператора |

---

## 1. Архитектурные решения

### 1.1. Источники данных RU/CN маркетплейсов

ТЗ говорит про "открытые данные маркетплейсов и/или готовые датасеты". Так как текущий `services/ml/dataset/dataset_example.csv` — синтетика, нужен реальный набор. Целевые источники:

**RU**
- **Ozon Seller API** — нужна авторизация продавца, не общий каталог. Отпадает для агрегатора.
- **Wildberries** — публичный поисковый эндпоинт `https://search.wb.ru/exactmatch/ru/common/v4/search?query=...` отдаёт JSON со списком товаров (id, name, brand, priceU, image). Используется без авторизации, но без формального SLA. Подходит как источник для seed-датасета.
- **Yandex Market Partner API** — закрыт для внешних разработчиков; парсинг HTML market.yandex.ru возможен, но нестабилен.
- **Megamarket / СберМегаМаркет** — открытого API нет; HTML-парсинг.

**CN**
- **AliExpress Affiliate API (через AliExpress Open Platform)** — официальный, требует регистрации в Aliexpress Portals; отдаёт product feed, цены, картинки, ссылки `aff` для трекинга.
- **Pinduoduo Open Platform** — официальное API товаров, требует регистрации.
- **Taobao Open Platform / TMall** — закрытое для нерезидентов.
- **DataYes / Kaggle datasets** — готовые выгрузки `aliexpress products dataset`, `taobao products dataset` (50–200k строк) для оффлайн-обучения.

**Решение для MVP:**
1. Для прода каталога — **Wildberries** (RU, без ключей) + **AliExpress Affiliate** (CN, с регистрацией ключа).
2. Для обучения ML — **готовые датасеты с Kaggle/HuggingFace** + дообогащение из live-источников.
3. Для дев/демо — оставляем `dataset_example.csv` как fixture, но добавляем второй CSV с **реальными выгрузками 1000+ товаров** через скрипт `scripts/data/fetch_wb.py` и `scripts/data/fetch_aliexpress.py`.

### 1.2. Расширение доменной модели

Добавляется сущность `gift_offers`: один `Gift` ↔ many `Offer` (магазин, URL, цена, валюта, наличие). Текущая `gifts` остаётся "канонической" карточкой, `offers` — покупка в конкретном маркетплейсе.

### 1.3. ML pipeline

```
[Dataset] -> features.parquet -> train.py (LightGBM/LambdaMART) -> model.pkl
                                                                       |
                                                                       v
[Backend gRPC RankRequest] -> [ML FastAPI+gRPC] -> infer (model.pkl) -> RankResponse
```

- **Метка** для LTR: рейтинг релевантности `[0..3]`, синтезированный из (а) совпадение категории с интересом, (б) попадание в бюджет, (в) совпадение возраста, (г) вес из tracking-истории (CTR по category × interest).
- **Признаки кандидата**: `category_id`, `price_bucket`, `age_min/age_max`, `popularity_score`, `text_match_score` (косинус TF-IDF между интересами и `name+description`).
- **Признаки запроса**: бюджет, возраст, повод (one-hot), интересы (multi-hot из словаря).
- **Cross-features**: `price_within_budget`, `age_fits`, `category_in_preferred`, `interest_overlap`.

---

## 2. Фаза 1. Данные подарков (RU + CN)

### 2.1. Скрипты выгрузки

**Создать**:
- `scripts/data/fetch_wb.py` — параметры: `--query "подарок женщине"`, `--pages N`, `--output CSV`. Использует `https://search.wb.ru/exactmatch/ru/common/v4/search` (документировать неофициальность). Поля: `id`, `name` (`product.name`), `description` (пусто, обогащается с product page), `category` (mapping из `subjectId`), `price` (`priceU/100`), `currency=RUB`, `store_url` (`https://www.wildberries.ru/catalog/{id}/detail.aspx`), `image_url` (`https://images.wbstatic.net/c246x328/new/{id/10000}0000/{id}-1.jpg`). Дописать ratelimit (1 запрос/сек) и retry на 429.
- `scripts/data/fetch_aliexpress.py` — через AliExpress Affiliate API (`aliexpress.affiliate.product.query`). Требует `APP_KEY`, `APP_SECRET` в `.env`. Поля маппятся в формат `gift_id=AE{productId}`, `currency=CNY`, `store_url` = `promotion_link`.
- `scripts/data/normalize_to_catalog.py` — приводит выгруженные CSV к схеме каталога (`services/backend/internal/modules/catalogimport/usecase/dto.go`). Маппит произвольные категории маркетплейсов в 10 внутренних:
  - `Электроника`, `Книги`, `Настольные игры`, `Косметика`, `Аксессуары`, `Одежда`, `Хобби`, `Украшения для дома`, `Еда и напитки`, `Детские товары`.
  - Маппинг — словарь в `scripts/data/category_mapping.yaml`.

**Acceptance:**
- `python scripts/data/fetch_wb.py --query "подарок мужчине" --pages 5 --output wb.csv` создаёт CSV с ≥ 500 строками
- `python scripts/data/normalize_to_catalog.py wb.csv ali.csv --output catalog_seed.csv` создаёт ≥ 1000 строк, валидируемый текущим `catalogimport`-парсером (`services/backend/internal/modules/catalogimport/infra/parser/csv.go`)
- В CSV нет пустых обязательных полей, цены > 0, store_url начинается с `https://`

### 2.2. Загрузка в БД через существующий импорт

- Запустить backend локально, через POST `/api/v1/admin/import-jobs` (handler `services/backend/internal/modules/catalogimport/delivery/http/handler.go`) загрузить `catalog_seed.csv`.
- Зафиксировать seed-CSV в `services/ml/dataset/catalog_seed.csv` (заменив или дополнив `dataset_example.csv`).
- Описать процесс в `docs/data.md`.

### 2.3. Расширение схемы под мультимагазинность (4.1.6 п.4.3 и "ссылки на разные магазины" из 6.3 ТЗ)

**Миграция `000014_create_gift_offers.up.sql`**:
```sql
CREATE TABLE gift_offers (
    id           UUID PRIMARY KEY,
    gift_id      UUID NOT NULL REFERENCES gifts(id) ON DELETE CASCADE,
    store_name   TEXT NOT NULL,
    store_url    TEXT NOT NULL,
    price_cents  BIGINT NOT NULL,
    currency     TEXT NOT NULL,
    available    BOOLEAN NOT NULL DEFAULT TRUE,
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (gift_id, store_url)
);
CREATE INDEX gift_offers_gift_id_idx ON gift_offers (gift_id);
```

**Доменная модель**:
- Добавить `services/backend/internal/modules/catalog/domain/offer.go` с `Offer{ID, GiftID, StoreName, StoreURL, Price, Available, FetchedAt}` и фабрикой.
- Расширить `Gift` агрегат: `Offers() []Offer`. Хранить как поле, но грузить лениво через репозиторий (`SELECT ... FROM gift_offers WHERE gift_id = ANY(...)`).

**Импорт**: расширить `services/backend/internal/modules/catalogimport/usecase/normalize.go` — если в строке файла указан массив магазинов (новое опциональное поле `offers` в JSON или повторяющиеся `store_url_2`, `store_url_3` в CSV), создавать соответствующие `gift_offers`.

**API**: в OpenAPI (`services/backend/docs/openapi/backend.yaml`) расширить `GiftCard` поле `offers: array<Offer>`. Frontend (`services/frontend/src/shared/api/generated/schema.ts`) перегенерируется через `npm run generate:api`.

**Acceptance:**
- Импорт CSV с двумя строками одного `gift_id` создаёт две offer-записи
- `GET /api/v1/catalog/gifts/{id}` возвращает массив `offers`
- Юнит-тесты в `services/backend/internal/modules/catalog/usecase/service_test.go` покрывают чтение оффер

---

## 3. Фаза 2. ML сервис

### 3.1. Структура сервиса

Создать `services/ml/`:
```
services/ml/
├── Dockerfile
├── pyproject.toml          # poetry/uv project
├── README.md
├── proto/
│   └── ranking.proto       # синхронизирован с backend
├── ml_service/
│   ├── __init__.py
│   ├── main.py             # FastAPI с /healthz + gRPC сервер на 50051
│   ├── grpc_server.py      # реализация Ranking сервиса
│   ├── feature_builder.py  # из RankRequest -> pandas.DataFrame
│   ├── ranker.py           # загрузка модели и predict
│   ├── explainer.py        # SHAP / правила -> Explanation[]
│   └── config.py           # pydantic-settings
├── training/
│   ├── prepare_dataset.py  # parquet из CSV каталога + tracking events
│   ├── train_lightgbm.py
│   ├── train_catboost.py
│   ├── eval.py             # NDCG@5, NDCG@10
│   └── synth_labels.py     # синтез релевантности до накопления tracking
├── models/
│   └── .gitkeep            # сюда сохраняются model.pkl
└── tests/
    ├── test_feature_builder.py
    ├── test_ranker.py
    └── test_grpc_server.py
```

### 3.2. gRPC контракт

**Создать `services/ml/proto/ranking.proto`** (синхронизировать с типами в `services/backend/internal/platform/mlgrpc/client.go:16-57`):
```proto
syntax = "proto3";
package gift.ranking.v1;

service RankingService {
  rpc Rank(RankRequest) returns (RankResponse);
}

message QueryContext {
  string occasion     = 1;
  string relationship = 2;
  repeated string interests = 3;
  int64 budget_cents  = 4;
  optional int32 recipient_age = 5;
  optional string gender = 6;          // добавляется в Фазе 4
}

message Candidate {
  string id              = 1;
  optional string category_id = 2;
  string category_name   = 3;
  int64 price_cents      = 4;
  optional int32 age_restriction = 5;
  string title           = 6;
  string description     = 7;
  bool already_in_wishlist = 8;
}

message RankRequest {
  string selection_id = 1;
  string user_id      = 2;
  int32 top_n         = 3;
  QueryContext query  = 4;
  repeated Candidate candidates = 5;
}

message Explanation {
  string code = 1;
  string text = 2;
}

message RankedItem {
  string candidate_id = 1;
  double score        = 2;
  repeated Explanation explanations = 3;
  repeated string alternative_candidate_ids = 4;
}

message RankResponse {
  repeated RankedItem items = 1;
  string model_version = 2;
}
```

**Backend**:
- Добавить `services/backend/api/proto/ranking.proto` (mirror) и `services/backend/internal/platform/mlgrpc/gen/` через `protoc-gen-go` + `protoc-gen-go-grpc`.
- Заменить заглушку `RankCandidates` (`services/backend/internal/platform/mlgrpc/client.go:123`) на реальный gRPC-вызов с deadline = `cfg.RequestTimeout`, retry до `cfg.MaxRetries`. Маппинг ошибок: `codes.Unavailable` → `ErrRankingUnavailable`, `codes.DeadlineExceeded` → `ErrRankingTimedOut`, `codes.Unimplemented` → `ErrRankingNotImplemented`, `codes.InvalidArgument`/невалидный response → `ErrInvalidRankingResponse`.
- Подключить `protoc` шаги в `services/backend/Taskfile.yaml` (`task gen:proto`).

### 3.3. Feature engineering

`services/ml/ml_service/feature_builder.py`:

```python
def build_features(request: RankRequest) -> pd.DataFrame:
    rows = []
    for c in request.candidates:
        rows.append({
            "candidate_id": c.id,
            "price_cents": c.price_cents,
            "price_within_budget": c.price_cents <= request.query.budget_cents,
            "budget_gap": request.query.budget_cents - c.price_cents,
            "category_in_preferred": c.category_id in PREFERRED,
            "age_fits": _age_fits(c, request.query.recipient_age),
            "interest_overlap": _interest_overlap(c, request.query.interests),
            "tfidf_match": _tfidf(c, request.query.interests, request.query.occasion),
            "popularity": _popularity(c.id),  # из таблицы tracking_events
            "already_in_wishlist": c.already_in_wishlist,
        })
    return pd.DataFrame(rows)
```

`popularity` хранится в Parquet `services/ml/data/popularity.parquet`, обновляется батч-задачей раз в сутки (см. 3.6).

### 3.4. Обучение

`services/ml/training/train_lightgbm.py`:
- Читает `services/ml/dataset/training.parquet` (созданный `prepare_dataset.py`).
- Группировка `query_id` (одна строка анкеты = одна группа).
- LightGBM с параметрами `objective="lambdarank"`, `metric="ndcg"`, `ndcg_eval_at=[5,10]`, `learning_rate=0.05`, `n_estimators=400`, `early_stopping=30`.
- Сохраняет `services/ml/models/lightgbm_v{semver}.pkl` + `metrics.json` (NDCG@5, NDCG@10, обучающие/валидационные).

`services/ml/training/synth_labels.py`:
- На холодный старт (нет tracking данных) — синтезирует целевую релевантность по правилам (бюджет = 0 если выходит за рамки, +1 за категорию, +1 за совпадение интересов, +1 за попадание возраста). Порядковая шкала [0..3].

`services/ml/training/eval.py`:
- На отложенной выборке считает `NDCG@5`, `NDCG@10`, `MAP`, `coverage` (доля категорий в топ-5).

**Acceptance:**
- `python -m training.train_lightgbm --dataset services/ml/dataset/training.parquet --out services/ml/models/lightgbm_v0_1_0.pkl` сохраняет модель и `metrics.json`
- На отложенной выборке `NDCG@5 ≥ 0.7` (синтетика, для холодного старта)

### 3.5. gRPC server

`services/ml/ml_service/grpc_server.py`:
- `grpc.aio.server` слушает `:50051`
- Имплементирует `RankingService.Rank`:
  1. Загрузка модели — singleton с `lru_cache`.
  2. `feature_builder.build_features(request)`.
  3. `model.predict(features)`.
  4. Top-N сортировка, расчёт `alternative_candidate_ids` (соседи по cosine similarity TF-IDF).
  5. `explainer.explain(candidate, features)` — для топ-N генерирует список объяснений на основе SHAP-значений или правил.
- Регистрирует `grpc_health_v1.HealthService`.
- Ответ ≤ 6 секунд для топ-200 кандидатов (бенчмарк в `tests/test_grpc_server.py`).

### 3.6. Cron-задача обновления popularity

**Создать `services/ml/jobs/refresh_popularity.py`**: читает `tracking_events` из БД (через psycopg, та же DSN), агрегирует за 30 дней по gift_id, складывает в `popularity.parquet`. Запуск: cron каждые 24h в `docker-compose.yml` (через `restart: on-failure` + `entrypoint` со sleep, либо отдельный `ml-jobs` сервис).

### 3.7. Docker и compose

**`services/ml/Dockerfile`**:
- base `python:3.13-slim`
- copy poetry/uv lock, install deps
- copy `ml_service/`
- ENTRYPOINT: `python -m ml_service.main`

**`docker-compose.yml`** добавить сервис:
```yaml
ml-service:
  build:
    context: .
    dockerfile: services/ml/Dockerfile
  environment:
    ML_GRPC_ADDR: ${ML_GRPC_ADDR:-0.0.0.0:50051}
    ML_MODEL_PATH: ${ML_MODEL_PATH:-/app/models/lightgbm_v0_1_0.pkl}
    ML_DB_DSN: ${DB_DSN}
  ports:
    - "50051:50051"
  depends_on:
    postgres:
      condition: service_healthy
  healthcheck:
    test: ["CMD", "grpc_health_probe", "-addr=:50051"]
    interval: 10s
    retries: 5
```

В `backend` сервисе включить `ML_GRPC_ENABLED=true` по умолчанию (`docker-compose.yml`).

**Acceptance:**
- `docker compose up ml-service` поднимает сервис
- `grpc_health_probe -addr=localhost:50051` отвечает SERVING
- Backend в `Recommend` использует ML, а не fallback (`recommendation.ranking_source = "ml"` в результате)

---

## 4. Фаза 3. Backend gaps

### 4.1. Расширение анкеты (поле "пол")

ТЗ 4.1.1 п.1.1 требует "пол получателя", а сейчас в `Questionnaire` его нет.

**Изменения**:
- `services/backend/internal/modules/recommendation/domain/questionnaire.go` — добавить `RecipientGender *string` (`male`/`female`/`other`) с валидацией.
- `services/backend/internal/modules/recommendation/usecase/dto.go` — `RecommendInput.RecipientGender`.
- Миграция `000015_add_gender_to_recommendation_requests.up.sql` — `ALTER TABLE recommendation_requests ADD COLUMN recipient_gender TEXT NULL`.
- OpenAPI `RecommendationRequest.recipient_gender`.
- Перегенерировать TS типы.
- Расширить `RankRequest.QueryContext` proto и mapping в gateway (`services/backend/internal/modules/recommendation/infra/grpc/`).
- Учитывать `gender` в `feature_builder.py` (one-hot encoding).

### 4.2. Endpoint "похожие подарки"

ТЗ 3.1 включает "Поиск похожих подарков и альтернатив". Текущие альтернативы возвращаются только в составе recommendation-запроса. Нужен публичный поиск.

**Создать `GET /api/v1/catalog/gifts/{gift_id}/similar?limit=N`**:
- handler в `services/backend/internal/modules/catalog/delivery/http/handler.go`
- usecase `Similar(ctx, giftID, limit) []Gift` — простая реализация: те же категория + тот же ценовой бакет ±20%, исключая исходный, сортировка по совпадению age_restriction.
- В перспективе — вызов ML с `mode="similar"` (новое RPC `RankSimilar`), но MVP — на правилах.

### 4.3. VK интеграция: реальная реализация

Заменить `services/backend/internal/modules/vkintegration/infra/vk/client.go:23-29` на реальное обращение к VK API:

- VK OAuth flow (Authorization Code Flow):
  - Frontend редирект на `https://oauth.vk.com/authorize?client_id=...&redirect_uri=...&scope=friends,groups&response_type=code&v=5.131`
  - Backend `POST /api/v1/integrations/vk/oauth/callback` принимает `code`, обменивает на `access_token` через `https://oauth.vk.com/access_token`
  - Токен шифруется существующим `vkintegrationcrypto.AESGCMProtector` и сохраняется в БД (`vk_connections`)
- VK API запросы:
  - `groups.get?extended=1&fields=activity` — даёт сообщества с категорией активности
  - `users.get?fields=interests,books,music,games,movies,about` — текстовые поля
- Маппинг в внутренние интересы:
  - Словарь `services/backend/internal/modules/vkintegration/usecase/interest_dictionary.go` (yaml), маппинг `activity → interest_tag`
- `ImportInterests(ctx, request)` возвращает `ImportInterestsResult{Interests: []domain.ImportedInterest{...}}` уже на правильном языке внутреннего словаря.

**Конфиг**: добавить `VK_APP_ID`, `VK_APP_SECRET`, `VK_REDIRECT_URI` в `services/backend/internal/platform/config/config.go` и `.env.example`.

**Acceptance:**
- При `VK_ENABLED=true` и валидных кредах: реальный OAuth flow в браузере, в БД появляется `vk_connection`, `imported_interests` заполняются на основе групп/интересов пользователя
- При интеграции отключённой — feature-flag по-прежнему 503 без падений
- Тесты с моком HTTP клиента (httptest.Server) в `services/backend/internal/modules/vkintegration/usecase/service_test.go`

### 4.4. HTTPS / прод

ТЗ 4.4.4 требует HTTPS. Добавить:
- `deploy/docker-compose.deploy.yml` — service `caddy` (или `traefik`) с автоматическим Let's Encrypt
- Переменные `DOMAIN_NAME`, `LETSENCRYPT_EMAIL` в `deploy/runtime.env.example`
- Документировать в `docs/ci-cd.md`

### 4.5. Логирование и метрики

ТЗ 4.7 п.2 — "все критичные операции". Уже есть `slog` через `services/backend/internal/platform/logger/logger.go`. Дополнительно:
- В `recommendation/usecase/service.go` — структурированные логи на каждый Recommend-запрос: `request_id, user_id, candidates_total, candidates_after_filters, ranking_source, fallback_used, fallback_reason, latency_ms`
- Аналогично в `tracking`, `vkintegration`, `catalogimport`
- Опционально — экспорт в OpenTelemetry (Prometheus metrics: `recommendation_latency_seconds`, `ml_call_failures_total`).

---

## 5. Фаза 4. Frontend gaps

### 5.1. Подключение ProfilePage в роутер

Файл `services/frontend/src/pages/profile/ui/profile-page.tsx` не используется. В `services/frontend/src/app/router/router.tsx:32` добавить:
```tsx
<Route element={<ProfilePage />} path="/profile" />
```
В `app-shell` (`services/frontend/src/shared/ui/layout/app-shell.tsx`) добавить ссылку `/profile` в меню для авторизованных.

### 5.2. Wizard рекомендаций (4.1.6 п.2.1, 2.2)

Сейчас `services/frontend/src/pages/recommendation/ui/recommendation-page.tsx` — единая форма. Превратить в пошаговый Wizard:
- Шаги: `Повод → Бюджет → Получатель (пол, возраст, отношения) → Интересы (multiselect) → Превью`
- Кнопки `Назад`, `Пропустить`, `Дальше`
- Состояние шагов в `useState` или `react-hook-form` с `formContext`
- Валидация только при попытке перейти/отправить

**Multiselect интересов**:
- Заменить comma-separated input на `<MultiselectField options={...} />`
- Источник опций: новый endpoint `GET /api/v1/catalog/interests` (возвращает 20–30 предопределённых тегов) — добавить в backend (миграция таблицы `interests` или статичный yaml `services/backend/internal/modules/catalog/usecase/interests.go`)

**Acceptance:**
- На Wizard видна индикация текущего шага (1/5)
- Можно вернуться назад без потери введённого
- Можно пропустить опциональный шаг

### 5.3. Страница результатов рекомендаций

Сейчас результаты выводятся на той же странице. Расширения:
- Карточка показывает `score`, объяснения (`explanations[].text`) — фрагменты "Почему это подходит"
- Блок `Альтернативы` под каждым основным результатом
- Фильтры (4.1.6 п.3.3): уточнить ценовой диапазон, категории — клиентская перезагрузка с новым `POST /recommendations`
- Сохранение в `wishlist` (кнопка `WishlistSaveButton` уже есть)

### 5.4. Карточка подарка с мультимагазинностью

`services/frontend/src/pages/gift/ui/gift-page.tsx`:
- Список магазинов: рендер `gift.offers[]` (см. 2.3)
- Кнопка `Купить` — переход на `offers[i].store_url` с `target="_blank"` и обязательной отправкой tracking-события `gift.link_clicked`
- Блок `Альтернативы` — fetch `GET /api/v1/catalog/gifts/{id}/similar?limit=4`

### 5.5. Tracking события из фронта

Создать `services/frontend/src/shared/tracking/tracker.ts`:
```ts
export function trackEvent(type: 'gift_view' | 'wishlist_add' | 'link_click', payload: Record<string, unknown>) {
  void httpClient.post('/api/v1/tracking/events', { type, payload, occurred_at: new Date().toISOString() });
}
```
Интеграция:
- На `GiftPage` mount → `trackEvent('gift_view', { gift_id })`
- На `WishlistSaveButton.onClick` → `trackEvent('wishlist_add', { gift_id })`
- На `Buy` button → `trackEvent('link_click', { gift_id, store_url })`

### 5.6. UI confirm flows для auth

- `services/frontend/src/pages/auth/ui/password-reset-confirm-page.tsx` — принимает `?token=...`, форма `new_password + confirm`, POST `/api/v1/auth/password-reset/confirm`
- `services/frontend/src/pages/auth/ui/email-verification-page.tsx` — принимает `?token=...`, на mount делает POST `/api/v1/auth/email-verification/confirm`, показывает success/error
- Подключить в роутере (`router.tsx`): `/password-reset/confirm`, `/email-verify`
- Backend уже отправляет письма со ссылками на `FRONTEND_BASE_URL` — обновить шаблоны писем (`services/backend/internal/modules/auth/infra/email/`) на новые URL

### 5.7. Admin import UI

Создать `services/frontend/src/pages/admin/import/ui/import-page.tsx`:
- Защищён ролью `admin` (использовать `user.role` из `/api/v1/users/me` — добавить в JWT claim)
- Форма загрузки файла (drag&drop) с выбором формата
- POST на `/api/v1/admin/import-jobs`
- Список последних `import_jobs`, статус, errors (пагинация 50)
- Раскрытие ошибок — `GET /api/v1/admin/import-jobs/{id}/errors`

### 5.8. VK интеграция в UI

В Profile или отдельной странице `/integrations/vk`:
- Кнопка "Подключить VK" → редирект на `oauth.vk.com/authorize?...`
- Callback `services/frontend/src/pages/integrations/vk/ui/vk-callback-page.tsx` принимает `?code=...`, отправляет на `POST /api/v1/integrations/vk/oauth/callback`
- После успеха показывает список импортированных интересов, кнопка `Синхронизировать снова`
- Кнопка `Отключить VK` → `DELETE /api/v1/integrations/vk/connection`

---

## 6. Фаза 5. Тесты, наблюдаемость, документация

### 6.1. Покрытие тестами ≥ 50% (ТЗ 4.7 п.3)

**Backend**:
- Сейчас тесты есть на usecase, парсеры, ряд handlers. Цель — поднять `go test -cover ./...` до 50%+.
- Добавить:
  - `services/backend/internal/modules/recommendation/usecase/service_test.go` — кейсы: ML success, ML timeout → fallback, ML partial response, нулевые кандидаты после фильтров
  - `services/backend/internal/modules/catalog/infra/postgres/*_test.go` — sqlx моки/`testcontainers-go` postgres
  - `services/backend/internal/platform/mlgrpc/client_test.go` — gRPC через `bufconn`

**ML**:
- pytest на `feature_builder`, `ranker`, `grpc_server`
- coverage в CI (`pytest --cov=ml_service --cov-fail-under=70`)

**Frontend**:
- vitest + Testing Library на критичные компоненты (Wizard, RecommendationResults, ImportPage)
- e2e через Playwright: happy path register → wizard → сохранение в wishlist

CI (`/.github/workflows/ci.yml`) обновить для запуска coverage и публикации артефактов.

### 6.2. Документация ГОСТ 19 (ТЗ 5.1)

Создать в папке `TP/`:
- `Пояснительная_записка.docx` (ГОСТ 19.404-79)
- `Программа_и_методика_испытаний.docx` (ГОСТ 19.301-79)
- `Текст_программы.docx` (ГОСТ 19.401-78) — обновить существующий `TP/Текст_программы_Тихонов_Илья-1.pdf`
- `Руководство_оператора.docx` (ГОСТ 19.505-79)

Содержание ПЗ — описание архитектуры (диаграммы из `docs/architecture.md` — новый файл), ML-подход, источники данных, метрики, тесты.

### 6.3. CI/CD расширения

`/.github/workflows/ci.yml`:
- Job `ml-tests`: `cd services/ml && uv sync && uv run pytest`
- Job `ml-build`: build Docker image
- Job `e2e`: `npx playwright test` после сборки compose

`/.github/workflows/cd.yml`:
- Билд и push образа `ml-service` рядом с `backend` и `frontend`
- В deploy compose добавлен `ml-service`

### 6.4. README и AGENTS

Обновить (`README.md`, `AGENTS.md`) под новые компоненты:
- Раздел "ML сервис" — как запустить, обучить модель, обновить
- Раздел "Источники данных" — как выгрузить WB и AliExpress
- Обновить "Что реализовано" / "Ограничения текущего состояния"

---

## 7. Фаза 6. Прод-готовность

### 7.1. Безопасность

- Все секреты — только через env / GitHub Secrets, не в repo
- `helmet`/CORS настроены в gin (`services/backend/internal/transport/httpapi/middleware.go`)
- Rate limiting на `/api/v1/auth/*` (например, ipratelimit через middleware)
- CSRF protection для cookie-based refresh: уже есть `SameSite=Lax`, добавить `Secure` в проде

### 7.2. Резервные копии и миграции (ТЗ 4.6)

- Скрипт `scripts/db/backup.sh` (pg_dump в S3-совместимое хранилище)
- Документировать в `docs/operations.md`

### 7.3. Мониторинг

- `/metrics` (Prometheus) на backend и ml-service
- Pre-built Grafana dashboards (json-export коммитится в `deploy/grafana/`)
- Алерт на `recommendation_latency_seconds_p99 > 6` и `ml_call_failures_total` > 5%

---

## 8. Порядок выполнения и оценка

| # | Фаза | Зависимости | Оценка |
|---|---|---|---|
| 1 | Скрипты выгрузки RU/CN + миграция `gift_offers` (Фаза 1) | — | 4–5 дней |
| 2 | Proto-контракт + перегенерация + замена ML-заглушки на gRPC-вызов | Фаза 1 | 1–2 дня |
| 3 | Расширение анкеты (gender), similar endpoint, VK реальный OAuth (Фаза 3) | Фаза 1 | 4–5 дней |
| 4 | ML сервис: dataset prep + LightGBM training + gRPC server (Фаза 2) | #2, #1 | 7–10 дней |
| 5 | Frontend Wizard, multiselect, multistore, alternatives, profile, tracking, admin import, VK UI, confirm flows (Фаза 4) | #2, #3 | 7–9 дней |
| 6 | Тесты до 50%, e2e, наблюдаемость (Фаза 5) | #4, #5 | 4–5 дней |
| 7 | HTTPS / Caddy / монитор / прод-deploy (Фаза 6) | #6 | 2–3 дня |
| 8 | Документация ГОСТ 19 | #1–7 | 3–4 дня |

Итого: ~32–43 рабочих дня.

---

## 9. Критерии готовности (Definition of Done)

Проект считается реализованным, если:

- `docker compose up` поднимает все 4 сервиса (postgres, backend, ml-service, frontend), все healthcheck → SERVING
- На главной странице доступны `Подобрать подарок`, `Каталог идей`, вход/регистрация
- Wizard позволяет пройти все шаги, в т.ч. указать пол и возраст, выбрать интересы из списка
- Результат запроса содержит топ-N карточек с объяснениями + по 2 альтернативы; источник `ranking_source = "ml"`; время ответа ≤ 6 сек на 200 кандидатов
- Карточка подарка показывает фото, цену, описание, **минимум 2 магазина** (из `gift_offers`), блок `Альтернативы`, кнопку `Сохранить в список желаний`
- VK OAuth работает end-to-end в dev-окружении: после подключения в анкете предзаполняются интересы
- Админка импорта позволяет загрузить CSV/JSON/XLSX и видит результат
- В каталоге есть **минимум 1000** товаров с **реальными** ссылками на Wildberries и AliExpress
- `go test -cover ./... | grep total` ≥ 50%; `pytest --cov` ≥ 70% для ML; frontend lint+build зелёные
- Tracking-события приходят на `/api/v1/tracking/events` при просмотре карточки и переходе по ссылке
- Документация ГОСТ 19 (5 документов) подготовлена в `TP/`
- HTTPS настроен на прод-окружении (Caddy/Traefik + Let's Encrypt)

---

## 10. Риски и митигации

| Риск | Митигация |
|---|---|
| Wildberries закроет/изменит публичный поиск | Заранее закэшировать снапшот; добавить второй RU-источник (Megamarket HTML-парсинг) |
| AliExpress Affiliate отклонит регистрацию | Альтернатива: HuggingFace dataset `aliexpress-products`, либо общий китайский каталог через Pinduoduo Open Platform |
| ML модель не достигает NDCG@5 ≥ 0.7 | Начать с rule-based ranker (улучшенный fallback, уже есть в `service.go`) и LightGBM добавить позже как A/B |
| VK закроет API для русских разработчиков | Сделать VK feature-flag и не блокировать релиз, если интеграция оказывается неработоспособной |
| Time-to-deliver overrun | Жёсткая приоритизация: фазы 1–4 — MVP; фазы 5–6 — после защиты |
