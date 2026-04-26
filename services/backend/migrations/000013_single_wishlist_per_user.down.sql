ALTER TABLE wishlists
    DROP CONSTRAINT IF EXISTS uq_wishlists_user_id;

ALTER TABLE wishlists
    ADD CONSTRAINT uq_wishlists_user_id_name UNIQUE (user_id, name);
