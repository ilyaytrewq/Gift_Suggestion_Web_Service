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
	registrationRepo := newFakeRegistrationRepository()
	emailVerificationRepo := newFakeEmailVerificationRepository()
	notifier := &fakeAuthEmailNotifier{}

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:                 userRepo,
		registrationRepo:         registrationRepo,
		emailVerificationRepo:    emailVerificationRepo,
		sessionRepo:              newFakeSessionRepository(),
		passwordResetRepo:        newFakePasswordResetRepository(),
		emailNotifier:            notifier,
		logger:                   &fakeLogger{},
		clock:                    clock,
		refreshTokenTTL:          7 * 24 * time.Hour,
		verificationTokenTTL:     24 * time.Hour,
		resetTokenTTL:            30 * time.Minute,
		verificationTokenResults: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
	})

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
	if len(registrationRepo.savedUsers) != 1 {
		t.Fatalf("expected one saved user, got %d", len(registrationRepo.savedUsers))
	}
	if registrationRepo.savedUsers[0].PasswordHash().String() == testPassword {
		t.Fatal("expected hashed password to differ from plaintext password")
	}
	if len(registrationRepo.savedVerificationTokens) != 1 {
		t.Fatalf("expected one verification token, got %d", len(registrationRepo.savedVerificationTokens))
	}
	if len(notifier.verificationCalls) != 1 || notifier.verificationCalls[0].rawToken != "verify-token" {
		t.Fatalf("expected verification email with raw token, got %+v", notifier.verificationCalls)
	}
}

func TestServiceRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	existing := mustUser(t)
	userRepo := newFakeUserRepository()
	userRepo.usersByEmail[testEmail] = existing

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:                 userRepo,
		registrationRepo:         newFakeRegistrationRepository(),
		emailVerificationRepo:    newFakeEmailVerificationRepository(),
		sessionRepo:              newFakeSessionRepository(),
		passwordResetRepo:        newFakePasswordResetRepository(),
		emailNotifier:            &fakeAuthEmailNotifier{},
		logger:                   &fakeLogger{},
		clock:                    fixedClock{now: time.Now().UTC()},
		refreshTokenTTL:          7 * 24 * time.Hour,
		verificationTokenTTL:     24 * time.Hour,
		resetTokenTTL:            30 * time.Minute,
		verificationTokenResults: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
	})

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
	user.MarkEmailVerified(now)

	userRepo := newFakeUserRepository()
	userRepo.usersByEmail[testEmail] = user
	userRepo.usersByID[testUserID] = user

	sessionRepo := newFakeSessionRepository()
	resetRepo := newFakePasswordResetRepository()

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           sessionRepo,
		passwordResetRepo:     resetRepo,
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
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

func TestServiceLoginRejectsUnverifiedEmail(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 12, 30, 0, 0, time.UTC)
	user := mustUser(t)

	userRepo := newFakeUserRepository()
	userRepo.usersByEmail[testEmail] = user
	userRepo.usersByID[testUserID] = user

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           newFakeSessionRepository(),
		passwordResetRepo:     newFakePasswordResetRepository(),
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})

	_, err := service.Login(context.Background(), LoginInput{
		Email:    testEmail,
		Password: testPassword,
	})
	if err == nil {
		t.Fatal("Login() expected error for unverified email")
	}
	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindForbidden {
		t.Fatalf("Login() kind = %q, want %q", appErr.Kind(), apperrors.KindForbidden)
	}
	if appErr.Code() != "email_not_verified" {
		t.Fatalf("Login() code = %q, want %q", appErr.Code(), "email_not_verified")
	}
}

