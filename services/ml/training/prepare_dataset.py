"""Подготовка обучающего датасета с признаками из feature_builder."""
from __future__ import annotations

import argparse
import logging
import math
import re
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pandas as pd

from training.synth_labels import build_training_dataset

log = logging.getLogger(__name__)

# Canonical feature list — must stay in sync with ml_service/feature_builder.py
FEATURE_COLS = [
    "price_cents",
    "price_bucket",
    "budget_cents",
    "budget_bucket",
    "price_within_budget",
    "budget_gap",
    "budget_ratio",
    "age_fits",
    "interest_overlap",
    "tfidf_score",
    "category_index",
    "already_in_wishlist",
    "gender_encoded",
]

PRICE_BUCKETS = [500, 1000, 2000, 5000, 10000, 20000, 50000]
BUDGET_BUCKETS = [500, 1000, 2000, 5000, 10000, 20000, 50000]

INTERNAL_CATEGORIES = [
    "электроника",
    "книги",
    "настольные игры",
    "косметика",
    "аксессуары",
    "одежда",
    "хобби",
    "украшения для дома",
    "еда и напитки",
    "детские товары",
]

GENDER_MAP = {"male": 1, "female": 2, "other": 0}


def _bucket(value: float, buckets: list[int]) -> int:
    for i, b in enumerate(buckets):
        if value <= b:
            return i
    return len(buckets)


def _tokenize(text: str) -> set[str]:
    return set(re.findall(r"\w+", text.lower()))


def _tfidf(tokens_a: set[str], tokens_b: set[str]) -> float:
    if not tokens_a or not tokens_b:
        return 0.0
    return len(tokens_a & tokens_b) / math.sqrt(len(tokens_a) * len(tokens_b))


def _category_index(category: str) -> int:
    cat_lower = category.lower()
    for i, c in enumerate(INTERNAL_CATEGORIES):
        if c in cat_lower:
            return i
    return -1


def enrich_features(df: pd.DataFrame) -> pd.DataFrame:
    budget = df["budget"].clip(lower=1)

    # Price / budget in cents (convert RUB float → integer cents)
    df["price_cents"] = (df["price"] * 100).round().astype(int)
    df["budget_cents"] = (df["budget"] * 100).round().astype(int)

    df["price_within_budget"] = (df["price"] <= df["budget"]).astype(int)
    df["budget_gap"] = (df["budget"] - df["price"]).clip(lower=0)
    df["budget_ratio"] = (df["price"] / budget).clip(upper=5.0)
    df["age_fits"] = (
        (df["age_min"] <= df["recipient_age"]) & (df["recipient_age"] <= df["age_max"])
    ).astype(int)

    df["price_bucket"] = df["price"].apply(lambda p: _bucket(p, PRICE_BUCKETS))
    df["budget_bucket"] = df["budget"].apply(lambda b: _bucket(b, BUDGET_BUCKETS))

    def interest_overlap(row: pd.Series) -> int:
        interests = str(row.get("interests", "")).split(",")
        corpus = _tokenize(
            str(row.get("title", ""))
            + " " + str(row.get("description", ""))
            + " " + str(row.get("category", ""))
        )
        return sum(1 for i in interests if i.strip().lower() in corpus)

    def tfidf_score(row: pd.Series) -> float:
        interests = str(row.get("interests", "")).split(",")
        occasion_tokens = _tokenize(str(row.get("occasion", "")))
        interest_tokens: set[str] = set()
        for i in interests:
            interest_tokens.update(_tokenize(i))
        interest_tokens.update(occasion_tokens)
        corpus = _tokenize(str(row.get("title", "")) + " " + str(row.get("category", "")))
        return _tfidf(interest_tokens, corpus)

    df["interest_overlap"] = df.apply(interest_overlap, axis=1)
    df["tfidf_score"] = df.apply(tfidf_score, axis=1)

    df["category_index"] = df["category"].apply(_category_index)

    # Synthetic training data has no wishlist context — always 0
    df["already_in_wishlist"] = 0

    # gender_encoded: use recipient_gender column if present, else -1 (unspecified)
    if "recipient_gender" in df.columns:
        df["gender_encoded"] = df["recipient_gender"].map(GENDER_MAP).fillna(-1).astype(int)
    else:
        df["gender_encoded"] = -1

    return df


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")

    parser = argparse.ArgumentParser(description="Подготовка обучающего датасета")
    parser.add_argument("--catalog", default="dataset/dataset_example.csv")
    parser.add_argument("--queries", type=int, default=1000)
    parser.add_argument("--output", default="dataset/training.parquet")
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    log.info("Генерируем синтетические запросы...")
    df = build_training_dataset(args.catalog, n_queries=args.queries, seed=args.seed)
    df = enrich_features(df)

    df.to_parquet(args.output, index=False)
    log.info("Готово: %d строк → %s", len(df), args.output)
    log.info("Label distribution: %s", df["label"].value_counts().to_dict())
    log.info("Feature columns present: %s", [c for c in FEATURE_COLS if c in df.columns])


if __name__ == "__main__":
    main()
