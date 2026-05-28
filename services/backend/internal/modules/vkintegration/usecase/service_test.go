package usecase

import (
	"context"
	"testing"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	vkintegrationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testVKUserID       = "550e8400-e29b-41d4-a716-446655448000"
	testVKConnectionID = "550e8400-e29b-41d4-a716-446655448001"
)

func TestServiceConnectSuccess(t *testing.T) {
	t.Parallel()

	service := mustVKService(t, vkServiceDeps{
		repo:       &fakeVKConnectionRepository{},
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, sealed: "sealed-token"},
	})

	output, err := service.Connect(context.Background(), ConnectInput{
		UserID:         testVKUserID,
		ProviderUserID: "vk_123",
		Consent: ConsentInput{
			Granted:    true,
			Version:    "v1",
			ObtainedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
		},
		Credential: CredentialInput{
			AccessToken: stringValuePtr("raw-token"),
			Scopes:      []string{"friends", "friends"},
		},
		Profile: ProfileInput{
			ScreenName: stringValuePtr("ilya"),
		},
	})
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if output.Connection.State != "sync_required" {
		t.Fatalf("Connect() state = %q, want %q", output.Connection.State, "sync_required")
	}
	if !output.Connection.Credential.Configured {
		t.Fatal("Connect() credential.configured = false, want true")
	}
	if len(output.Connection.Credential.Scopes) != 1 || output.Connection.Credential.Scopes[0] != "friends" {
		t.Fatalf("Connect() scopes = %v, want [friends]", output.Connection.Credential.Scopes)
	}
}

func TestServiceConnectRejectsDifferentActiveProviderUser(t *testing.T) {
	t.Parallel()

	connection := mustVKConnection(t, "vk_old", nil, time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC))
	service := mustVKService(t, vkServiceDeps{
		repo:       &fakeVKConnectionRepository{connection: &connection},
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, sealed: "sealed-token"},
	})

	_, err := service.Connect(context.Background(), ConnectInput{
		UserID:         testVKUserID,
		ProviderUserID: "vk_new",
		Consent: ConsentInput{
			Granted:    true,
			Version:    "v1",
			ObtainedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
		},
	})
	if err == nil {
		t.Fatal("Connect() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "vk_connection_already_exists" {
		t.Fatalf("Connect() code = %q, want %q", appErr.Code(), "vk_connection_already_exists")
	}
}

func TestServiceConnectRejectsTokenWhenStorageNotConfigured(t *testing.T) {
	t.Parallel()

	service := mustVKService(t, vkServiceDeps{
		repo:       &fakeVKConnectionRepository{},
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{},
	})

	_, err := service.Connect(context.Background(), ConnectInput{
		UserID:         testVKUserID,
		ProviderUserID: "vk_123",
		Consent: ConsentInput{
			Granted:    true,
			Version:    "v1",
			ObtainedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
		},
		Credential: CredentialInput{
			AccessToken: stringValuePtr("raw-token"),
		},
	})
	if err == nil {
		t.Fatal("Connect() expected unavailable error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "vk_token_storage_not_configured" {
		t.Fatalf("Connect() code = %q, want %q", appErr.Code(), "vk_token_storage_not_configured")
	}
}

func TestServiceDisconnectWithoutActiveConnection(t *testing.T) {
	t.Parallel()

	service := mustVKService(t, vkServiceDeps{
		repo:       &fakeVKConnectionRepository{},
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{},
	})

	output, err := service.Disconnect(context.Background(), DisconnectInput{
		UserID: testVKUserID,
	})
	if err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}

	if output.Connection.State != "disconnected" {
		t.Fatalf("Disconnect() state = %q, want %q", output.Connection.State, "disconnected")
	}
}

func TestServiceDisconnectActiveConnection(t *testing.T) {
	t.Parallel()

	connection := mustVKConnection(
		t,
		"vk_123",
		timePtr(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)),
		time.Date(2026, 4, 19, 12, 30, 0, 0, time.UTC),
	)
	importedInterest, err := vkintegrationdomain.NewImportedInterest(
		"Books",
		"vk_group",
		1,
		time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewImportedInterest() error = %v", err)
	}

	repo := &fakeVKConnectionRepository{
		connection: &connection,
		interests:  []vkintegrationdomain.ImportedInterest{importedInterest},
	}
	service := mustVKService(t, vkServiceDeps{
		repo:       repo,
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true},
	})

	output, err := service.Disconnect(context.Background(), DisconnectInput{UserID: testVKUserID})
	if err != nil {
		t.Fatalf("Disconnect() error = %v", err)
	}

	if output.Connection.State != "disconnected" {
		t.Fatalf("Disconnect() state = %q, want disconnected", output.Connection.State)
	}
	if repo.connection == nil || repo.connection.ConnectionState() != vkintegrationdomain.ConnectionStateDisconnected {
		t.Fatalf("repo connection state = %v, want disconnected", repo.connection.ConnectionState())
	}
	if len(repo.interests) != 0 {
		t.Fatalf("repo interests len = %d, want 0", len(repo.interests))
	}
	if len(repo.connection.Scopes()) != 0 {
		t.Fatalf("repo scopes = %v, want empty", repo.connection.Scopes())
	}
}

