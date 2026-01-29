package registration

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

type RegistrationUseCase struct {
	repo  UserRepository
	idGen IDGenerator
}

func NewRegistrationUseCase(repo UserRepository, idGen IDGenerator) (*RegistrationUseCase, error) {
	if repo == nil {
		return nil, ErrNilUserRepository
	}
	if idGen == nil {
		return nil, ErrNilIDGenerator
	}
	return &RegistrationUseCase{
		repo:  repo,
		idGen: idGen,
	}, nil
}

func (uc *RegistrationUseCase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	email, err := user.NewEmail(input.Email)
	if err != nil {
		return nil, err
	}

	acc, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user by email")
	}
	if acc != nil {
		return nil, ErrEmailAlreadyExists
	}

	id, err := uc.idGen.NewUserID()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate id")
	}

	newUser, err := user.NewUser(id, email, user.Password(input.Password), user.UserRoleUser)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new user")
	}

	if err = uc.repo.Save(ctx, newUser); err != nil {
		return nil, errors.Wrap(err, "failed to save new user")
	}

	return &RegisterOutput{
		UserID: newUser.ID(),
		Email:  newUser.Email(),
		Role:   newUser.Role(),
	}, nil
}
