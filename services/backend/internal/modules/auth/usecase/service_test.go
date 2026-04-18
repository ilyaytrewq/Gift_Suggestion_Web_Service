package usecase

import (
	"context"
	"testing"
	"time"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testUserID            = "550e8400-e29b-41d4-a716-446655440000"
	testSessionID         = "550e8400-e29b-41d4-a716-446655440001"
	testResetTokenID      = "550e8400-e29b-41d4-a716-446655440002"
	testEmail             = "user@example.com"
	testPassword          = "ValidPass1!"
	testAlternatePassword = "AnotherPass1!"
	testAccessToken       = "access-token"
	testRefreshToken      = "refresh-token"
	testRefreshTokenHash  = "refresh-token-hash"
	testNextRefreshToken  = "next-refresh-token"
	testNextRefreshHash   = "next-refresh-token-hash"
	testResetToken        = "reset-token"
	testResetTokenHash    = "reset-token-hash"
)

func TestServiceRegisterSuccess(t *testing.T) {
	t.Parallel()

	clock := fixedClock{now: time.Date(2026, 4, 18, 11, 0, 0, 0, time.UTC)}
	userRepo := newFakeUserRepository()

	service := mustAuthService(t, userRepo, newFakeSessionRepository(), newFakePasswordResetRepository(), clock)

	output, err := service.Register(context.Background(), RegisterInput{
		Email:       " USER@example.com ",
		Password:    testPassword,
		DisplayName: "  Alice  ",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if output.User.Email != testEmail {
		t.Fatalf("Register() email = %q, want %q", output.User.Email, testEmail)
	}
	if output.User.DisplayName != "Alice" {
		t.Fatalf("Register() display name = %q, want %q", output.User.DisplayName, "Alice")
	}
	if len(userRepo.savedUsers) != 1 {
		t.Fatalf("expected one saved user, got %d", len(userRepo.savedUsers))
	}
	if userRepo.savedUsers[0].PasswordHash().String() == testPassword {
		t.Fatal("expected hashed password to differ from plaintext password")
	}
}

func TestServiceRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	existing := mustUser(t)
	userRepo := newFakeUserRepository()
	userRepo.usersByEmail[testEmail] = existing

	service := mustAuthService(t, userRepo, newFakeSessionRepository(), newFakePasswordResetRepository(), fixedClock{now: time.Now().UTC()})

	_, err := service.Register(context.Background(), RegisterInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err == nil {
		t.Fatal("Register() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindConflict {
		t.Fatalf("Register() error kind = %q, want %q", appErr.Kind(), apperrors.KindConflict)
	}
	if appErr.Code() != "email_already_exists" {
		t.Fatalf("Register() error code = %q, want %q", appErr.Code(), "email_already_exists")
	}
}

func TestServiceLoginSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 12, 30, 0, 0, time.UTC)
	user := mustUser(t)

	userRepo := newFakeUserRepository()
	userRepo.usersByEmail[testEmail] = user
	userRepo.usersByID[testUserID] = user

	sessionRepo := newFakeSessionRepository()
	resetRepo := newFakePasswordResetRepository()

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: resetRepo,
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		clock:           fixedClock{now: now},
		refreshTokenTTL: 7 * 24 * time.Hour,
		resetTokenTTL:   30 * time.Minute,
	})

	output, err := service.Login(context.Background(), LoginInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if output.Auth.AccessToken != testAccessToken {
		t.Fatalf("Login() access token = %q, want %q", output.Auth.AccessToken, testAccessToken)
	}
	if output.Auth.RefreshToken != testRefreshToken {
		t.Fatalf("Login() refresh token = %q, want %q", output.Auth.RefreshToken, testRefreshToken)
	}
	if output.User.LastLoginAt == nil || !output.User.LastLoginAt.Equal(now) {
		t.Fatalf("Login() last login = %v, want %v", output.User.LastLoginAt, now)
	}
	if len(sessionRepo.savedSessions) != 1 {
		t.Fatalf("expected one saved session, got %d", len(sessionRepo.savedSessions))
	}
	if len(userRepo.markLastLoginCalls) != 1 {
		t.Fatalf("expected one mark last login call, got %d", len(userRepo.markLastLoginCalls))
	}
	if !userRepo.markLastLoginCalls[0].at.Equal(now) {
		t.Fatalf("MarkLastLogin() timestamp = %v, want %v", userRepo.markLastLoginCalls[0].at, now)
	}
}

func TestServiceRefreshSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 13, 0, 0, 0, time.UTC)
	user := mustUser(t)
	session := mustSession(t, testSessionID, testUserID, testRefreshTokenHash, now.Add(-time.Hour), now.Add(time.Hour))

	userRepo := newFakeUserRepository()
	userRepo.usersByID[testUserID] = user

	sessionRepo := newFakeSessionRepository()
	sessionRepo.sessionsByHash[testRefreshTokenHash] = &session

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: newFakePasswordResetRepository(),
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testNextRefreshToken, hash: testNextRefreshHash}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		},
		clock:           fixedClock{now: now},
		refreshTokenTTL: 7 * 24 * time.Hour,
		resetTokenTTL:   30 * time.Minute,
	})
	service.refreshTokenGenerator = &fakeTokenGenerator{
		results: []tokenResult{{raw: testNextRefreshToken, hash: testNextRefreshHash}},
		hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
	}

	output, err := service.Refresh(context.Background(), RefreshInput{RefreshToken: testRefreshToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if output.Auth.AccessToken != testAccessToken {
		t.Fatalf("Refresh() access token = %q, want %q", output.Auth.AccessToken, testAccessToken)
	}
	if output.Auth.RefreshToken != testNextRefreshToken {
		t.Fatalf("Refresh() refresh token = %q, want %q", output.Auth.RefreshToken, testNextRefreshToken)
	}
	if len(sessionRepo.updatedSessions) != 1 {
		t.Fatalf("expected one updated session, got %d", len(sessionRepo.updatedSessions))
	}
	if sessionRepo.updatedSessions[0].RefreshTokenHash() != testNextRefreshHash {
		t.Fatalf("updated session refresh token hash = %q, want %q", sessionRepo.updatedSessions[0].RefreshTokenHash(), testNextRefreshHash)
	}
	if sessionRepo.updatedSessions[0].LastUsedAt() == nil || !sessionRepo.updatedSessions[0].LastUsedAt().Equal(now) {
		t.Fatalf("updated session last used at = %v, want %v", sessionRepo.updatedSessions[0].LastUsedAt(), now)
	}
}

func TestServiceRequestPasswordResetStoresHashedToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 14, 0, 0, 0, time.UTC)
	user := mustUser(t)

	userRepo := newFakeUserRepository()
	userRepo.usersByEmail[testEmail] = user

	resetRepo := newFakePasswordResetRepository()
	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:          userRepo,
		sessionRepo:       newFakeSessionRepository(),
		passwordResetRepo: resetRepo,
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		clock:           fixedClock{now: now},
		refreshTokenTTL: 7 * 24 * time.Hour,
		resetTokenTTL:   45 * time.Minute,
	})

	output, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Email: testEmail})
	if err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}

	if !output.Accepted {
		t.Fatal("RequestPasswordReset() accepted = false, want true")
	}
	if len(resetRepo.savedTokens) != 1 {
		t.Fatalf("expected one saved reset token, got %d", len(resetRepo.savedTokens))
	}
	if resetRepo.savedTokens[0].TokenHash() != testResetTokenHash {
		t.Fatalf("saved reset token hash = %q, want %q", resetRepo.savedTokens[0].TokenHash(), testResetTokenHash)
	}
	if !resetRepo.savedTokens[0].ExpiresAt().Equal(now.Add(45 * time.Minute)) {
		t.Fatalf("saved reset token expiry = %v, want %v", resetRepo.savedTokens[0].ExpiresAt(), now.Add(45*time.Minute))
	}
}

type authServiceDeps struct {
	userRepo              *fakeUserRepository
	sessionRepo           *fakeSessionRepository
	passwordResetRepo     *fakePasswordResetRepository
	accessTokenManager    *fakeAccessTokenManager
	refreshTokenGenerator *fakeTokenGenerator
	resetTokenGenerator   *fakeTokenGenerator
	clock                 fixedClock
	refreshTokenTTL       time.Duration
	resetTokenTTL         time.Duration
}

func mustAuthService(t *testing.T, userRepo *fakeUserRepository, sessionRepo *fakeSessionRepository, resetRepo *fakePasswordResetRepository, clock fixedClock) *Service {
	t.Helper()

	return mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		passwordResetRepo: resetRepo,
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		clock:           clock,
		refreshTokenTTL: 7 * 24 * time.Hour,
		resetTokenTTL:   30 * time.Minute,
	})
}