func TestServiceSyncInterestsRejectsWithoutConsent(t *testing.T) {
	t.Parallel()

	connection := mustVKRestoredConnection(t,
		"vk_123",
		string(vkintegrationdomain.ConnectionStateConnected),
		string(vkintegrationdomain.ConsentStateRevoked),
		"sealed-token",
		timePtr(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)),
		nil,
	)
	service := mustVKService(t, vkServiceDeps{
		repo:       &fakeVKConnectionRepository{connection: &connection},
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, opened: "raw-token"},
	})

	_, err := service.SyncInterests(context.Background(), SyncInterestsInput{UserID: testVKUserID})
	if err == nil {
		t.Fatal("SyncInterests() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "vk_consent_required" {
		t.Fatalf("SyncInterests() code = %q, want %q", appErr.Code(), "vk_consent_required")
	}
}

func TestServiceSyncInterestsRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	connection := mustVKConnection(t, "vk_123", timePtr(time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC)), time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC))
	service := mustVKService(t, vkServiceDeps{
		repo:       &fakeVKConnectionRepository{connection: &connection},
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, opened: "raw-token"},
	})

	_, err := service.SyncInterests(context.Background(), SyncInterestsInput{UserID: testVKUserID})
	if err == nil {
		t.Fatal("SyncInterests() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "vk_token_expired" {
		t.Fatalf("SyncInterests() code = %q, want %q", appErr.Code(), "vk_token_expired")
	}
}

func TestServiceSyncInterestsSucceedsWithoutGroupsScope(t *testing.T) {
	t.Parallel()

	connection := mustVKRestoredConnection(
		t,
		"vk_123",
		string(vkintegrationdomain.ConnectionStateConnected),
		string(vkintegrationdomain.ConsentStateGranted),
		"sealed-token",
		timePtr(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)),
		[]string{"vkid.personal_info"},
	)
	repo := &fakeVKConnectionRepository{connection: &connection}
	screenName := "Иван Иванов"
	service := mustVKService(t, vkServiceDeps{
		repo:       repo,
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, opened: "raw-token"},
		importer: fakeVKInterestImporter{
			result: ImportInterestsResult{
				Interests:         []ImportedInterestRecord{},
				ProfileScreenName: &screenName,
			},
		},
	})

	output, err := service.SyncInterests(context.Background(), SyncInterestsInput{UserID: testVKUserID})
	if err != nil {
		t.Fatalf("SyncInterests() error = %v", err)
	}
	if output.Connection.State != "connected" {
		t.Fatalf("SyncInterests() state = %q, want connected", output.Connection.State)
	}
	if repo.connection.Metadata().ScreenName() == nil || *repo.connection.Metadata().ScreenName() != screenName {
		t.Fatalf("SyncInterests() screen_name = %v, want %q", repo.connection.Metadata().ScreenName(), screenName)
	}
}

func TestServiceSyncInterestsSucceedsWithEmptyResult(t *testing.T) {
	t.Parallel()

	connection := mustVKConnection(t, "vk_123", timePtr(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)), time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC))
	repo := &fakeVKConnectionRepository{connection: &connection}
	service := mustVKService(t, vkServiceDeps{
		repo:       repo,
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, opened: "raw-token"},
		importer:   fakeVKInterestImporter{result: ImportInterestsResult{}},
	})

	output, err := service.SyncInterests(context.Background(), SyncInterestsInput{UserID: testVKUserID})
	if err != nil {
		t.Fatalf("SyncInterests() error = %v", err)
	}

	if output.Connection.State != "connected" {
		t.Fatalf("SyncInterests() state = %q, want %q", output.Connection.State, "connected")
	}
	if output.Connection.Sync.ImportedInterestsCount != 0 {
		t.Fatalf("SyncInterests() imported count = %d, want 0", output.Connection.Sync.ImportedInterestsCount)
	}
	if repo.connection.LastSyncState() != vkintegrationdomain.SyncStateSucceeded {
		t.Fatalf("SyncInterests() last sync state = %q, want %q", repo.connection.LastSyncState(), vkintegrationdomain.SyncStateSucceeded)
	}
}

