ALTER TABLE users
    ADD COLUMN display_name TEXT,
    ADD COLUMN last_login_at TIMESTAMPTZ,
    ADD COLUMN password_changed_at TIMESTAMPTZ;