func TestServiceRefreshRevokesSessionWhenEmailUnverified(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 13, 0, 0, 0, time.UTC)
	user := mustUser(t)
	session := mustSession(t, testSessionID, testUserID, testRefreshTokenHash, now.Add(-time.Hour), now.Add(time.Hour))

	userRepo := newFakeUserRepository()
	userRepo.usersByID[testUserID] = user

	sessionRepo := newFakeSessionRepository()
	sessionPtr := session
	sessionRepo.sessionsByHash[testRefreshTokenHash] = &sessionPtr

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           sessionRepo,
		passwordResetRepo:     newFakePasswordResetRepository(),
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testNextRefreshToken, hash: testNextRefreshHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})

	_, err := service.Refresh(context.Background(), RefreshInput{RefreshToken: testRefreshToken})
	if err == nil {
		t.Fatal("Refresh() expected error for unverified email")
	}
	appErr := apperrors.From(err)
	if appErr.Code() != "email_not_verified" {
		t.Fatalf("Refresh() code = %q, want %q", appErr.Code(), "email_not_verified")
	}
	if len(sessionRepo.updatedSessions) != 1 || !sessionRepo.updatedSessions[0].IsRevoked() {
		t.Fatalf("expected revoked session update, got %+v", sessionRepo.updatedSessions)
	}
}

func TestServiceAuthorizeRejectsUnverifiedEmail(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 14, 0, 0, 0, time.UTC)
	user := mustUser(t)

	userRepo := newFakeUserRepository()
	userRepo.usersByID[testUserID] = user

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           newFakeSessionRepository(),
		passwordResetRepo:     newFakePasswordResetRepository(),
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})

	_, err := service.Authorize(context.Background(), "any-access-token")
	if err == nil {
		t.Fatal("Authorize() expected error")
	}
	if apperrors.From(err).Code() != "email_not_verified" {
		t.Fatalf("Authorize() code = %q", apperrors.From(err).Code())
	}

	user.MarkEmailVerified(now)
	userRepo.usersByID[testUserID] = user

	actor, err := service.Authorize(context.Background(), "any-access-token")
	if err != nil {
		t.Fatalf("Authorize() error = %v", err)
	}
	if actor.UserID != testUserID {
		t.Fatalf("actor user id = %q", actor.UserID)
	}
}

func TestServiceRefreshSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 13, 0, 0, 0, time.UTC)
	user := mustUser(t)
	user.MarkEmailVerified(now)
	session := mustSession(t, testSessionID, testUserID, testRefreshTokenHash, now.Add(-time.Hour), now.Add(time.Hour))

	userRepo := newFakeUserRepository()
	userRepo.usersByID[testUserID] = user

	sessionRepo := newFakeSessionRepository()
	sessionRepo.sessionsByHash[testRefreshTokenHash] = &session

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           sessionRepo,
		passwordResetRepo:     newFakePasswordResetRepository(),
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testNextRefreshToken, hash: testNextRefreshHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
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
	notifier := &fakeAuthEmailNotifier{}
	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           newFakeSessionRepository(),
		passwordResetRepo:     resetRepo,
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        notifier,
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        45 * time.Minute,
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
	if len(notifier.resetCalls) != 1 || notifier.resetCalls[0].rawToken != testResetToken {
		t.Fatalf("expected reset email with raw token, got %+v", notifier.resetCalls)
	}
}

func TestServiceConfirmEmailVerificationConsumesHashedToken(t *testing.T) {
	t.Parallel()

	verificationRepo := newFakeEmailVerificationRepository()
	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              newFakeUserRepository(),
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: verificationRepo,
		sessionRepo:           newFakeSessionRepository(),
		passwordResetRepo:     newFakePasswordResetRepository(),
		verificationTokenGenerator: &fakeTokenGenerator{
			hashes: map[string]string{"verify-token": "verify-token-hash"},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: time.Now().UTC()},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})

	output, err := service.ConfirmEmailVerification(context.Background(), ConfirmEmailVerificationInput{Token: "verify-token"})
	if err != nil {
		t.Fatalf("ConfirmEmailVerification() error = %v", err)
	}
	if !output.Accepted {
		t.Fatal("ConfirmEmailVerification() accepted = false, want true")
	}
	if len(verificationRepo.consumed) != 1 || verificationRepo.consumed[0] != "verify-token-hash" {
		t.Fatalf("expected consumed verification token hash, got %+v", verificationRepo.consumed)
	}
}

