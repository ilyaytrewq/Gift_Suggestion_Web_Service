"""Тесты gRPC сервера — echo stub."""
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from ml_service.grpc_server import RankingServicer
from ml_service.ranker import Ranker
from ml_service.proto.ranking.v1 import ranking_pb2


def _make_candidate(cid: str, price: int, title: str = "Gift") -> ranking_pb2.Candidate:
    return ranking_pb2.Candidate(
        id=cid,
        price_cents=price,
        title=title,
        description="desc",
        category_name="Books",
    )


def _make_query(budget: int = 5000) -> ranking_pb2.QueryContext:
    return ranking_pb2.QueryContext(
        occasion="birthday",
        relationship="friend",
        interests=["books"],
        budget_cents=budget,
    )


def test_rank_returns_top_n():
    ranker = Ranker()
    servicer = RankingServicer(ranker)

    request = ranking_pb2.RankRequest(
        selection_id="req-1",
        user_id="user-1",
        top_n=2,
        query=_make_query(),
        candidates=[
            _make_candidate("g1", 1000),
            _make_candidate("g2", 2000),
            _make_candidate("g3", 3000),
        ],
    )

    response = servicer.Rank(request, None)
    assert len(response.items) == 2


def test_rank_empty_candidates():
    ranker = Ranker()
    servicer = RankingServicer(ranker)

    request = ranking_pb2.RankRequest(
        selection_id="req-2",
        user_id="user-2",
        top_n=5,
        query=_make_query(),
        candidates=[],
    )

    response = servicer.Rank(request, None)
    assert len(response.items) == 0


def test_rank_scores_are_ordered():
    ranker = Ranker()
    servicer = RankingServicer(ranker)

    request = ranking_pb2.RankRequest(
        selection_id="req-3",
        user_id="user-3",
        top_n=3,
        query=_make_query(budget=5000),
        candidates=[
            _make_candidate("g1", 4900),
            _make_candidate("g2", 100),
            _make_candidate("g3", 3000),
        ],
    )

    response = servicer.Rank(request, None)
    scores = [item.score for item in response.items]
    assert scores == sorted(scores, reverse=True)


def test_rank_has_explanations():
    ranker = Ranker()
    servicer = RankingServicer(ranker)

    request = ranking_pb2.RankRequest(
        selection_id="req-4",
        user_id="user-4",
        top_n=1,
        query=_make_query(),
        candidates=[_make_candidate("g1", 1000, title="Book about coding")],
    )

    response = servicer.Rank(request, None)
    assert len(response.items) == 1
    assert len(response.items[0].explanations) > 0
