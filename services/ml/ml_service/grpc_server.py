"""gRPC сервер RankingService с feature_builder и реальным ранжировщиком."""
from __future__ import annotations

import logging
import math
import re
from concurrent import futures

import grpc
from grpc_health.v1 import health, health_pb2, health_pb2_grpc

from ml_service.explainer import explain
from ml_service.feature_builder import build_features, FEATURE_COLS
from ml_service.proto.ranking.v1 import ranking_pb2, ranking_pb2_grpc
from ml_service.ranker import Ranker

log = logging.getLogger(__name__)


def _tokenize(text: str) -> set[str]:
    return set(re.findall(r"\w+", text.lower()))


def _cosine(a: set[str], b: set[str]) -> float:
    if not a or not b:
        return 0.0
    return len(a & b) / math.sqrt(len(a) * len(b))


def _candidate_tokens(c: ranking_pb2.Candidate) -> set[str]:
    return _tokenize(c.title + " " + c.description + " " + c.category_name)


def _find_alternatives(
    ranked_top_ids: set[str],
    all_ranked: list[tuple[ranking_pb2.Candidate, float, dict]],
    candidate: ranking_pb2.Candidate,
    n: int = 2,
) -> list[str]:
    """Альтернативы: похожие на кандидата товары, не вошедшие в топ."""
    query_tokens = _candidate_tokens(candidate)
    scored = []
    for c, score, _ in all_ranked:
        if c.id == candidate.id or c.id in ranked_top_ids:
            continue
        sim = _cosine(query_tokens, _candidate_tokens(c))
        scored.append((c.id, sim))
    scored.sort(key=lambda x: x[1], reverse=True)
    return [cid for cid, _ in scored[:n]]


class RankingServicer(ranking_pb2_grpc.RankingServiceServicer):
    def __init__(self, ranker: Ranker) -> None:
        self._ranker = ranker

    def Rank(self, request: ranking_pb2.RankRequest, context: grpc.ServicerContext) -> ranking_pb2.RankResponse:
        if not request.candidates:
            return ranking_pb2.RankResponse()

        features_list = build_features(request)
        scores = self._ranker.predict(features_list)

        # zip candidates with their features and scores
        cand_data = list(zip(request.candidates, scores, features_list))
        cand_data.sort(key=lambda x: x[1], reverse=True)

        top_n = request.top_n if request.top_n > 0 else len(cand_data)
        top_slice = cand_data[:top_n]
        top_ids = {c.id for c, _, _ in top_slice}

        items = []
        for candidate, score, feats in top_slice:
            alts = _find_alternatives(top_ids, cand_data, candidate, n=2)
            explanations = explain(candidate, request.query, feats)
            items.append(
                ranking_pb2.RankedItem(
                    candidate_id=candidate.id,
                    score=score,
                    explanations=explanations,
                    alternative_candidate_ids=alts,
                )
            )

        return ranking_pb2.RankResponse(items=items, model_version=self._ranker.model_version)


def create_server(ranker: Ranker, addr: str) -> grpc.Server:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    ranking_pb2_grpc.add_RankingServiceServicer_to_server(RankingServicer(ranker), server)

    health_servicer = health.HealthServicer()
    health_pb2_grpc.add_HealthServicer_to_server(health_servicer, server)
    health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)
    health_servicer.set("gift.ranking.v1.RankingService", health_pb2.HealthCheckResponse.SERVING)

    server.add_insecure_port(addr)
    return server
