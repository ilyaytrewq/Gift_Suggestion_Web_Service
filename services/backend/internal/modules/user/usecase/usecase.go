package usecase

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type Service struct {
	repo  Repository
	clock Clock
}

func NewService(repo Repository, clock Clock) (*Service, error) {
	if repo == nil {
		return nil, ErrNilUserRepository
	}
	if clock == nil {
		return nil, ErrNilClock
	}
	return &Service{
		repo:  repo,
		clock: clock,
	}, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID string) (Profile, error) {
	id, err := domain.NewUserID(userID)
	if err != nil {
		return Profile{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_user_id",
			"invalid user id",
			err,
		)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Profile{}, err
	}
	if user == nil {
		return Profile{}, apperrors.New(
			apperrors.KindNotFound,
			"user_not_found",
			"user not found",
		)
	}

	return newProfile(user), nil
}

func (s *Service) UpdateProfile(ctx context.Context, input UpdateProfileInput) (Profile, error) {
	id, err := domain.NewUserID(input.UserID)
	if err != nil {
		return Profile{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_user_id",
			"invalid user id",
			err,
		)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Profile{}, err
	}
	if user == nil {
		return Profile{}, apperrors.New(
			apperrors.KindNotFound,
			"user_not_found",
			"user not found",
		)
	}

	if err := user.UpdateDisplayName(input.DisplayName, s.clock.Now()); err != nil {
		return Profile{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_display_name",
			"invalid display name",
			err,
		)
	}

	if err := s.repo.UpdateProfile(ctx, user); err != nil {
		return Profile{}, err
	}

	return newProfile(user), nil
}

func (s *Service) PromoteUserToAdmin(ctx context.Context, rawEmail string) (Profile, error) {
	em, err := domain.NewEmail(rawEmail)
	if err != nil {
		return Profile{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_email",
			"email has invalid format",
			err,
		)
	}

	user, err := s.repo.GetByEmail(ctx, em)
	if err != nil {
		return Profile{}, err
	}
	if user == nil {
		return Profile{}, apperrors.New(
			apperrors.KindNotFound,
			"user_not_found",
			"user not found",
		)
	}

	if user.Role() == domain.UserRoleAdmin {
		return newProfile(user), nil
	}

	if err := user.PromoteToAdmin(s.clock.Now()); err != nil {
		return Profile{}, err
	}

	if err := s.repo.UpdateUserRole(ctx, user); err != nil {
		return Profile{}, err
	}

	return newProfile(user), nil
}
