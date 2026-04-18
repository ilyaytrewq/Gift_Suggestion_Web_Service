package usecase

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase/register"
)

type RegistrationUseCase struct {
	repo  register.UserRepository
	idGen register.IDGenerator
}

func NewRegistrationUseCase(
	repo register.UserRepository,
	idGen register.IDGenerator,
) (RegistrationUseCase, error) {
	if repo == nil {
		return RegistrationUseCase{}, ErrNilUserRepository
	}
	if idGen == nil {
		return RegistrationUseCase{}, ErrNilIDGenerator
	}
	return RegistrationUseCase{
		repo:  repo,
		idGen: idGen,
	}, nil
}

func (uc *RegistrationUseCase) Register(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return RegisterOutput{}, err
	}

	acc, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to get user by email")
	}
	if acc != nil {
		return RegisterOutput{}, ErrEmailAlreadyExists
	}

	id, err := uc.idGen.NewUserID()
	if err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to generate id")
	}

	newUser, err := domain.NewUser(id.String(), input.Email, input.Password, string(domain.UserRoleUser))
	if err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to create new user")
	}

	if err = uc.repo.Save(ctx, &newUser); err != nil {
		return RegisterOutput{}, errors.Wrap(err, "failed to save new user")
	}

	return RegisterOutput{
		UserID: newUser.ID(),
		Email:  newUser.Email(),
		Role:   newUser.Role(),
	}, nil
}
