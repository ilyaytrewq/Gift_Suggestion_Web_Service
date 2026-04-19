CREATE TABLE recommendation_requests (
    id UUID PRIMARY KEY,
    requested_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    status TEXT NOT NULL,
    ranking_source TEXT NOT NULL DEFAULT 'none',
    criteria_version TEXT NOT NULL DEFAULT 'v1',
    questionnaire JSONB NOT NULL,
    hard_filters JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_top_n SMALLINT NOT NULL,
    candidate_count_before_filters INTEGER NOT NULL DEFAULT 0,
    candidate_count_after_filters INTEGER NOT NULL DEFAULT 0,
    returned_primary_count INTEGER NOT NULL DEFAULT 0,
    returned_alternative_count INTEGER NOT NULL DEFAULT 0,
    fallback_reason_code TEXT,
    failure_code TEXT,
    failure_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_recommendation_requests_status CHECK (
        status IN ('pending', 'running', 'completed', 'completed_with_fallback', 'completed_empty', 'failed')
    ),
    CONSTRAINT chk_recommendation_requests_ranking_source CHECK (
        ranking_source IN ('none', 'ml', 'fallback')
    ),
    CONSTRAINT chk_recommendation_requests_top_n CHECK (
        requested_top_n > 0 AND requested_top_n <= 10
    ),
    CONSTRAINT chk_recommendation_requests_candidate_count_before_filters CHECK (
        candidate_count_before_filters >= 0
    ),
    CONSTRAINT chk_recommendation_requests_candidate_count_after_filters CHECK (
        candidate_count_after_filters >= 0
    ),
    CONSTRAINT chk_recommendation_requests_returned_primary_count CHECK (
        returned_primary_count >= 0
    ),
    CONSTRAINT chk_recommendation_requests_returned_alternative_count CHECK (
        returned_alternative_count >= 0
    )
);

CREATE TABLE recommendation_results (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL REFERENCES recommendation_requests (id) ON DELETE CASCADE,
    gift_id UUID NOT NULL REFERENCES gifts (id),
    slot_position SMALLINT NOT NULL,
    result_kind TEXT NOT NULL,
    alternative_rank SMALLINT,
    ranking_source TEXT NOT NULL,
    score DOUBLE PRECISION,
    explanations JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_recommendation_results_slot_position CHECK (slot_position > 0),
    CONSTRAINT chk_recommendation_results_result_kind CHECK (
        result_kind IN ('primary', 'alternative')
    ),
    CONSTRAINT chk_recommendation_results_ranking_source CHECK (
        ranking_source IN ('ml', 'fallback')
    ),
    CONSTRAINT chk_recommendation_results_alternative_rank CHECK (
        (result_kind = 'primary' AND alternative_rank IS NULL)
        OR
        (result_kind = 'alternative' AND alternative_rank IS NOT NULL AND alternative_rank > 0)
    )
);

CREATE INDEX idx_recommendation_requests_user_created_at
    ON recommendation_requests (requested_by_user_id, created_at DESC);

CREATE INDEX idx_recommendation_requests_status_created_at
    ON recommendation_requests (status, created_at DESC);

CREATE INDEX idx_recommendation_results_request_gift
    ON recommendation_results (request_id, gift_id);

CREATE UNIQUE INDEX uq_recommendation_results_primary_slot
    ON recommendation_results (request_id, slot_position)
    WHERE result_kind = 'primary';

CREATE UNIQUE INDEX uq_recommendation_results_alternative_slot
    ON recommendation_results (request_id, slot_position, alternative_rank)
    WHERE result_kind = 'alternative';
