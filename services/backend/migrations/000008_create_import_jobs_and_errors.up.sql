ALTER TABLE gifts
    ADD COLUMN source_name TEXT;

CREATE TABLE import_jobs (
    id UUID PRIMARY KEY,
    requested_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    status TEXT NOT NULL,
    source_format TEXT NOT NULL,
    source_filename TEXT NOT NULL,
    source_label TEXT,
    source_size_bytes BIGINT NOT NULL,
    total_rows INTEGER NOT NULL DEFAULT 0,
    processed_rows INTEGER NOT NULL DEFAULT 0,
    imported_rows INTEGER NOT NULL DEFAULT 0,
    updated_rows INTEGER NOT NULL DEFAULT 0,
    skipped_rows INTEGER NOT NULL DEFAULT 0,
    duplicate_in_file_rows INTEGER NOT NULL DEFAULT 0,
    duplicate_in_catalog_rows INTEGER NOT NULL DEFAULT 0,
    error_rows INTEGER NOT NULL DEFAULT 0,
    failure_code TEXT,
    failure_message TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_import_jobs_status CHECK (
        status IN ('pending', 'running', 'completed', 'completed_with_errors', 'failed')
    ),
    CONSTRAINT chk_import_jobs_source_format CHECK (
        source_format IN ('csv', 'json', 'xlsx')
    ),
    CONSTRAINT chk_import_jobs_source_size_bytes CHECK (source_size_bytes >= 0),
    CONSTRAINT chk_import_jobs_total_rows CHECK (total_rows >= 0),
    CONSTRAINT chk_import_jobs_processed_rows CHECK (processed_rows >= 0),
    CONSTRAINT chk_import_jobs_imported_rows CHECK (imported_rows >= 0),
    CONSTRAINT chk_import_jobs_updated_rows CHECK (updated_rows >= 0),
    CONSTRAINT chk_import_jobs_skipped_rows CHECK (skipped_rows >= 0),
    CONSTRAINT chk_import_jobs_duplicate_in_file_rows CHECK (duplicate_in_file_rows >= 0),
    CONSTRAINT chk_import_jobs_duplicate_in_catalog_rows CHECK (duplicate_in_catalog_rows >= 0),
    CONSTRAINT chk_import_jobs_error_rows CHECK (error_rows >= 0)
);

CREATE TABLE import_errors (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES import_jobs (id) ON DELETE CASCADE,
    row_number INTEGER,
    record_key TEXT,
    field_name TEXT,
    error_code TEXT NOT NULL,
    message TEXT NOT NULL,
    raw_record JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_import_jobs_status_created_at ON import_jobs (status, created_at DESC);
CREATE INDEX idx_import_jobs_requested_by_created_at ON import_jobs (requested_by_user_id, created_at DESC);
CREATE INDEX idx_import_errors_job_row_number ON import_errors (job_id, row_number);
CREATE INDEX idx_import_errors_job_code ON import_errors (job_id, error_code);
