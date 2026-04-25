CREATE TABLE vk_connections (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    vk_user_id TEXT NOT NULL,
    connection_state TEXT NOT NULL,
    consent_state TEXT NOT NULL,
    consent_version TEXT NOT NULL DEFAULT 'v1',
    token_ciphertext TEXT NULL,
    token_expires_at TIMESTAMPTZ NULL,
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    integration_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_sync_state TEXT NOT NULL DEFAULT 'idle',
    last_sync_error_code TEXT NULL,
    last_synced_at TIMESTAMPTZ NULL,
    connected_at TIMESTAMPTZ NULL,
    disconnected_at TIMESTAMPTZ NULL,
    consent_granted_at TIMESTAMPTZ NULL,
    consent_revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT vk_connections_status_check CHECK (
        connection_state IN ('connected', 'disconnected')
    ),
    CONSTRAINT vk_connections_consent_check CHECK (
        consent_state IN ('pending', 'granted', 'revoked')
    ),
    CONSTRAINT vk_connections_sync_status_check CHECK (
        last_sync_state IN ('idle', 'succeeded', 'failed')
    ),
    CONSTRAINT vk_connections_scopes_array_check CHECK (
        jsonb_typeof(scopes) = 'array'
    ),
    CONSTRAINT vk_connections_integration_metadata_object_check CHECK (
        jsonb_typeof(integration_metadata) = 'object'
    ),
    CONSTRAINT vk_connections_connected_at_check CHECK (
        connection_state <> 'connected' OR connected_at IS NOT NULL
    ),
    CONSTRAINT vk_connections_consent_granted_at_check CHECK (
        consent_state <> 'granted' OR consent_granted_at IS NOT NULL
    )
);

CREATE UNIQUE INDEX vk_connections_active_vk_user_uidx
    ON vk_connections (vk_user_id)
    WHERE connection_state = 'connected';

CREATE INDEX vk_connections_user_status_idx
    ON vk_connections (user_id, connection_state, updated_at DESC);

CREATE INDEX vk_connections_sync_state_idx
    ON vk_connections (last_sync_state, last_synced_at DESC);

CREATE TABLE vk_imported_interests (
    connection_id UUID NOT NULL REFERENCES vk_connections(id) ON DELETE CASCADE,
    raw_value TEXT NOT NULL,
    normalized_value TEXT NOT NULL,
    source_label TEXT NOT NULL,
    position INTEGER NOT NULL,
    imported_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (connection_id, normalized_value),
    CONSTRAINT vk_imported_interests_normalized_check CHECK (
        length(normalized_value) > 0
    ),
    CONSTRAINT vk_imported_interests_position_check CHECK (
        position > 0
    )
);

CREATE INDEX vk_imported_interests_connection_position_idx
    ON vk_imported_interests (connection_id, position ASC);
