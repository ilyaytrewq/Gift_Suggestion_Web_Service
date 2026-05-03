# Руководство по обучению ML модели

## Требования

```bash
pip install lightgbm pandas pyarrow scikit-learn grpcio grpcio-health-checking protobuf pyyaml
```

## Шаг 1. Подготовка каталога

Сначала нужны реальные данные. Если хочешь использовать синтетику — пропусти этот шаг.

```bash
# Выгрузка из Wildberries (реальные данные)
python scripts/data/fetch_wb.py --query "подарок" --pages 20 --output /tmp/wb.csv

# Нормализация
python scripts/data/normalize_to_catalog.py /tmp/wb.csv --output services/ml/dataset/catalog_seed.csv
```

Или используй готовый `services/ml/dataset/dataset_example.csv` — 10 000 синтетических товаров.

## Шаг 2. Генерация обучающего датасета

```bash
cd services/ml

python3 -m training.prepare_dataset \
  --catalog dataset/dataset_example.csv \
  --queries 2000 \
  --output dataset/training.parquet \
  --seed 42
```

Ожидаемый вывод:
```
INFO Готово: 100000 строк → dataset/training.parquet
INFO Label distribution: {2: ..., 1: ..., 0: ..., 3: ...}
```

Параметр `--queries` задаёт количество синтетических запросов (каждый содержит ~50 кандидатов).
При `--queries 2000` получится ~100 000 строк — достаточно для обучения.

## Шаг 3. Обучение LightGBM

```bash
cd services/ml

python3 -m training.train_lightgbm \
  --dataset dataset/training.parquet \
  --out models/lightgbm_v0_1_0.pkl \
  --metrics models/metrics.json \
  --seed 42
```

Ожидаемое время: 1–3 минуты на 100k строк.

Ожидаемые метрики на синтетике:
```
NDCG@5  = 0.72–0.85
NDCG@10 = 0.70–0.82
```

Файлы создаются:
- `models/lightgbm_v0_1_0.pkl` — обученная модель
- `models/metrics.json` — метрики валидации

## Шаг 4. Проверка метрик

```bash
python3 -m training.eval \
  --model models/lightgbm_v0_1_0.pkl \
  --dataset dataset/training.parquet
```

## Шаг 5. Запуск ML сервиса с моделью

```bash
ML_MODEL_PATH=models/lightgbm_v0_1_0.pkl python3 -m ml_service.main
```

Или через Docker Compose (модель должна быть в `services/ml/models/`):

```bash
docker compose up ml-service
```

Здоровье сервиса:
```bash
curl http://localhost:8081/healthz
# {"status": "ok"}
```

gRPC health (требует grpc_health_probe):
```bash
grpc_health_probe -addr=localhost:50051
# status: SERVING
```

## Переобучение на реальных данных

Когда накопится трекинг из БД (просмотры, добавления в избранное), используй `popularity.parquet`:

```bash
# Обновить popularity.parquet из БД
python3 -m jobs.refresh_popularity --dsn "postgres://..."

# Переобучить с актуальными данными
python3 -m training.prepare_dataset \
  --catalog dataset/catalog_seed.csv \
  --queries 5000 \
  --output dataset/training_v2.parquet

python3 -m training.train_lightgbm \
  --dataset dataset/training_v2.parquet \
  --out models/lightgbm_v0_2_0.pkl \
  --metrics models/metrics_v2.json
```

Обновить используемую модель без перезапуска контейнера:

```bash
ML_MODEL_PATH=models/lightgbm_v0_2_0.pkl docker compose restart ml-service
```

## Структура признаков

| Признак | Описание |
|---------|----------|
| `price_cents` | Цена товара в копейках |
| `price_bucket` | Ценовой бакет [0–7] |
| `budget_cents` | Бюджет запроса в копейках |
| `budget_bucket` | Бакет бюджета [0–7] |
| `price_within_budget` | 1, если цена ≤ бюджет |
| `budget_gap` | бюджет − цена (≥ 0) |
| `budget_ratio` | цена / бюджет, clip(5.0) |
| `age_fits` | 1, если возраст получателя в допустимом диапазоне |
| `interest_overlap` | Кол-во совпадений интересов с текстом товара |
| `tfidf_score` | TF-IDF-косинус интересов и текста товара |
| `category_index` | Индекс категории в INTERNAL_CATEGORIES [0–9] или −1 |
| `already_in_wishlist` | 1, если товар уже в списке желаний пользователя |
| `gender_encoded` | male→1, female→2, other→0, не указан→−1 |

Канонический список определён в `ml_service/feature_builder.py` (константа `FEATURE_COLS`).
`training/train_lightgbm.py` и `training/prepare_dataset.py` должны использовать
тот же список — иначе обученная модель не сможет работать на инференсе.

## Метрика качества

**NDCG@5** (Normalized Discounted Cumulative Gain) — основная метрика. Значение ≥ 0.70 считается
приемлемым для холодного старта. На реальных данных после накопления трекинга ожидается 0.75–0.85.
