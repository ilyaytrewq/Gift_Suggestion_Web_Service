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
	userRepo              UserRepository
	sessionRepo           SessionRepository
	passwordResetRepo     PasswordResetRepository
	accessTokenManager    AccessTokenManager
	refreshTokenGenerator TokenGenerator
	resetTokenGenerator   TokenGenerator
	userIDGenerator       UserIDGenerator
	sessionIDGenerator    SessionIDGenerator
	resetTokenIDGenerator PasswordResetTokenIDGenerator
	clock                 Clock
	refreshTokenTTL       time.Duration
	resetTokenTTL         time.Duration
}

func NewService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	passwordResetRepo PasswordResetRepository,
	accessTokenManager AccessTokenManager,
	refreshTokenGenerator TokenGenerator,
	resetTokenGenerator TokenGenerator,
	userIDGenerator UserIDGenerator,
	sessionIDGenerator SessionIDGenerator,
	resetTokenIDGenerator PasswordResetTokenIDGenerator,
	refreshTokenTTL time.Duration,
	resetTokenTTL time.Duration,
	clock Clock,
) (*Service, error) {
	switch {
	case userRepo == nil:
		return nil, ErrNilUserRepository
	case sessionRepo == nil:
		return nil, ErrNilSessionRepository
	case passwordResetRepo == nil:
		return nil, ErrNilPasswordResetRepo
	case accessTokenManager == nil:
		return nil, ErrNilAccessTokenManager
	case refreshTokenGenerator == nil:
		return nil, ErrNilRefreshTokenGenerator
	case resetTokenGenerator == nil:
		return nil, ErrNilResetTokenGenerator
	case userIDGenerator == nil:
		return nil, ErrNilUserIDGenerator
	case sessionIDGenerator == nil:
		return nil, ErrNilSessionIDGenerator
	case resetTokenIDGenerator == nil:
		return nil, ErrNilResetTokenIDGenerator
	case clock == nil:
		return nil, ErrNilClock
	}

	return &Service{
		userRepo:              userRepo,
		sessionRepo:           sessionRepo,
		passwordResetRepo:     passwordResetRepo,
		accessTokenManager:    accessTokenManager,
		refreshTokenGenerator: refreshTokenGenerator,
		resetTokenGenerator:   resetTokenGenerator,
		userIDGenerator:       userIDGenerator,
		sessionIDGenerator:    sessionIDGenerator,
		resetTokenIDGenerator: resetTokenIDGenerator,
		clock:                 clock,
		refreshTokenTTL:       refreshTokenTTL,
		resetTokenTTL:         resetTokenTTL,
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

	if err := s.userRepo.Save(ctx, &user); err != nil {
		if errors.Is(err, userdomain.ErrUserExists) {
			return RegisterOutput{}, apperrors.New(
				apperrors.KindConflict,
				"email_already_exists",
				"email is already registered",
			)
		}

		return RegisterOutput{}, err
	}

	return RegisterOutput{
		User: userusecase.Profile{
			ID:          user.ID().String(),
			Email:       user.Email().String(),
			Role:        string(user.Role()),
			DisplayName: user.DisplayName(),
			CreatedAt:   user.CreatedAt(),
			UpdatedAt:   user.UpdatedAt(),
			LastLoginAt: user.LastLoginAt(),
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
		ID:          user.ID().String(),
		Email:       user.Email().String(),
		Role:        string(user.Role()),
		DisplayName: user.DisplayName(),
		CreatedAt:   user.CreatedAt(),
		UpdatedAt:   user.UpdatedAt(),
		LastLoginAt: user.LastLoginAt(),
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

	_, resetTokenHash, err := s.resetTokenGenerator.NewToken()
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

	return AcceptedOutput{Accepted: true}, nil
}

func (s *Service) Authorize(_ context.Context, rawAccessToken string) (Actor, error) {
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

	if _, err := userdomain.NewUserID(actor.UserID); err != nil {
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
