"""Синтез меток релевантности для холодного старта.

Метки [0..3]:
  3 — отлично подходит (в бюджете + совпадение интереса + совпадение возраста + качество)
  2 — хорошо подходит (2 из 3 условий)
  1 — слабое соответствие (в бюджете, но ничего не совпало)
  0 — не подходит (вне бюджета)

Намеренный шум 15 % предотвращает переобучение на детерминированных правилах.
"""
from __future__ import annotations

import argparse
import logging
import random

import pandas as pd

log = logging.getLogger(__name__)

# Расширенный маппинг: русские + английские ключевые слова для Myntra/Online Retail
INTEREST_MAP: dict[str, list[str]] = {
    "Электроника": [
        "гаджет", "смартфон", "наушники", "ноутбук", "электроника",
        "gadget", "phone", "earphone", "laptop", "tech", "electronic", "headphone",
    ],
    "Книги": [
        "книга", "литература", "чтение", "роман", "учеба",
        "book", "novel", "reading", "literature", "fiction",
    ],
    "Настольные игры": [
        "игра", "настольная", "пазл", "шахматы", "хобби",
        "game", "puzzle", "board", "chess", "lego", "cards",
    ],
    "Косметика": [
        "косметика", "уход", "красота", "парфюм", "макияж",
        "beauty", "perfume", "skincare", "makeup", "lipstick", "fragrance", "serum",
    ],
    "Аксессуары": [
        "аксессуар", "сумка", "кошелек", "украшение", "часы",
        "bag", "wallet", "watch", "accessories", "jewelry", "backpack",
        "luggage", "trolley", "handbag", "sunglasses", "bracelet", "necklace",
    ],
    "Одежда": [
        "одежда", "платье", "рубашка", "стиль", "мода",
        "clothing", "dress", "shirt", "fashion", "style", "top", "jeans",
        "jacket", "coat", "sweater", "kurta", "saree",
    ],
    "Хобби": [
        "творчество", "хобби", "рукоделие", "живопись", "музыка",
        "craft", "hobby", "art", "music", "creative", "painting", "guitar",
    ],
    "Украшения для дома": [
        "декор", "интерьер", "свеча", "ваза", "уют",
        "decor", "candle", "home", "interior", "vase", "lamp", "frame",
    ],
    "Еда и напитки": [
        "еда", "напиток", "чай", "кофе", "гастрономия",
        "food", "drink", "tea", "coffee", "beverage", "gourmet", "snack",
    ],
    "Детские товары": [
        "ребенок", "детский", "игрушка", "малыш", "развитие",
        "children", "baby", "kids", "toy", "infant", "child",
    ],
}

# Реалистичные возрастные ограничения по категории
CATEGORY_AGE_RANGE: dict[str, tuple[int, int]] = {
    "Детские товары": (0, 12),
    "Настольные игры": (5, 99),
    "Электроника": (14, 99),
    "Косметика": (16, 99),
    "Еда и напитки": (0, 99),
    "Одежда": (0, 99),
    "Аксессуары": (12, 99),
    "Хобби": (6, 99),
    "Украшения для дома": (18, 99),
    "Книги": (10, 99),
}

OCCASIONS = [
    "день рождения", "новый год", "8 марта", "23 февраля",
    "свадьба", "юбилей", "выпускной", "корпоратив", "крестины", "просто так",
]
RELATIONSHIPS = [
    "друг", "коллега", "родитель", "партнер", "ребенок",
    "брат/сестра", "бабушка/дедушка", "начальник", "учитель",
]
GENDERS = ["male", "female", "other"]

# Дискретные бюджеты в рублях (INR ≈ RUB по грубому паритету ~1.1)
BUDGETS_RUB = [
    500, 800, 1000, 1500, 2000, 2500, 3000, 4000,
    5000, 7000, 10000, 15000, 20000, 30000,
]
AGE_GROUPS = [(0, 12), (13, 17), (18, 25), (26, 35), (36, 50), (51, 70), (71, 99)]


def _age_range_for_category(category: str) -> tuple[int, int]:
    for key, rng in CATEGORY_AGE_RANGE.items():
        if key.lower() in category.lower():
            return rng
    return (0, 99)


