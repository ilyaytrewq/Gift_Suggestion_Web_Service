DROP INDEX IF EXISTS idx_import_errors_job_code;
DROP INDEX IF EXISTS idx_import_errors_job_row_number;
DROP INDEX IF EXISTS idx_import_jobs_requested_by_created_at;
DROP INDEX IF EXISTS idx_import_jobs_status_created_at;

DROP TABLE IF EXISTS import_errors;
DROP TABLE IF EXISTS import_jobs;

ALTER TABLE gifts
    DROP COLUMN IF EXISTS source_name;
