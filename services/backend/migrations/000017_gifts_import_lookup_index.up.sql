-- Ускоряет GiftExists при импорте: LOWER(name) + LOWER(store_link)
CREATE INDEX IF NOT EXISTS idx_gifts_name_store_link_lower
    ON gifts (lower(name), lower(store_link));
