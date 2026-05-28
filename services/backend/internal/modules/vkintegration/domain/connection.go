package domain

import (
	"net/url"
	"slices"
	"strings"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

const (
	maxProviderUserIDLength  = 128
	maxConsentVersionLength  = 32
	maxScreenNameLength      = 64
	maxProfileURLLength      = 512
	maxScopeLength           = 64
	maxTokenCiphertextLength = 4096
	maxSyncErrorCodeLength   = 64
)

type ConnectionState string

const (
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateDisconnected ConnectionState = "disconnected"
)

type ConsentState string

const (
	ConsentStatePending ConsentState = "pending"
	ConsentStateGranted ConsentState = "granted"
	ConsentStateRevoked ConsentState = "revoked"
)

type SyncState string

const (
	SyncStateIdle      SyncState = "idle"
	SyncStateSucceeded SyncState = "succeeded"
	SyncStateFailed    SyncState = "failed"
)

type IntegrationMetadata struct {
	screenName *string
	profileURL *string
}

type Connection struct {
	id                ConnectionID
	userID            userdomain.UserID
	providerUserID    string
	connectionState   ConnectionState
	consentState      ConsentState
	consentVersion    string
	tokenCiphertext   *string
	tokenExpiresAt    *time.Time
	scopes            []string
	metadata          IntegrationMetadata
	lastSyncState     SyncState
	lastSyncErrorCode *string
	lastSyncedAt      *time.Time
	connectedAt       *time.Time
	disconnectedAt    *time.Time
	consentGrantedAt  *time.Time
	consentRevokedAt  *time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

func NewConnection(
	id ConnectionID,
	userID userdomain.UserID,
	providerUserID string,
	consentVersion string,
	tokenCiphertext *string,
	tokenExpiresAt *time.Time,
	scopes []string,
	metadata IntegrationMetadata,
	now time.Time,
	consentGrantedAt time.Time,
) (Connection, error) {
	normalizedProviderUserID, err := normalizeProviderUserID(providerUserID)
	if err != nil {
		return Connection{}, err
	}
	normalizedConsentVersion, err := normalizeConsentVersion(consentVersion)
	if err != nil {
		return Connection{}, err
	}
	normalizedTokenCiphertext, err := normalizeCiphertext(tokenCiphertext)
	if err != nil {
		return Connection{}, err
	}
	normalizedScopes, err := normalizeScopes(scopes)
	if err != nil {
		return Connection{}, err
	}
	if consentGrantedAt.IsZero() {
		return Connection{}, ErrConsentGrantedAtRequired
	}
	if now.IsZero() {
		now = consentGrantedAt
	}

	connectedAt := now.UTC()
	grantedAt := consentGrantedAt.UTC()

	return Connection{
		id:               id,
		userID:           userID,
		providerUserID:   normalizedProviderUserID,
		connectionState:  ConnectionStateConnected,
		consentState:     ConsentStateGranted,
		consentVersion:   normalizedConsentVersion,
		tokenCiphertext:  cloneStringPtr(normalizedTokenCiphertext),
		tokenExpiresAt:   cloneTimePtr(tokenExpiresAt),
		scopes:           append([]string(nil), normalizedScopes...),
		metadata:         metadata,
		lastSyncState:    SyncStateIdle,
		connectedAt:      &connectedAt,
		consentGrantedAt: &grantedAt,
		createdAt:        now.UTC(),
		updatedAt:        now.UTC(),
	}, nil
}

func RestoreConnection(
	id string,
	userID string,
	providerUserID string,
	connectionState string,
	consentState string,
	consentVersion string,
	tokenCiphertext *string,
	tokenExpiresAt *time.Time,
	scopes []string,
	metadata IntegrationMetadata,
	lastSyncState string,
	lastSyncErrorCode *string,
	lastSyncedAt *time.Time,
	connectedAt *time.Time,
	disconnectedAt *time.Time,
	consentGrantedAt *time.Time,
	consentRevokedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (Connection, error) {
	connectionID, err := NewConnectionID(id)
	if err != nil {
		return Connection{}, err
	}

	parsedUserID, err := userdomain.NewUserID(userID)
	if err != nil {
		return Connection{}, err
	}

	normalizedProviderUserID, err := normalizeProviderUserID(providerUserID)
	if err != nil {
		return Connection{}, err
	}
	normalizedConnectionState, err := NewConnectionState(connectionState)
	if err != nil {
		return Connection{}, err
	}
	normalizedConsentState, err := NewConsentState(consentState)
	if err != nil {
		return Connection{}, err
	}
	normalizedConsentVersion, err := normalizeConsentVersion(consentVersion)
	if err != nil {
		return Connection{}, err
	}
	normalizedTokenCiphertext, err := normalizeCiphertext(tokenCiphertext)
	if err != nil {
		return Connection{}, err
	}
	normalizedScopes, err := normalizeScopes(scopes)
	if err != nil {
		return Connection{}, err
	}
	normalizedLastSyncState, err := NewSyncState(lastSyncState)
	if err != nil {
		return Connection{}, err
	}
	normalizedLastSyncErrorCode, err := normalizeSyncErrorCode(lastSyncErrorCode)
	if err != nil {
		return Connection{}, err
	}
	if normalizedConnectionState == ConnectionStateConnected && (connectedAt == nil || connectedAt.IsZero()) {
		return Connection{}, ErrConnectedAtRequired
	}
	if normalizedConsentState == ConsentStateGranted && (consentGrantedAt == nil || consentGrantedAt.IsZero()) {
		return Connection{}, ErrConsentGrantedAtRequired
	}

	return Connection{
		id:                connectionID,
		userID:            parsedUserID,
		providerUserID:    normalizedProviderUserID,
		connectionState:   normalizedConnectionState,
		consentState:      normalizedConsentState,
		consentVersion:    normalizedConsentVersion,
		tokenCiphertext:   cloneStringPtr(normalizedTokenCiphertext),
		tokenExpiresAt:    cloneTimePtr(tokenExpiresAt),
		scopes:            append([]string(nil), normalizedScopes...),
		metadata:          metadata,
		lastSyncState:     normalizedLastSyncState,
		lastSyncErrorCode: cloneStringPtr(normalizedLastSyncErrorCode),
		lastSyncedAt:      cloneTimePtr(lastSyncedAt),
		connectedAt:       cloneTimePtr(connectedAt),
		disconnectedAt:    cloneTimePtr(disconnectedAt),
		consentGrantedAt:  cloneTimePtr(consentGrantedAt),
		consentRevokedAt:  cloneTimePtr(consentRevokedAt),
		createdAt:         createdAt.UTC(),
		updatedAt:         updatedAt.UTC(),
	}, nil
}

func NewConnectionState(raw string) (ConnectionState, error) {
	switch ConnectionState(strings.TrimSpace(raw)) {
	case ConnectionStateConnected:
		return ConnectionStateConnected, nil
	case ConnectionStateDisconnected:
		return ConnectionStateDisconnected, nil
	default:
		return "", ErrInvalidConnectionState
	}
}

func NewConsentState(raw string) (ConsentState, error) {
	switch ConsentState(strings.TrimSpace(raw)) {
	case ConsentStatePending:
		return ConsentStatePending, nil
	case ConsentStateGranted:
		return ConsentStateGranted, nil
	case ConsentStateRevoked:
		return ConsentStateRevoked, nil
	default:
		return "", ErrInvalidConsentState
	}
}

func NewSyncState(raw string) (SyncState, error) {
	switch SyncState(strings.TrimSpace(raw)) {
	case SyncStateIdle:
		return SyncStateIdle, nil
	case SyncStateSucceeded:
		return SyncStateSucceeded, nil
	case SyncStateFailed:
		return SyncStateFailed, nil
	default:
		return "", ErrInvalidSyncState
	}
}

func NewIntegrationMetadata(screenName, profileURL *string) (IntegrationMetadata, error) {
	normalizedScreenName, err := normalizeOptionalString(screenName, maxScreenNameLength, ErrScreenNameTooLong)
	if err != nil {
		return IntegrationMetadata{}, err
	}
	normalizedProfileURL, err := normalizeProfileURL(profileURL)
	if err != nil {
		return IntegrationMetadata{}, err
	}

	return IntegrationMetadata{
		screenName: cloneStringPtr(normalizedScreenName),
		profileURL: cloneStringPtr(normalizedProfileURL),
	}, nil
}

func (c *Connection) Reconnect(
	providerUserID string,
	consentVersion string,
	tokenCiphertext *string,
	tokenExpiresAt *time.Time,
	scopes []string,
	metadata IntegrationMetadata,
	now time.Time,
	consentGrantedAt time.Time,
) error {
	normalizedProviderUserID, err := normalizeProviderUserID(providerUserID)
	if err != nil {
		return err
	}
	normalizedConsentVersion, err := normalizeConsentVersion(consentVersion)
	if err != nil {
		return err
	}
	normalizedTokenCiphertext, err := normalizeCiphertext(tokenCiphertext)
	if err != nil {
		return err
	}
	normalizedScopes, err := normalizeScopes(scopes)
	if err != nil {
		return err
	}
	if consentGrantedAt.IsZero() {
		return ErrConsentGrantedAtRequired
	}
	if now.IsZero() {
		now = consentGrantedAt
	}

	connectedAt := now.UTC()
	grantedAt := consentGrantedAt.UTC()

	c.providerUserID = normalizedProviderUserID
	c.connectionState = ConnectionStateConnected
	c.consentState = ConsentStateGranted
	c.consentVersion = normalizedConsentVersion
	c.tokenCiphertext = cloneStringPtr(normalizedTokenCiphertext)
	c.tokenExpiresAt = cloneTimePtr(tokenExpiresAt)
	c.scopes = append([]string(nil), normalizedScopes...)
	c.metadata = metadata
	c.lastSyncState = SyncStateIdle
	c.lastSyncErrorCode = nil
	c.lastSyncedAt = nil
	c.connectedAt = &connectedAt
	c.disconnectedAt = nil
	c.consentGrantedAt = &grantedAt
	c.consentRevokedAt = nil
	c.updatedAt = now.UTC()

	return nil
}

func (c *Connection) Disconnect(now time.Time) {
	disconnectedAt := now.UTC()
	c.connectionState = ConnectionStateDisconnected
	c.consentState = ConsentStateRevoked
	c.tokenCiphertext = nil
	c.tokenExpiresAt = nil
	c.scopes = []string{}
	c.lastSyncState = SyncStateIdle
	c.lastSyncErrorCode = nil
	c.disconnectedAt = &disconnectedAt
	c.consentRevokedAt = &disconnectedAt
	c.updatedAt = disconnectedAt
}

func (c *Connection) MarkSyncSucceeded(now time.Time) {
	syncedAt := now.UTC()
	c.lastSyncState = SyncStateSucceeded
	c.lastSyncErrorCode = nil
	c.lastSyncedAt = &syncedAt
	c.updatedAt = syncedAt
}

func (c *Connection) MarkSyncFailed(code string, now time.Time) error {
	normalizedCode, err := normalizeSyncErrorCode(&code)
	if err != nil {
		return err
	}

	failedAt := now.UTC()
	c.lastSyncState = SyncStateFailed
	c.lastSyncErrorCode = cloneStringPtr(normalizedCode)
	c.updatedAt = failedAt

	return nil
}

func (c Connection) ID() ConnectionID {
	return c.id
}

func (c Connection) UserID() userdomain.UserID {
	return c.userID
}

func (c Connection) ProviderUserID() string {
	return c.providerUserID
}

func (c Connection) ConnectionState() ConnectionState {
	return c.connectionState
}

func (c Connection) ConsentState() ConsentState {
	return c.consentState
}

func (c Connection) ConsentVersion() string {
	return c.consentVersion
}

func (c Connection) TokenCiphertext() *string {
	return cloneStringPtr(c.tokenCiphertext)
}

func (c Connection) TokenExpiresAt() *time.Time {
	return cloneTimePtr(c.tokenExpiresAt)
}

func (c Connection) Scopes() []string {
	return append([]string(nil), c.scopes...)
}

func (c Connection) Metadata() IntegrationMetadata {
	return c.metadata
}

func (c Connection) LastSyncState() SyncState {
	return c.lastSyncState
}

func (c Connection) LastSyncErrorCode() *string {
	return cloneStringPtr(c.lastSyncErrorCode)
}

func (c Connection) LastSyncedAt() *time.Time {
	return cloneTimePtr(c.lastSyncedAt)
}

func (c Connection) ConnectedAt() *time.Time {
	return cloneTimePtr(c.connectedAt)
}

func (c Connection) DisconnectedAt() *time.Time {
	return cloneTimePtr(c.disconnectedAt)
}

func (c Connection) ConsentGrantedAt() *time.Time {
	return cloneTimePtr(c.consentGrantedAt)
}

func (c Connection) ConsentRevokedAt() *time.Time {
	return cloneTimePtr(c.consentRevokedAt)
}

func (c Connection) CreatedAt() time.Time {
	return c.createdAt
}

func (c Connection) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c Connection) HasGrantedConsent() bool {
	return c.consentState == ConsentStateGranted
}

func (c Connection) IsConnected() bool {
	return c.connectionState == ConnectionStateConnected
}

func (c Connection) HasTokenConfigured() bool {
	return c.tokenCiphertext != nil && strings.TrimSpace(*c.tokenCiphertext) != ""
}

func (c Connection) TokenExpired(now time.Time) bool {
	if c.tokenExpiresAt == nil {
		return false
	}

	return !c.tokenExpiresAt.After(now.UTC())
}

func (m IntegrationMetadata) ScreenName() *string {
	return cloneStringPtr(m.screenName)
}

func (m IntegrationMetadata) ProfileURL() *string {
	return cloneStringPtr(m.profileURL)
}

func normalizeProviderUserID(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrProviderUserIDEmpty
	}
	if len(value) > maxProviderUserIDLength {
		return "", ErrProviderUserIDTooLong
	}

	return value, nil
}

func normalizeConsentVersion(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrConsentVersionEmpty
	}
	if len(value) > maxConsentVersionLength {
		return "", ErrConsentVersionTooLong
	}

	return value, nil
}

