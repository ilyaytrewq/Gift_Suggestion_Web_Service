ALTER TABLE users
    DROP COLUMN IF EXISTS password_changed_at,
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS display_name;
