CREATE TABLE tracking_events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recommendation_request_id UUID NULL,
    wishlist_id UUID NULL,
    gift_id UUID NULL,
    client_event_id TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tracking_events_event_type_check CHECK (
        event_type IN ('recommendation_request', 'card_view', 'wishlist_add', 'outbound_click')
    ),
    CONSTRAINT tracking_events_client_event_id_length_check CHECK (
        client_event_id IS NULL OR length(client_event_id) <= 64
    ),
    CONSTRAINT tracking_events_metadata_object_check CHECK (
        jsonb_typeof(metadata) = 'object'
    ),
    CONSTRAINT tracking_events_required_refs_check CHECK (
        (event_type = 'recommendation_request' AND recommendation_request_id IS NOT NULL AND gift_id IS NULL AND wishlist_id IS NULL)
        OR
        (event_type = 'card_view' AND gift_id IS NOT NULL AND wishlist_id IS NULL)
        OR
        (event_type = 'wishlist_add' AND gift_id IS NOT NULL AND wishlist_id IS NOT NULL)
        OR
        (event_type = 'outbound_click' AND gift_id IS NOT NULL AND wishlist_id IS NULL)
    )
);

CREATE INDEX tracking_events_user_occurred_at_idx
    ON tracking_events (user_id, occurred_at DESC);

CREATE INDEX tracking_events_event_type_occurred_at_idx
    ON tracking_events (event_type, occurred_at DESC);

CREATE INDEX tracking_events_recommendation_request_occurred_at_idx
    ON tracking_events (recommendation_request_id, occurred_at DESC)
    WHERE recommendation_request_id IS NOT NULL;

CREATE INDEX tracking_events_gift_occurred_at_idx
    ON tracking_events (gift_id, occurred_at DESC)
    WHERE gift_id IS NOT NULL;

CREATE INDEX tracking_events_wishlist_occurred_at_idx
    ON tracking_events (wishlist_id, occurred_at DESC)
    WHERE wishlist_id IS NOT NULL;

CREATE UNIQUE INDEX tracking_events_user_client_event_id_uidx
    ON tracking_events (user_id, client_event_id)
    WHERE client_event_id IS NOT NULL;
