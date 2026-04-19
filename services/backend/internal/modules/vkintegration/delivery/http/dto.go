package http

import "time"

type connectRequest struct {
	Consent        consentRequest     `json:"consent"`
	ProviderUserID string             `json:"provider_user_id"`
	Credential     *credentialRequest `json:"credential,omitempty"`
	Profile        *profileRequest    `json:"profile,omitempty"`
}

type consentRequest struct {
	Granted    bool      `json:"granted"`
	Version    string    `json:"version"`
	ObtainedAt time.Time `json:"obtained_at"`
}

type credentialRequest struct {
	AccessToken *string    `json:"access_token,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Scopes      []string   `json:"scopes,omitempty"`
}

type profileRequest struct {
	ScreenName *string `json:"screen_name,omitempty"`
	ProfileURL *string `json:"profile_url,omitempty"`
}