def synth_label(row: pd.Series, query: dict, rng: random.Random) -> int:
    # Цена в рублях: INR → RUB (коэфф. ~1.1), GBP → RUB (~105)
    raw_price = float(row.get("price", 0) or 0)
    currency = str(row.get("currency", "RUB")).upper()
    if currency == "INR":
        price_rub = raw_price * 1.1
    elif currency == "GBP":
        price_rub = raw_price * 105
    else:
        price_rub = raw_price

    if price_rub > query["budget"]:
        return 0

    score = 1  # в бюджете

    # Совпадение интереса: проверяем все ключевые слова в title + category
    corpus = (
        str(row.get("title", "")).lower()
        + " " + str(row.get("description", "")).lower()
        + " " + str(row.get("category", "")).lower()
    )
    interests = query.get("interests", [])
    if any(kw in corpus for kw in interests):
        score += 1

    # Возраст: используем реалистичные диапазоны по категории
    category = str(row.get("category", ""))
    age_min, age_max = _age_range_for_category(category)
    age = query.get("recipient_age", 30)
    if age_min <= age <= age_max:
        score += 1

    # Скрытое качество товара (не является фичей модели) — создаёт неустранимый шум
    quality_bonus = rng.gauss(0, 0.4)
    score_f = score + quality_bonus
    label = min(3, max(0, round(score_f)))

    # Случайный шум в метках (15 %) — предотвращает заучивание формулы
    if rng.random() < 0.15:
        label = max(0, min(3, label + rng.choice([-1, 1])))

    return label


def generate_queries(n_queries: int, seed: int = 42) -> list[dict]:
    rng = random.Random(seed)
    queries = []
    for qid in range(n_queries):
        age_group = rng.choice(AGE_GROUPS)
        # Иногда выбираем смешанные интересы из двух категорий
        primary_cat = rng.choice(list(INTEREST_MAP.keys()))
        secondary_cat = rng.choice(list(INTEREST_MAP.keys()))
        primary_kws = rng.sample(INTEREST_MAP[primary_cat], k=min(2, len(INTEREST_MAP[primary_cat])))
        secondary_kws = rng.sample(INTEREST_MAP[secondary_cat], k=1)
        interests = list(dict.fromkeys(primary_kws + secondary_kws))  # дедупликация с сохранением порядка

        # Бюджет: иногда случайный в диапазоне, иногда из фиксированного списка
        if rng.random() < 0.4:
            budget = rng.randint(300, 25000)
        else:
            budget = rng.choice(BUDGETS_RUB)

        queries.append({
            "query_id": qid,
            "occasion": rng.choice(OCCASIONS),
            "relationship": rng.choice(RELATIONSHIPS),
            "budget": budget,
            "recipient_age": rng.randint(*age_group),
            "recipient_gender": rng.choice(GENDERS),
            "interests": interests,
            "preferred_category": primary_cat,
        })
    return queries


def build_training_dataset(catalog_path: str, n_queries: int = 1000, seed: int = 42) -> pd.DataFrame:
    catalog = pd.read_csv(catalog_path)
    catalog["price"] = catalog["price"].astype(str).str.replace(",", ".").astype(float)
    catalog["age_min"] = pd.to_numeric(catalog.get("age_min", 0), errors="coerce").fillna(0).astype(int)
    catalog["age_max"] = pd.to_numeric(catalog.get("age_max", 99), errors="coerce").fillna(99).astype(int)

    queries = generate_queries(n_queries, seed)
    rng = random.Random(seed + 1)

    rows = []
    for query in queries:
        sample_size = min(50, len(catalog))
        candidates = catalog.sample(n=sample_size, random_state=rng.randint(0, 99999))

        for _, gift_row in candidates.iterrows():
            label = synth_label(gift_row, query, rng)

            # Цена в рублях для фичей
            raw_price = float(gift_row.get("price", 0) or 0)
            currency = str(gift_row.get("currency", "RUB")).upper()
            if currency == "INR":
                price_rub = raw_price * 1.1
            elif currency == "GBP":
                price_rub = raw_price * 105
            else:
                price_rub = raw_price

            rows.append({
                "query_id": query["query_id"],
                "gift_id": gift_row.get("gift_id", ""),
                "occasion": query["occasion"],
                "relationship": query["relationship"],
                "budget": query["budget"],
                "recipient_age": query["recipient_age"],
                "recipient_gender": query["recipient_gender"],
                "interests": ",".join(query["interests"]),
                "preferred_category": query["preferred_category"],
                "category": gift_row.get("category", ""),
                "price": price_rub,
                "age_min": int(gift_row["age_min"]),
                "age_max": int(gift_row["age_max"]),
                "title": str(gift_row.get("title", gift_row.get("name", ""))),
                "description": str(gift_row.get("description", "")),
                "currency": currency,
                "label": label,
            })

    return pd.DataFrame(rows)


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")

    parser = argparse.ArgumentParser(description="Синтез обучающего датасета")
    parser.add_argument("--catalog", default="dataset/dataset_example.csv")
    parser.add_argument("--queries", type=int, default=1000)
    parser.add_argument("--output", default="dataset/training.parquet")
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    df = build_training_dataset(args.catalog, n_queries=args.queries, seed=args.seed)
    df.to_parquet(args.output, index=False)

    log.info("Датасет: %d строк, label dist: %s", len(df), df["label"].value_counts().to_dict())
    log.info("Сохранено → %s", args.output)


if __name__ == "__main__":
    main()
