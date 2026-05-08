package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	userusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type Service struct {
	userRepo                     UserRepository
	registrationRepo             RegistrationRepository
	emailVerificationRepo        EmailVerificationRepository
	sessionRepo                  SessionRepository
	passwordResetRepo            PasswordResetRepository
	accessTokenManager           AccessTokenManager
	refreshTokenGenerator        TokenGenerator
	verificationTokenGenerator   TokenGenerator
	resetTokenGenerator          TokenGenerator
	emailNotifier                AuthEmailNotifier
	logger                       Logger
	userIDGenerator              UserIDGenerator
	sessionIDGenerator           SessionIDGenerator
	verificationTokenIDGenerator EmailVerificationTokenIDGenerator
	resetTokenIDGenerator        PasswordResetTokenIDGenerator
	clock                        Clock
	refreshTokenTTL              time.Duration
	verificationTokenTTL         time.Duration
	resetTokenTTL                time.Duration
}

func NewService(
	userRepo UserRepository,
	registrationRepo RegistrationRepository,
	emailVerificationRepo EmailVerificationRepository,
	sessionRepo SessionRepository,
	passwordResetRepo PasswordResetRepository,
	accessTokenManager AccessTokenManager,
	refreshTokenGenerator TokenGenerator,
	verificationTokenGenerator TokenGenerator,
	resetTokenGenerator TokenGenerator,
	emailNotifier AuthEmailNotifier,
	logger Logger,
	userIDGenerator UserIDGenerator,
	sessionIDGenerator SessionIDGenerator,
	verificationTokenIDGenerator EmailVerificationTokenIDGenerator,
	resetTokenIDGenerator PasswordResetTokenIDGenerator,
	refreshTokenTTL time.Duration,
	verificationTokenTTL time.Duration,
	resetTokenTTL time.Duration,
	clock Clock,
) (*Service, error) {
	switch {
	case userRepo == nil:
		return nil, ErrNilUserRepository
	case registrationRepo == nil:
		return nil, ErrNilRegistrationRepo
	case emailVerificationRepo == nil:
		return nil, ErrNilEmailVerificationRepo
	case sessionRepo == nil:
		return nil, ErrNilSessionRepository
	case passwordResetRepo == nil:
		return nil, ErrNilPasswordResetRepo
	case accessTokenManager == nil:
		return nil, ErrNilAccessTokenManager
	case refreshTokenGenerator == nil:
		return nil, ErrNilRefreshTokenGenerator
	case verificationTokenGenerator == nil:
		return nil, ErrNilVerificationTokenGenerator
	case resetTokenGenerator == nil:
		return nil, ErrNilResetTokenGenerator
	case emailNotifier == nil:
		return nil, ErrNilEmailNotifier
	case logger == nil:
		return nil, ErrNilLogger
	case userIDGenerator == nil:
		return nil, ErrNilUserIDGenerator
	case sessionIDGenerator == nil:
		return nil, ErrNilSessionIDGenerator
	case verificationTokenIDGenerator == nil:
		return nil, ErrNilVerificationTokenIDGenerator
	case resetTokenIDGenerator == nil:
		return nil, ErrNilResetTokenIDGenerator
	case clock == nil:
		return nil, ErrNilClock
	}

	return &Service{
		userRepo:                     userRepo,
		registrationRepo:             registrationRepo,
		emailVerificationRepo:        emailVerificationRepo,
		sessionRepo:                  sessionRepo,
		passwordResetRepo:            passwordResetRepo,
		accessTokenManager:           accessTokenManager,
		refreshTokenGenerator:        refreshTokenGenerator,
		verificationTokenGenerator:   verificationTokenGenerator,
		resetTokenGenerator:          resetTokenGenerator,
		emailNotifier:                emailNotifier,
		logger:                       logger,
		userIDGenerator:              userIDGenerator,
		sessionIDGenerator:           sessionIDGenerator,
		verificationTokenIDGenerator: verificationTokenIDGenerator,
		resetTokenIDGenerator:        resetTokenIDGenerator,
		clock:                        clock,
		refreshTokenTTL:              refreshTokenTTL,
		verificationTokenTTL:         verificationTokenTTL,
		resetTokenTTL:                resetTokenTTL,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	email, err := userdomain.NewEmail(input.Email)
	if err != nil {
		return RegisterOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_email",
			"email has invalid format",
			err,
		)
	}

	if _, err := userdomain.NewPassword(input.Password); err != nil {
		return RegisterOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_password",
			"password does not satisfy constraints",
			err,
		)
	}

	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return RegisterOutput{}, err
	}
	if existingUser != nil {
		return RegisterOutput{}, apperrors.New(
			apperrors.KindConflict,
			"email_already_exists",
			"email is already registered",
		)
	}

	userID, err := s.userIDGenerator.NewUserID()
	if err != nil {
		return RegisterOutput{}, err
	}

	user, err := userdomain.NewUser(userID.String(), email.String(), input.Password, string(userdomain.UserRoleUser))
	if err != nil {
		return RegisterOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_user",
			"failed to create user",
			err,
		)
	}

	if strings.TrimSpace(input.DisplayName) != "" {
		if err := user.UpdateDisplayName(input.DisplayName, s.clock.Now()); err != nil {
			return RegisterOutput{}, apperrors.Wrap(
				apperrors.KindValidation,
				"invalid_display_name",
				"invalid display name",
				err,
			)
		}
	}

	verificationTokenID, err := s.verificationTokenIDGenerator.NewEmailVerificationTokenID()
	if err != nil {
		return RegisterOutput{}, err
	}

	rawVerificationToken, verificationTokenHash, err := s.verificationTokenGenerator.NewToken()
	if err != nil {
		return RegisterOutput{}, err
	}

	verificationToken, err := authdomain.NewEmailVerificationToken(
		verificationTokenID,
		user.ID(),
		verificationTokenHash,
		s.clock.Now(),
		s.clock.Now().Add(s.verificationTokenTTL),
	)
	if err != nil {
		return RegisterOutput{}, err
	}

	if err := s.registrationRepo.SaveUserWithVerificationToken(ctx, &user, &verificationToken); err != nil {
		if errors.Is(err, userdomain.ErrUserExists) {
			return RegisterOutput{}, apperrors.New(
				apperrors.KindConflict,
				"email_already_exists",
				"email is already registered",
			)
		}

		return RegisterOutput{}, err
	}

	s.logEmailDeliveryError(
		"failed to send registration verification email",
		&user,
		s.emailNotifier.SendVerificationEmail(ctx, &user, rawVerificationToken, input.FrontendBaseURL),
	)

	return RegisterOutput{
		User: userusecase.Profile{
			ID:            user.ID().String(),
			Email:         user.Email().String(),
			EmailVerified: user.IsEmailVerified(),
			Role:          string(user.Role()),
			DisplayName:   user.DisplayName(),
			CreatedAt:     user.CreatedAt(),
			UpdatedAt:     user.UpdatedAt(),
			LastLoginAt:   user.LastLoginAt(),
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginOutput, error) {
	email, err := userdomain.NewEmail(input.Email)
	if err != nil {
		return LoginOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_email",
			"email has invalid format",
			err,
		)
	}
	if strings.TrimSpace(input.Password) == "" {
		return LoginOutput{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_password",
			"password is required",
		)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return LoginOutput{}, err
	}
	if user == nil || !user.ComparePasswordString(input.Password) {
		return LoginOutput{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_credentials",
			"invalid email or password",
		)
	}

	if !user.IsEmailVerified() {
		return LoginOutput{}, apperrors.New(
			apperrors.KindForbidden,
			"email_not_verified",
			"confirm your email before signing in",
		)
	}

	output, err := s.createSession(ctx, user)
	if err != nil {
		return LoginOutput{}, err
	}

	now := s.clock.Now()
	user.MarkLoggedIn(now)
	if err := s.userRepo.MarkLastLogin(ctx, user.ID(), now); err != nil {
		return LoginOutput{}, err
	}

	output.User = userusecase.Profile{
		ID:            user.ID().String(),
		Email:         user.Email().String(),
		EmailVerified: user.IsEmailVerified(),
		Role:          string(user.Role()),
		DisplayName:   user.DisplayName(),
		CreatedAt:     user.CreatedAt(),
		UpdatedAt:     user.UpdatedAt(),
		LastLoginAt:   user.LastLoginAt(),
	}

	return output, nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (RefreshOutput, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return RefreshOutput{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_refresh_token",
			"refresh token is required",
		)
	}

	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, s.refreshTokenGenerator.Hash(refreshToken))
	if err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			return RefreshOutput{}, apperrors.New(
				apperrors.KindUnauthorized,
				"invalid_refresh_token",
				"refresh token is invalid",
			)
		}

		return RefreshOutput{}, err
	}

	now := s.clock.Now()
	if session == nil || session.IsRevoked() || session.IsExpired(now) {
		return RefreshOutput{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_refresh_token",
			"refresh token is invalid",
		)
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID())
	if err != nil {
		return RefreshOutput{}, err
	}
	if user == nil {
		return RefreshOutput{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_refresh_token",
			"refresh token is invalid",
		)
	}

	if !user.IsEmailVerified() {
		session.Revoke(now)
		if err := s.sessionRepo.Update(ctx, session); err != nil {
			return RefreshOutput{}, err
		}
		return RefreshOutput{}, apperrors.New(
			apperrors.KindForbidden,
			"email_not_verified",
			"confirm your email before continuing",
		)
	}

	rawRefreshToken, refreshTokenHash, err := s.refreshTokenGenerator.NewToken()
	if err != nil {
		return RefreshOutput{}, err
	}

	if err := session.Rotate(refreshTokenHash, now, now.Add(s.refreshTokenTTL)); err != nil {
		return RefreshOutput{}, err
	}

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return RefreshOutput{}, err
	}

	accessToken, err := s.accessTokenManager.IssueToken(Actor{
		UserID:    user.ID().String(),
		SessionID: session.ID().String(),
		Role:      string(user.Role()),
	}, now)
	if err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{
		Auth: AuthPayload{
			AccessToken:  accessToken,
			RefreshToken: rawRefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.accessTokenManager.TokenTTL().Seconds()),
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) (AcceptedOutput, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return AcceptedOutput{Accepted: true}, nil
	}

	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, s.refreshTokenGenerator.Hash(refreshToken))
	if err != nil {
		if errors.Is(err, authdomain.ErrSessionNotFound) {
			return AcceptedOutput{Accepted: true}, nil
		}

		return AcceptedOutput{}, err
	}
	if session == nil || session.IsRevoked() {
		return AcceptedOutput{Accepted: true}, nil
	}

	session.Revoke(s.clock.Now())

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return AcceptedOutput{}, err
	}

	return AcceptedOutput{Accepted: true}, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, input RequestPasswordResetInput) (AcceptedOutput, error) {
	email, err := userdomain.NewEmail(input.Email)
	if err != nil {
		return AcceptedOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_email",
			"email has invalid format",
			err,
		)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return AcceptedOutput{}, err
	}
	if user == nil {
		return AcceptedOutput{Accepted: true}, nil
	}

	resetTokenID, err := s.resetTokenIDGenerator.NewPasswordResetTokenID()
	if err != nil {
		return AcceptedOutput{}, err
	}

	rawResetToken, resetTokenHash, err := s.resetTokenGenerator.NewToken()
	if err != nil {
		return AcceptedOutput{}, err
	}

	resetToken, err := authdomain.NewPasswordResetToken(
		resetTokenID,
		user.ID(),
		resetTokenHash,
		s.clock.Now(),
		s.clock.Now().Add(s.resetTokenTTL),
	)
	if err != nil {
		return AcceptedOutput{}, err
	}

	if err := s.passwordResetRepo.Save(ctx, &resetToken); err != nil {
		return AcceptedOutput{}, err
	}

	s.logEmailDeliveryError(
		"failed to send password reset email",
		user,
		s.emailNotifier.SendPasswordResetEmail(ctx, user, rawResetToken, input.FrontendBaseURL),
	)

	return AcceptedOutput{Accepted: true}, nil
}