func TestServiceSyncInterestsMarksFailureOnImporterError(t *testing.T) {
	t.Parallel()

	connection := mustVKConnection(t, "vk_123", timePtr(time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)), time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC))
	repo := &fakeVKConnectionRepository{connection: &connection}
	service := mustVKService(t, vkServiceDeps{
		repo:       repo,
		userReader: fakeVKUserReader{user: mustVKUser(t)},
		protector:  fakeVKTokenProtector{configured: true, opened: "raw-token"},
		importer:   fakeVKInterestImporter{err: ErrInterestImportUnavailable},
	})

	_, err := service.SyncInterests(context.Background(), SyncInterestsInput{UserID: testVKUserID})
	if err == nil {
		t.Fatal("SyncInterests() expected unavailable error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "vk_interest_import_unavailable" {
		t.Fatalf("SyncInterests() code = %q, want %q", appErr.Code(), "vk_interest_import_unavailable")
	}
	if repo.connection.LastSyncState() != vkintegrationdomain.SyncStateFailed {
		t.Fatalf("SyncInterests() last sync state = %q, want %q", repo.connection.LastSyncState(), vkintegrationdomain.SyncStateFailed)
	}
}

func TestServiceGetCurrentConnectionWhenFeatureDisabled(t *testing.T) {
	t.Parallel()

	service := mustVKService(t, vkServiceDeps{
		repo:            &fakeVKConnectionRepository{},
		userReader:      fakeVKUserReader{user: mustVKUser(t)},
		protector:       fakeVKTokenProtector{configured: true},
		featureDisabled: true,
	})

	output, err := service.GetCurrentConnection(context.Background(), GetCurrentConnectionInput{
		UserID: testVKUserID,
	})
	if err != nil {
		t.Fatalf("GetCurrentConnection() error = %v", err)
	}
	if output.Connection.FeatureEnabled {
		t.Fatal("GetCurrentConnection() feature_enabled = true, want false")
	}
	if !output.Connection.TokenStorageConfigured {
		t.Fatal("GetCurrentConnection() token_storage_configured = false, want true")
	}
	if output.Connection.State != "disconnected" {
		t.Fatalf("GetCurrentConnection() state = %q, want %q", output.Connection.State, "disconnected")
	}
}

