# Источники данных каталога подарков

## Структура скриптов

```
scripts/data/
├── fetch_wb.py           — выгрузка из Wildberries
├── fetch_aliexpress.py   — выгрузка из AliExpress Affiliate API
├── normalize_to_catalog.py — нормализация и слияние CSV
└── category_mapping.yaml — маппинг категорий маркетплейсов
```

## Шаг 1. Выгрузка из Wildberries

```bash
python scripts/data/fetch_wb.py \
  --query "подарок" \
  --pages 10 \
  --delay 1.0 \
  --output /tmp/wb_raw.csv
```

Флаг `--mock` используется в CI и возвращает 2 тестовые строки без HTTP запросов.

**Ограничения:** использует неофициальный публичный поиск WB без авторизации. При смене структуры ответа WB скрипт может потребовать правок.

## Шаг 2. Выгрузка из AliExpress

Нужна регистрация в [AliExpress Open Platform](https://portals.aliexpress.com):

```bash
export ALIEXPRESS_APP_KEY="ваш_ключ"
export ALIEXPRESS_APP_SECRET="ваш_секрет"

python scripts/data/fetch_aliexpress.py \
  --keywords "gift" \
  --pages 5 \
  --delay 1.0 \
  --output /tmp/ali_raw.csv
```

Флаг `--mock` работает без ключей:

```bash
python scripts/data/fetch_aliexpress.py --keywords "x" --pages 1 --mock --output /tmp/ali_mock.csv
```

## Шаг 3. Нормализация

```bash
python scripts/data/normalize_to_catalog.py \
  /tmp/wb_raw.csv /tmp/ali_raw.csv \
  --output services/ml/dataset/catalog_seed.csv
```

Скрипт:
- Применяет маппинг категорий из `category_mapping.yaml`
- Удаляет дубликаты по `store_url`
- Пропускает строки с пустыми обязательными полями или невалидной ценой
- Нормализует цену к формату `"NNNN.00"`

## Шаг 4. Загрузка в БД

Через API администратора (нужна авторизация с ролью `admin`):

```bash
curl -X POST http://localhost:8080/api/v1/admin/import-jobs \
  -H "Authorization: Bearer <token>" \
  -F "file=@services/ml/dataset/catalog_seed.csv" \
  -F "format=csv"
```

## Формат файла импорта

| Поле | Обязательное | Описание |
|------|--------------|----------|
| gift_id | да | Уникальный идентификатор (напр. `WB12345`) |
| title | да | Название товара |
| description | нет | Описание |
| category | да | Одна из 10 внутренних категорий |
| price | да | Цена в формате `1234.00` |
| currency | да | `RUB` или `CNY` |
| store_url | да | Ссылка на товар (https://...) |
| image_url | нет | Ссылка на изображение |
| age_min | нет | Минимальный возраст (0 по умолчанию) |
| age_max | нет | Максимальный возраст (99 по умолчанию) |

## Внутренние категории

- Электроника
- Книги
- Настольные игры
- Косметика
- Аксессуары
- Одежда
- Хобби
- Украшения для дома
- Еда и напитки
- Детские товары
