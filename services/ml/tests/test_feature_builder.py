"""Тесты feature_builder."""
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from ml_service.feature_builder import build_row, build_features, FEATURE_COLS
from ml_service.proto.ranking.v1 import ranking_pb2


def _query(budget: int = 5000, age: int | None = None, interests: list[str] | None = None) -> ranking_pb2.QueryContext:
    q = ranking_pb2.QueryContext(
        occasion="birthday",
        relationship="friend",
        budget_cents=budget,
        interests=interests or [],
    )
    if age is not None:
        q.recipient_age = age
    return q


def _candidate(cid: str = "g1", price: int = 1000, cat: str = "Книги", age_restriction: int | None = None) -> ranking_pb2.Candidate:
    c = ranking_pb2.Candidate(
        id=cid,
        price_cents=price,
        category_name=cat,
        title="Test Gift",
        description="Good gift",
    )
    if age_restriction is not None:
        c.age_restriction = age_restriction
    return c


def test_within_budget():
    row = build_row(_candidate(price=3000), _query(budget=5000))
    assert row["price_within_budget"] == 1
    assert row["budget_gap"] == 2000


def test_over_budget():
    row = build_row(_candidate(price=6000), _query(budget=5000))
    assert row["price_within_budget"] == 0
    assert row["budget_gap"] == 0


def test_interest_overlap():
    row = build_row(
        _candidate(cat="Книги"),
        _query(interests=["книги", "чтение"]),
    )
    assert row["interest_overlap"] >= 1


def test_all_feature_cols_present():
    row = build_row(_candidate(), _query())
    for col in FEATURE_COLS:
        assert col in row, f"Missing feature: {col}"


def test_build_features_multiple():
    request = ranking_pb2.RankRequest(
        selection_id="r1", user_id="u1", top_n=3,
        query=_query(),
        candidates=[_candidate("g1"), _candidate("g2"), _candidate("g3")],
    )
    features = build_features(request)
    assert len(features) == 3
    assert all("tfidf_score" in f for f in features)
