package usecase

import (
	"time"

	vkintegrationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/domain"
)

type ConnectInput struct {
	UserID         string
	ProviderUserID string
	Consent        ConsentInput
	Credential     CredentialInput
	Profile        ProfileInput
}

type GetCurrentConnectionInput struct {
	UserID string
}

type DisconnectInput struct {
	UserID string
}

type SyncInterestsInput struct {
	UserID string
}

type ConsentInput struct {
	Granted    bool
	Version    string
	ObtainedAt time.Time
}

type CredentialInput struct {
	AccessToken *string
	ExpiresAt   *time.Time
	Scopes      []string
}

type ProfileInput struct {
	ScreenName *string
	ProfileURL *string
}

type ConnectOutput struct {
	Connection Connection `json:"connection"`
}

type GetCurrentConnectionOutput struct {
	Connection Connection `json:"connection"`
}

type DisconnectOutput struct {
	Connection Connection `json:"connection"`
}

type SyncInterestsOutput struct {
	Connection Connection `json:"connection"`
}

type Connection struct {
	Provider          string             `json:"provider"`
	State             string             `json:"state"`
	FeatureEnabled    bool               `json:"feature_enabled"`
	Consent           Consent            `json:"consent"`
	Profile           Profile            `json:"profile"`
	Credential        Credential         `json:"credential"`
	Sync              SyncStatus         `json:"sync"`
	ImportedInterests []ImportedInterest `json:"imported_interests"`
}

type Consent struct {
	State      string     `json:"state"`
	Granted    bool       `json:"granted"`
	Version    *string    `json:"version,omitempty"`
	ObtainedAt *time.Time `json:"obtained_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

type Profile struct {
	ProviderUserID *string `json:"provider_user_id,omitempty"`
	ScreenName     *string `json:"screen_name,omitempty"`
	ProfileURL     *string `json:"profile_url,omitempty"`
}

type Credential struct {
	Configured bool       `json:"configured"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Scopes     []string   `json:"scopes"`
}

type SyncStatus struct {
	LastSyncedAt           *time.Time `json:"last_synced_at,omitempty"`
	LastStatus             string     `json:"last_status"`
	LastErrorCode          *string    `json:"last_error_code,omitempty"`
	ImportedInterestsCount int        `json:"imported_interests_count"`
}

type ImportedInterest struct {
	Name           string    `json:"name"`
	NormalizedName string    `json:"normalized_name"`
	SourceLabel    string    `json:"source_label"`
	Position       int       `json:"position"`
	ImportedAt     time.Time `json:"imported_at"`
}

func newConnectionOutput(featureEnabled bool, connection *vkintegrationdomain.Connection, interests []vkintegrationdomain.ImportedInterest) Connection {
	if connection == nil {
		return Connection{
			Provider:       "vk",
			State:          "disconnected",
			FeatureEnabled: featureEnabled,
			Consent: Consent{
				State:   string(vkintegrationdomain.ConsentStatePending),
				Granted: false,
			},
			Credential: Credential{
				Configured: false,
				Scopes:     []string{},
			},
			Sync: SyncStatus{
				LastStatus:             string(vkintegrationdomain.SyncStateIdle),
				ImportedInterestsCount: 0,
			},
			ImportedInterests: []ImportedInterest{},
		}
	}

	view := Connection{
		Provider:       "vk",
		State:          resolveConnectionState(*connection),
		FeatureEnabled: featureEnabled,
		Consent: Consent{
			State:      string(connection.ConsentState()),
			Granted:    connection.HasGrantedConsent(),
			Version:    stringPtr(connection.ConsentVersion()),
			ObtainedAt: connection.ConsentGrantedAt(),
			RevokedAt:  connection.ConsentRevokedAt(),
		},
		Profile: Profile{
			ProviderUserID: stringPtr(connection.ProviderUserID()),
			ScreenName:     connection.Metadata().ScreenName(),
			ProfileURL:     connection.Metadata().ProfileURL(),
		},
		Credential: Credential{
			Configured: connection.HasTokenConfigured(),
			ExpiresAt:  connection.TokenExpiresAt(),
			Scopes:     connection.Scopes(),
		},
		Sync: SyncStatus{
			LastSyncedAt:           connection.LastSyncedAt(),
			LastStatus:             string(connection.LastSyncState()),
			LastErrorCode:          connection.LastSyncErrorCode(),
			ImportedInterestsCount: len(interests),
		},
		ImportedInterests: make([]ImportedInterest, 0, len(interests)),
	}

	for _, interest := range interests {
		view.ImportedInterests = append(view.ImportedInterests, ImportedInterest{
			Name:           interest.RawValue(),
			NormalizedName: interest.NormalizedValue(),
			SourceLabel:    interest.SourceLabel(),
			Position:       interest.Position(),
			ImportedAt:     interest.ImportedAt(),
		})
	}

	return view
}

func resolveConnectionState(connection vkintegrationdomain.Connection) string {
	if !connection.IsConnected() {
		return "disconnected"
	}
	if connection.LastSyncState() == vkintegrationdomain.SyncStateFailed {
		return "error"
	}
	if connection.LastSyncedAt() == nil {
		return "sync_required"
	}

	return "connected"
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	current := value
	return &current
}
