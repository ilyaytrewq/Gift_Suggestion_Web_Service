DO $$
DECLARE
    canonical RECORD;
BEGIN
    FOR canonical IN
        SELECT DISTINCT ON (user_id)
            id,
            user_id
        FROM wishlists
        ORDER BY user_id, created_at ASC, id ASC
    LOOP
        INSERT INTO wishlist_items (id, wishlist_id, gift_id, created_at)
        SELECT
            gen_random_uuid(),
            canonical.id,
            wi.gift_id,
            wi.created_at
        FROM wishlist_items wi
        INNER JOIN wishlists w ON w.id = wi.wishlist_id
        WHERE w.user_id = canonical.user_id
          AND wi.wishlist_id <> canonical.id
        ON CONFLICT (wishlist_id, gift_id) DO NOTHING;
    END LOOP;
END $$;

DELETE FROM wishlists w
WHERE EXISTS (
    SELECT 1
    FROM wishlists keep
    WHERE keep.user_id = w.user_id
      AND (
          keep.created_at < w.created_at
          OR (keep.created_at = w.created_at AND keep.id < w.id)
      )
);

UPDATE wishlists
SET
    name = 'Список желаний',
    updated_at = GREATEST(updated_at, NOW());

ALTER TABLE wishlists
    DROP CONSTRAINT IF EXISTS uq_wishlists_user_id_name;

ALTER TABLE wishlists
    ADD CONSTRAINT uq_wishlists_user_id UNIQUE (user_id);