func TestServiceConfirmPasswordResetConsumesHashedToken(t *testing.T) {
	t.Parallel()

	resetRepo := newFakePasswordResetRepository()
	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              newFakeUserRepository(),
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           newFakeSessionRepository(),
		passwordResetRepo:     resetRepo,
		resetTokenGenerator: &fakeTokenGenerator{
			hashes: map[string]string{"reset-token": "reset-token-hash"},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: time.Now().UTC()},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})

	output, err := service.ConfirmPasswordReset(context.Background(), ConfirmPasswordResetInput{
		Token:       "reset-token",
		NewPassword: testAlternatePassword,
	})
	if err != nil {
		t.Fatalf("ConfirmPasswordReset() error = %v", err)
	}
	if !output.Accepted {
		t.Fatal("ConfirmPasswordReset() accepted = false, want true")
	}
	if len(resetRepo.consumed) != 1 || resetRepo.consumed[0].tokenHash != "reset-token-hash" {
		t.Fatalf("expected consumed reset token hash, got %+v", resetRepo.consumed)
	}
	if resetRepo.consumed[0].newPasswordHash == testAlternatePassword {
		t.Fatal("expected stored password hash to differ from plaintext password")
	}
}

func TestServiceLogoutRevokesActiveSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 15, 0, 0, 0, time.UTC)
	session := mustSession(t, testSessionID, testUserID, testRefreshTokenHash, now.Add(-time.Hour), now.Add(time.Hour))

	sessionRepo := newFakeSessionRepository()
	sessionRepo.sessionsByHash[testRefreshTokenHash] = &session

	service := mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              newFakeUserRepository(),
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           sessionRepo,
		passwordResetRepo:     newFakePasswordResetRepository(),
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                fixedClock{now: now},
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})

	output, err := service.Logout(context.Background(), LogoutInput{RefreshToken: testRefreshToken})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if !output.Accepted {
		t.Fatal("Logout() accepted = false, want true")
	}
	if len(sessionRepo.updatedSessions) != 1 {
		t.Fatalf("expected one updated session, got %d", len(sessionRepo.updatedSessions))
	}
	if !sessionRepo.updatedSessions[0].IsRevoked() {
		t.Fatal("expected session to be revoked")
	}
	revokedAt := sessionRepo.updatedSessions[0].RevokedAt()
	if revokedAt == nil || !revokedAt.Equal(now) {
		t.Fatalf("revoked at = %v, want %v", revokedAt, now)
	}
}

func TestServiceLogoutMissingRefreshTokenAccepted(t *testing.T) {
	t.Parallel()

	sessionRepo := newFakeSessionRepository()
	service := mustAuthService(t, newFakeUserRepository(), sessionRepo, newFakePasswordResetRepository(), fixedClock{now: time.Now().UTC()})

	output, err := service.Logout(context.Background(), LogoutInput{RefreshToken: "   "})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if !output.Accepted {
		t.Fatal("Logout() accepted = false, want true")
	}
	if len(sessionRepo.updatedSessions) != 0 {
		t.Fatalf("expected no session updates, got %d", len(sessionRepo.updatedSessions))
	}
}

func TestServiceLogoutUnknownRefreshTokenAccepted(t *testing.T) {
	t.Parallel()

	sessionRepo := newFakeSessionRepository()
	service := mustAuthService(t, newFakeUserRepository(), sessionRepo, newFakePasswordResetRepository(), fixedClock{now: time.Now().UTC()})

	output, err := service.Logout(context.Background(), LogoutInput{RefreshToken: testRefreshToken})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if !output.Accepted {
		t.Fatal("Logout() accepted = false, want true")
	}
	if len(sessionRepo.updatedSessions) != 0 {
		t.Fatalf("expected no session updates, got %d", len(sessionRepo.updatedSessions))
	}
}