func normalizeOptionalString(value *string, maxLen int, tooLongErr error) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, nil
	}
	if len(normalized) > maxLen {
		return nil, tooLongErr
	}

	return &normalized, nil
}

func normalizeProfileURL(value *string) (*string, error) {
	normalized, err := normalizeOptionalString(value, maxProfileURLLength, ErrProfileURLTooLong)
	if err != nil || normalized == nil {
		return normalized, err
	}

	parsed, err := url.ParseRequestURI(*normalized)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, ErrInvalidProfileURL
	}

	return normalized, nil
}

func normalizeScopes(scopes []string) ([]string, error) {
	if len(scopes) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(scopes))
	normalized := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		value := strings.ToLower(strings.TrimSpace(scope))
		if value == "" || len(value) > maxScopeLength {
			return nil, ErrInvalidScope
		}
		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	slices.Sort(normalized)
	return normalized, nil
}

func normalizeCiphertext(value *string) (*string, error) {
	normalized, err := normalizeOptionalString(value, maxTokenCiphertextLength, ErrTokenCiphertextTooLong)
	if err != nil {
		return nil, err
	}

	return normalized, nil
}

func normalizeSyncErrorCode(value *string) (*string, error) {
	normalized, err := normalizeOptionalString(value, maxSyncErrorCodeLength, ErrSyncErrorCodeTooLong)
	if err != nil {
		return nil, err
	}

	return normalized, nil
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := value.UTC()
	return &cloned
}