func (s *Service) ConfirmEmailVerification(ctx context.Context, input ConfirmEmailVerificationInput) (AcceptedOutput, error) {
	rawToken := strings.TrimSpace(input.Token)
	if rawToken == "" {
		return AcceptedOutput{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_email_verification_token",
			"verification token is required",
		)
	}

	if err := s.emailVerificationRepo.Consume(ctx, s.verificationTokenGenerator.Hash(rawToken), s.clock.Now()); err != nil {
		switch {
		case errors.Is(err, authdomain.ErrEmailVerificationTokenNotFound),
			errors.Is(err, authdomain.ErrEmailVerificationTokenExpired),
			errors.Is(err, authdomain.ErrEmailVerificationTokenUsed):
			return AcceptedOutput{}, apperrors.New(
				apperrors.KindValidation,
				"invalid_email_verification_token",
				"verification token is invalid or expired",
			)
		default:
			return AcceptedOutput{}, err
		}
	}

	return AcceptedOutput{Accepted: true}, nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, input ConfirmPasswordResetInput) (AcceptedOutput, error) {
	rawToken := strings.TrimSpace(input.Token)
	if rawToken == "" {
		return AcceptedOutput{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_reset_token",
			"reset token is required",
		)
	}

	password, err := userdomain.NewPassword(input.NewPassword)
	if err != nil {
		return AcceptedOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_password",
			"password does not satisfy constraints",
			err,
		)
	}

	passwordHash, err := userdomain.NewPasswordHash(password)
	if err != nil {
		return AcceptedOutput{}, err
	}

	if err := s.passwordResetRepo.Consume(ctx, s.resetTokenGenerator.Hash(rawToken), passwordHash.String(), s.clock.Now()); err != nil {
		switch {
		case errors.Is(err, authdomain.ErrPasswordResetTokenNotFound),
			errors.Is(err, authdomain.ErrPasswordResetTokenExpired),
			errors.Is(err, authdomain.ErrPasswordResetTokenUsed):
			return AcceptedOutput{}, apperrors.New(
				apperrors.KindValidation,
				"invalid_reset_token",
				"reset token is invalid or expired",
			)
		default:
			return AcceptedOutput{}, err
		}
	}

	return AcceptedOutput{Accepted: true}, nil
}