type authServiceDeps struct {
	userRepo                   *fakeUserRepository
	registrationRepo           *fakeRegistrationRepository
	emailVerificationRepo      *fakeEmailVerificationRepository
	sessionRepo                *fakeSessionRepository
	passwordResetRepo          *fakePasswordResetRepository
	accessTokenManager         *fakeAccessTokenManager
	refreshTokenGenerator      *fakeTokenGenerator
	verificationTokenGenerator *fakeTokenGenerator
	resetTokenGenerator        *fakeTokenGenerator
	emailNotifier              *fakeAuthEmailNotifier
	logger                     *fakeLogger
	clock                      fixedClock
	refreshTokenTTL            time.Duration
	verificationTokenTTL       time.Duration
	resetTokenTTL              time.Duration
	verificationTokenResults   []tokenResult
}

func mustAuthService(t *testing.T, userRepo *fakeUserRepository, sessionRepo *fakeSessionRepository, resetRepo *fakePasswordResetRepository, clock fixedClock) *Service {
	t.Helper()

	return mustAuthServiceWithDeps(t, authServiceDeps{
		userRepo:              userRepo,
		registrationRepo:      newFakeRegistrationRepository(),
		emailVerificationRepo: newFakeEmailVerificationRepository(),
		sessionRepo:           sessionRepo,
		passwordResetRepo:     resetRepo,
		accessTokenManager: &fakeAccessTokenManager{
			token: testAccessToken,
			ttl:   15 * time.Minute,
		},
		refreshTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		},
		verificationTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
		},
		resetTokenGenerator: &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
		},
		emailNotifier:        &fakeAuthEmailNotifier{},
		logger:               &fakeLogger{},
		clock:                clock,
		refreshTokenTTL:      7 * 24 * time.Hour,
		verificationTokenTTL: 24 * time.Hour,
		resetTokenTTL:        30 * time.Minute,
	})
}