func mustAuthServiceWithDeps(t *testing.T, deps authServiceDeps) *Service {
	t.Helper()

	service, err := NewService(
		deps.userRepo,
		deps.sessionRepo,
		deps.passwordResetRepo,
		deps.accessTokenManager,
		deps.refreshTokenGenerator,
		deps.resetTokenGenerator,
		fakeUserIDGenerator{id: testUserID},
		fakeSessionIDGenerator{id: testSessionID},
		fakePasswordResetTokenIDGenerator{id: testResetTokenID},
		deps.refreshTokenTTL,
		deps.resetTokenTTL,
		deps.clock,
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

type fakeUserRepository struct {
	usersByID          map[string]*userdomain.User
	usersByEmail       map[string]*userdomain.User
	savedUsers         []*userdomain.User
	markLastLoginCalls []markLastLoginCall
}

type markLastLoginCall struct {
	id userdomain.UserID
	at time.Time
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		usersByID:    make(map[string]*userdomain.User),
		usersByEmail: make(map[string]*userdomain.User),
	}
}

func (r *fakeUserRepository) Save(_ context.Context, user *userdomain.User) error {
	r.savedUsers = append(r.savedUsers, user)
	r.usersByID[user.ID().String()] = user
	r.usersByEmail[user.Email().String()] = user
	return nil
}

func (r *fakeUserRepository) GetByID(_ context.Context, id userdomain.UserID) (*userdomain.User, error) {
	return r.usersByID[id.String()], nil
}

func (r *fakeUserRepository) GetByEmail(_ context.Context, email userdomain.Email) (*userdomain.User, error) {
	return r.usersByEmail[email.String()], nil
}

func (r *fakeUserRepository) MarkLastLogin(_ context.Context, id userdomain.UserID, at time.Time) error {
	r.markLastLoginCalls = append(r.markLastLoginCalls, markLastLoginCall{id: id, at: at.UTC()})
	return nil
}

type fakeSessionRepository struct {
	savedSessions   []*authdomain.Session
	updatedSessions []*authdomain.Session
	sessionsByHash  map[string]*authdomain.Session
}

func newFakeSessionRepository() *fakeSessionRepository {
	return &fakeSessionRepository{sessionsByHash: make(map[string]*authdomain.Session)}
}

func (r *fakeSessionRepository) Save(_ context.Context, session *authdomain.Session) error {
	r.savedSessions = append(r.savedSessions, session)
	r.sessionsByHash[session.RefreshTokenHash()] = session
	return nil
}

func (r *fakeSessionRepository) GetByRefreshTokenHash(_ context.Context, tokenHash string) (*authdomain.Session, error) {
	session, ok := r.sessionsByHash[tokenHash]
	if !ok {
		return nil, authdomain.ErrSessionNotFound
	}
	return session, nil
}

func (r *fakeSessionRepository) Update(_ context.Context, session *authdomain.Session) error {
	r.updatedSessions = append(r.updatedSessions, session)
	r.sessionsByHash[session.RefreshTokenHash()] = session
	return nil
}

type fakePasswordResetRepository struct {
	savedTokens []*authdomain.PasswordResetToken
}

func newFakePasswordResetRepository() *fakePasswordResetRepository {
	return &fakePasswordResetRepository{}
}

func (r *fakePasswordResetRepository) Save(_ context.Context, token *authdomain.PasswordResetToken) error {
	r.savedTokens = append(r.savedTokens, token)
	return nil
}

type fakeAccessTokenManager struct {
	token string
	ttl   time.Duration
}

func (m *fakeAccessTokenManager) IssueToken(Actor, time.Time) (string, error) {
	return m.token, nil
}

func (m *fakeAccessTokenManager) ParseToken(string) (Actor, error) {
	return Actor{}, nil
}

func (m *fakeAccessTokenManager) TokenTTL() time.Duration {
	return m.ttl
}

type tokenResult struct {
	raw  string
	hash string
}

type fakeTokenGenerator struct {
	results []tokenResult
	hashes  map[string]string
}

func (g *fakeTokenGenerator) NewToken() (string, string, error) {
	if len(g.results) == 0 {
		return "", "", nil
	}

	result := g.results[0]
	g.results = g.results[1:]
	return result.raw, result.hash, nil
}

func (g *fakeTokenGenerator) Hash(rawToken string) string {
	if hash, ok := g.hashes[rawToken]; ok {
		return hash
	}
	return rawToken
}

type fakeUserIDGenerator struct {
	id string
}

func (g fakeUserIDGenerator) NewUserID() (userdomain.UserID, error) {
	return userdomain.NewUserID(g.id)
}

type fakeSessionIDGenerator struct {
	id string
}

func (g fakeSessionIDGenerator) NewSessionID() (authdomain.SessionID, error) {
	return authdomain.NewSessionID(g.id)
}

type fakePasswordResetTokenIDGenerator struct {
	id string
}

func (g fakePasswordResetTokenIDGenerator) NewPasswordResetTokenID() (authdomain.PasswordResetTokenID, error) {
	return authdomain.NewPasswordResetTokenID(g.id)
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now.UTC()
}

func mustUser(t *testing.T) *userdomain.User {
	t.Helper()

	user, err := userdomain.NewUser(testUserID, testEmail, testPassword, string(userdomain.UserRoleUser))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	return &user
}

func mustSession(t *testing.T, id, userID, refreshTokenHash string, now, expiresAt time.Time) authdomain.Session {
	t.Helper()

	sessionID, err := authdomain.NewSessionID(id)
	if err != nil {
		t.Fatalf("NewSessionID() error = %v", err)
	}
	uid, err := userdomain.NewUserID(userID)
	if err != nil {
		t.Fatalf("NewUserID() error = %v", err)
	}

	session, err := authdomain.NewSession(sessionID, uid, refreshTokenHash, now, expiresAt)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	return session
}