func (s *Service) Authorize(ctx context.Context, rawAccessToken string) (Actor, error) {
	token := strings.TrimSpace(rawAccessToken)
	if token == "" {
		return Actor{}, apperrors.New(
			apperrors.KindUnauthorized,
			"missing_access_token",
			"access token is required",
		)
	}

	actor, err := s.accessTokenManager.ParseToken(token)
	if err != nil {
		return Actor{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_access_token",
			"access token is invalid",
		)
	}

	userID, err := userdomain.NewUserID(actor.UserID)
	if err != nil {
		return Actor{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_access_token",
			"access token is invalid",
		)
	}
	if _, err := authdomain.NewSessionID(actor.SessionID); err != nil {
		return Actor{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_access_token",
			"access token is invalid",
		)
	}

	switch actor.Role {
	case string(userdomain.UserRoleAdmin), string(userdomain.UserRoleUser):
	default:
		return Actor{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_access_token",
			"access token is invalid",
		)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return Actor{}, err
	}
	if user == nil {
		return Actor{}, apperrors.New(
			apperrors.KindUnauthorized,
			"invalid_access_token",
			"access token is invalid",
		)
	}
	if !user.IsEmailVerified() {
		return Actor{}, apperrors.New(
			apperrors.KindForbidden,
			"email_not_verified",
			"confirm your email before continuing",
		)
	}

	return actor, nil
}