func mustAuthServiceWithDeps(t *testing.T, deps authServiceDeps) *Service {
	t.Helper()

	if deps.registrationRepo == nil {
		deps.registrationRepo = newFakeRegistrationRepository()
	}
	if deps.emailVerificationRepo == nil {
		deps.emailVerificationRepo = newFakeEmailVerificationRepository()
	}
	if deps.accessTokenManager == nil {
		deps.accessTokenManager = &fakeAccessTokenManager{token: testAccessToken, ttl: 15 * time.Minute}
	}
	if deps.refreshTokenGenerator == nil {
		deps.refreshTokenGenerator = &fakeTokenGenerator{
			results: []tokenResult{{raw: testRefreshToken, hash: testRefreshTokenHash}},
			hashes:  map[string]string{testRefreshToken: testRefreshTokenHash},
		}
	}
	if deps.verificationTokenGenerator == nil {
		deps.verificationTokenGenerator = &fakeTokenGenerator{
			results: []tokenResult{{raw: "verify-token", hash: "verify-token-hash"}},
			hashes:  map[string]string{"verify-token": "verify-token-hash"},
		}
	}
	if deps.resetTokenGenerator == nil {
		deps.resetTokenGenerator = &fakeTokenGenerator{
			results: []tokenResult{{raw: testResetToken, hash: testResetTokenHash}},
			hashes:  map[string]string{testResetToken: testResetTokenHash},
		}
	}
	if deps.emailNotifier == nil {
		deps.emailNotifier = &fakeAuthEmailNotifier{}
	}
	if deps.logger == nil {
		deps.logger = &fakeLogger{}
	}

	service, err := NewService(
		deps.userRepo,
		deps.registrationRepo,
		deps.emailVerificationRepo,
		deps.sessionRepo,
		deps.passwordResetRepo,
		deps.accessTokenManager,
		deps.refreshTokenGenerator,
		deps.verificationTokenGenerator,
		deps.resetTokenGenerator,
		deps.emailNotifier,
		deps.logger,
		fakeUserIDGenerator{id: testUserID},
		fakeSessionIDGenerator{id: testSessionID},
		fakeEmailVerificationTokenIDGenerator{id: "550e8400-e29b-41d4-a716-446655440003"},
		fakePasswordResetTokenIDGenerator{id: testResetTokenID},
		deps.refreshTokenTTL,
		deps.verificationTokenTTL,
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

type fakeRegistrationRepository struct {
	savedUsers              []*userdomain.User
	savedVerificationTokens []*authdomain.EmailVerificationToken
}

func newFakeRegistrationRepository() *fakeRegistrationRepository {
	return &fakeRegistrationRepository{}
}

func (r *fakeRegistrationRepository) SaveUserWithVerificationToken(
	_ context.Context,
	user *userdomain.User,
	token *authdomain.EmailVerificationToken,
) error {
	r.savedUsers = append(r.savedUsers, user)
	r.savedVerificationTokens = append(r.savedVerificationTokens, token)
	return nil
}

type fakeEmailVerificationRepository struct {
	consumed []string
	err      error
}

func newFakeEmailVerificationRepository() *fakeEmailVerificationRepository {
	return &fakeEmailVerificationRepository{}
}

func (r *fakeEmailVerificationRepository) Consume(_ context.Context, tokenHash string, _ time.Time) error {
	r.consumed = append(r.consumed, tokenHash)
	return r.err
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
	consumed    []consumedReset
	err         error
}

type consumedReset struct {
	tokenHash       string
	newPasswordHash string
}

func newFakePasswordResetRepository() *fakePasswordResetRepository {
	return &fakePasswordResetRepository{}
}

func (r *fakePasswordResetRepository) Save(_ context.Context, token *authdomain.PasswordResetToken) error {
	r.savedTokens = append(r.savedTokens, token)
	return nil
}

func (r *fakePasswordResetRepository) Consume(
	_ context.Context,
	tokenHash string,
	newPasswordHash string,
	_ time.Time,
) error {
	r.consumed = append(r.consumed, consumedReset{
		tokenHash:       tokenHash,
		newPasswordHash: newPasswordHash,
	})
	return r.err
}

type fakeAccessTokenManager struct {
	token string
	ttl   time.Duration
}

func (m *fakeAccessTokenManager) IssueToken(Actor, time.Time) (string, error) {
	return m.token, nil
}

func (m *fakeAccessTokenManager) ParseToken(string) (Actor, error) {
	return Actor{
		UserID:    testUserID,
		SessionID: testSessionID,
		Role:      string(userdomain.UserRoleUser),
	}, nil
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

type fakeEmailVerificationTokenIDGenerator struct {
	id string
}

func (g fakeEmailVerificationTokenIDGenerator) NewEmailVerificationTokenID() (authdomain.EmailVerificationTokenID, error) {
	return authdomain.NewEmailVerificationTokenID(g.id)
}

type fakeAuthEmailNotifier struct {
	verificationCalls []emailCall
	resetCalls        []emailCall
	err               error
}

type emailCall struct {
	userID   string
	rawToken string
}

func (n *fakeAuthEmailNotifier) SendVerificationEmail(
	_ context.Context,
	user *userdomain.User,
	rawToken, _ string,
) error {
	n.verificationCalls = append(n.verificationCalls, emailCall{userID: user.ID().String(), rawToken: rawToken})
	return n.err
}

func (n *fakeAuthEmailNotifier) SendPasswordResetEmail(
	_ context.Context,
	user *userdomain.User,
	rawToken, _ string,
) error {
	n.resetCalls = append(n.resetCalls, emailCall{userID: user.ID().String(), rawToken: rawToken})
	return n.err
}

type fakeLogger struct {
	errors []string
}

func (l *fakeLogger) Error(msg string, _ ...any) {
	l.errors = append(l.errors, msg)
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
