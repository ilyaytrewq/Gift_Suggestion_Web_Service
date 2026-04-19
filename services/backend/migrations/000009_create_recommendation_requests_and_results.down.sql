DROP INDEX IF EXISTS uq_recommendation_results_alternative_slot;
DROP INDEX IF EXISTS uq_recommendation_results_primary_slot;
DROP INDEX IF EXISTS idx_recommendation_results_request_gift;
DROP INDEX IF EXISTS idx_recommendation_requests_status_created_at;
DROP INDEX IF EXISTS idx_recommendation_requests_user_created_at;

DROP TABLE IF EXISTS recommendation_results;
DROP TABLE IF EXISTS recommendation_requests;