func (s *Service) createSession(ctx context.Context, user *userdomain.User) (LoginOutput, error) {
	sessionID, err := s.sessionIDGenerator.NewSessionID()
	if err != nil {
		return LoginOutput{}, err
	}

	refreshToken, refreshTokenHash, err := s.refreshTokenGenerator.NewToken()
	if err != nil {
		return LoginOutput{}, err
	}

	now := s.clock.Now()
	session, err := authdomain.NewSession(
		sessionID,
		user.ID(),
		refreshTokenHash,
		now,
		now.Add(s.refreshTokenTTL),
	)
	if err != nil {
		return LoginOutput{}, err
	}

	if err := s.sessionRepo.Save(ctx, &session); err != nil {
		return LoginOutput{}, err
	}

	accessToken, err := s.accessTokenManager.IssueToken(Actor{
		UserID:    user.ID().String(),
		SessionID: session.ID().String(),
		Role:      string(user.Role()),
	}, now)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		Auth: AuthPayload{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(s.accessTokenManager.TokenTTL().Seconds()),
		},
	}, nil
}

func (s *Service) logEmailDeliveryError(message string, user *userdomain.User, err error) {
	if err == nil || user == nil {
		return
	}

	s.logger.Error(
		message,
		"user_id",
		user.ID().String(),
		"email",
		user.Email().String(),
		"error",
		err,
	)
}
