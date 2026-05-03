CREATE TABLE gift_offers (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    gift_id     UUID        NOT NULL REFERENCES gifts(id) ON DELETE CASCADE,
    store_name  TEXT        NOT NULL,
    store_url   TEXT        NOT NULL,
    price_cents BIGINT      NOT NULL,
    currency    TEXT        NOT NULL DEFAULT 'RUB',
    available   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_gift_offers_gift_store UNIQUE (gift_id, store_url),
    CONSTRAINT chk_gift_offers_price_non_negative CHECK (price_cents >= 0),
    CONSTRAINT chk_gift_offers_currency CHECK (currency <> '')
);

CREATE INDEX idx_gift_offers_gift_id ON gift_offers (gift_id);
