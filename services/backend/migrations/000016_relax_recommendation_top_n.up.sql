ALTER TABLE recommendation_requests
    DROP CONSTRAINT IF EXISTS chk_recommendation_requests_top_n;

ALTER TABLE recommendation_requests
    ADD CONSTRAINT chk_recommendation_requests_top_n
    CHECK (requested_top_n > 0 AND requested_top_n <= 200);
