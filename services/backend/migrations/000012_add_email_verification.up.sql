ALTER TABLE users
    ADD COLUMN email_verified_at TIMESTAMPTZ;

CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    CONSTRAINT uq_email_verification_tokens_token_hash UNIQUE (token_hash)
);

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);
CREATE INDEX idx_email_verification_tokens_expires_at ON email_verification_tokens (expires_at);