func TestServiceConnectRejectsWhenFeatureDisabled(t *testing.T) {
	t.Parallel()

	service := mustVKService(t, vkServiceDeps{
		repo:            &fakeVKConnectionRepository{},
		userReader:      fakeVKUserReader{user: mustVKUser(t)},
		protector:       fakeVKTokenProtector{},
		featureDisabled: true,
	})

	_, err := service.Connect(context.Background(), ConnectInput{
		UserID:         testVKUserID,
		ProviderUserID: "vk_123",
		Consent: ConsentInput{
			Granted:    true,
			Version:    "v1",
			ObtainedAt: time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
		},
	})
	if err == nil {
		t.Fatal("Connect() expected unavailable error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "vk_integration_disabled" {
		t.Fatalf("Connect() code = %q, want %q", appErr.Code(), "vk_integration_disabled")
	}
}

type vkServiceDeps struct {
	repo            *fakeVKConnectionRepository
	userReader      fakeVKUserReader
	protector       fakeVKTokenProtector
	importer        fakeVKInterestImporter
	featureDisabled bool
}

func mustVKService(t *testing.T, deps vkServiceDeps) *Service {
	t.Helper()

	repo := deps.repo
	if repo == nil {
		repo = &fakeVKConnectionRepository{}
	}

	service, err := NewService(
		repo,
		deps.userReader,
		deps.protector,
		deps.importer,
		fakeVKConnectionIDGenerator{id: testVKConnectionID},
		nil,
		"",
		"",
		!deps.featureDisabled,
		2*time.Second,
		fakeVKClock{now: time.Date(2026, 4, 19, 12, 30, 0, 0, time.UTC)},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

type fakeVKConnectionRepository struct {
	connection *vkintegrationdomain.Connection
	interests  []vkintegrationdomain.ImportedInterest
	err        error
}

func (r *fakeVKConnectionRepository) GetByUserID(context.Context, userdomain.UserID) (*vkintegrationdomain.Connection, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.connection == nil {
		return nil, nil
	}

	connection := *r.connection
	return &connection, nil
}

func (r *fakeVKConnectionRepository) Save(_ context.Context, connection *vkintegrationdomain.Connection) error {
	if r.err != nil {
		return r.err
	}

	cloned := *connection
	r.connection = &cloned
	return nil
}

func (r *fakeVKConnectionRepository) ListImportedInterests(context.Context, vkintegrationdomain.ConnectionID) ([]vkintegrationdomain.ImportedInterest, error) {
	if r.err != nil {
		return nil, r.err
	}

	return append([]vkintegrationdomain.ImportedInterest(nil), r.interests...), nil
}

func (r *fakeVKConnectionRepository) ReplaceImportedInterests(ctx context.Context, connection *vkintegrationdomain.Connection, interests []vkintegrationdomain.ImportedInterest) error {
	if err := r.Save(ctx, connection); err != nil {
		return err
	}

	r.interests = append([]vkintegrationdomain.ImportedInterest(nil), interests...)
	return nil
}

type fakeVKUserReader struct {
	user *userdomain.User
	err  error
}

func (r fakeVKUserReader) GetByID(context.Context, userdomain.UserID) (*userdomain.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.user, nil
}

type fakeVKTokenProtector struct {
	configured bool
	sealed     string
	opened     string
	sealErr    error
	openErr    error
}

func (p fakeVKTokenProtector) Configured() bool {
	return p.configured
}

func (p fakeVKTokenProtector) Seal(plain string) (string, error) {
	if p.sealErr != nil {
		return "", p.sealErr
	}
	if p.sealed != "" {
		return p.sealed, nil
	}

	return "sealed:" + plain, nil
}

func (p fakeVKTokenProtector) Open(string) (string, error) {
	if p.openErr != nil {
		return "", p.openErr
	}
	if p.opened != "" {
		return p.opened, nil
	}

	return "opened", nil
}

type fakeVKInterestImporter struct {
	result ImportInterestsResult
	err    error
}

func (i fakeVKInterestImporter) ImportInterests(context.Context, ImportInterestsRequest) (ImportInterestsResult, error) {
	if i.err != nil {
		return ImportInterestsResult{}, i.err
	}

	return i.result, nil
}

type fakeVKConnectionIDGenerator struct {
	id string
}

func (g fakeVKConnectionIDGenerator) NewVKConnectionID() (vkintegrationdomain.ConnectionID, error) {
	return vkintegrationdomain.NewConnectionID(g.id)
}

type fakeVKClock struct {
	now time.Time
}

func (c fakeVKClock) Now() time.Time {
	return c.now
}

func mustVKUser(t *testing.T) *userdomain.User {
	t.Helper()

	user, err := userdomain.NewUser(testVKUserID, "vk-user@example.com", "Password123!", "user")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	return &user
}

func mustVKConnection(
	t *testing.T,
	providerUserID string,
	tokenExpiresAt *time.Time,
	now time.Time,
) vkintegrationdomain.Connection {
	t.Helper()

	connectionID, err := vkintegrationdomain.NewConnectionID(testVKConnectionID)
	if err != nil {
		t.Fatalf("NewConnectionID() error = %v", err)
	}
	userID, err := userdomain.NewUserID(testVKUserID)
	if err != nil {
		t.Fatalf("NewUserID() error = %v", err)
	}
	metadata, err := vkintegrationdomain.NewIntegrationMetadata(stringPtr("ilya"), nil)
	if err != nil {
		t.Fatalf("NewIntegrationMetadata() error = %v", err)
	}
	connection, err := vkintegrationdomain.NewConnection(
		connectionID,
		userID,
		providerUserID,
		"v1",
		stringPtr("sealed-token"),
		tokenExpiresAt,
		[]string{"groups"},
		metadata,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("NewConnection() error = %v", err)
	}

	return connection
}

func mustVKRestoredConnection(
	t *testing.T,
	providerUserID string,
	connectionState string,
	consentState string,
	tokenCiphertext string,
	tokenExpiresAt *time.Time,
	scopes []string,
) vkintegrationdomain.Connection {
	t.Helper()

	if scopes == nil {
		scopes = []string{"groups"}
	}

	metadata, err := vkintegrationdomain.NewIntegrationMetadata(stringPtr("ilya"), nil)
	if err != nil {
		t.Fatalf("NewIntegrationMetadata() error = %v", err)
	}

	connection, err := vkintegrationdomain.RestoreConnection(
		testVKConnectionID,
		testVKUserID,
		providerUserID,
		connectionState,
		consentState,
		"v1",
		&tokenCiphertext,
		tokenExpiresAt,
		scopes,
		metadata,
		string(vkintegrationdomain.SyncStateIdle),
		nil,
		nil,
		timePtr(time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)),
		nil,
		timePtr(time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)),
		nil,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreConnection() error = %v", err)
	}

	return connection
}

func stringValuePtr(value string) *string {
	current := value
	return &current
}

func timePtr(value time.Time) *time.Time {
	current := value
	return &current
}
