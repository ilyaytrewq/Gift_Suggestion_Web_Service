"""Генерация объяснений рекомендаций на основе признаков."""
from __future__ import annotations

from ml_service.proto.ranking.v1 import ranking_pb2


def explain(
    candidate: ranking_pb2.Candidate,
    query: ranking_pb2.QueryContext,
    features: dict,
) -> list[ranking_pb2.Explanation]:
    result = []

    if features.get("price_within_budget"):
        result.append(ranking_pb2.Explanation(code="fits_budget", text="Подходит по бюджету."))

    if features.get("interest_overlap", 0) > 0:
        result.append(ranking_pb2.Explanation(code="matches_interests", text="Соотносится с интересами."))

    if features.get("tfidf_score", 0.0) > 0.1:
        result.append(ranking_pb2.Explanation(code="text_relevance", text="Релевантен запросу."))

    if features.get("age_fits", 1):
        if query.HasField("recipient_age"):
            result.append(ranking_pb2.Explanation(code="age_appropriate", text="Учитывает возраст получателя."))

    if features.get("category_match", 0) or features.get("category_index", -1) >= 0:
        result.append(ranking_pb2.Explanation(code="category_match", text="Совпадает с категорией."))

    if not result:
        result.append(ranking_pb2.Explanation(code="matches_request", text="Соответствует параметрам."))

    return result
