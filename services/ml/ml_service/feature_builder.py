"""Построение признаков из запроса ранжирования."""
from __future__ import annotations

import math
import re
from typing import Any

from ml_service.proto.ranking.v1 import ranking_pb2


BUDGET_BUCKETS = [500, 1000, 2000, 5000, 10000, 20000, 50000]
PRICE_BUCKETS = [500, 1000, 2000, 5000, 10000, 20000, 50000]

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


def _bucket(value: int, buckets: list[int]) -> int:
    for i, b in enumerate(buckets):
        if value <= b:
            return i
    return len(buckets)


def _tokenize(text: str) -> set[str]:
    return set(re.findall(r"\w+", text.lower()))


def _tfidf_overlap(tokens_a: set[str], tokens_b: set[str]) -> float:
    if not tokens_a or not tokens_b:
        return 0.0
    intersection = tokens_a & tokens_b
    return len(intersection) / math.sqrt(len(tokens_a) * len(tokens_b))


def build_row(
    candidate: ranking_pb2.Candidate,
    query: ranking_pb2.QueryContext,
) -> dict[str, Any]:
    budget = max(query.budget_cents, 1)
    price = candidate.price_cents

    within_budget = price <= budget
    budget_gap = max(0, budget - price)
    budget_ratio = price / budget
    price_bucket = _bucket(price, PRICE_BUCKETS)
    budget_bucket = _bucket(budget, BUDGET_BUCKETS)

    age_fits = True
    if query.HasField("recipient_age") and candidate.HasField("age_restriction"):
        age_fits = candidate.age_restriction <= query.recipient_age

    interest_tokens = set()
    for interest in query.interests:
        interest_tokens.update(_tokenize(interest))
    if query.occasion:
        interest_tokens.update(_tokenize(query.occasion))

    corpus_tokens = _tokenize(
        candidate.title + " " + candidate.description + " " + candidate.category_name
    )

    interest_overlap = sum(
        1 for t in _tokenize(" ".join(query.interests)) if t in corpus_tokens
    ) if query.interests else 0
    tfidf_score = _tfidf_overlap(interest_tokens, corpus_tokens)

    cat_lower = candidate.category_name.lower()
    category_index = next(
        (i for i, c in enumerate(INTERNAL_CATEGORIES) if c in cat_lower),
        -1,
    )

    gender_map = {"male": 1, "female": 2, "other": 0}
    gender_encoded = gender_map.get(query.recipient_gender, -1)

    return {
        "candidate_id": candidate.id,
        "price_cents": price,
        "price_bucket": price_bucket,
        "budget_cents": budget,
        "budget_bucket": budget_bucket,
        "price_within_budget": int(within_budget),
        "budget_gap": budget_gap,
        "budget_ratio": min(budget_ratio, 5.0),
        "age_fits": int(age_fits),
        "interest_overlap": interest_overlap,
        "tfidf_score": tfidf_score,
        "category_index": category_index,
        "already_in_wishlist": int(candidate.already_in_wishlist),
        "gender_encoded": gender_encoded,
    }


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


def build_features(request: ranking_pb2.RankRequest) -> "list[dict]":
    return [build_row(c, request.query) for c in request.candidates]
