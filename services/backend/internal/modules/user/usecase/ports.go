package usecase

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	Save(ctx context.Context, user *domain.User) error
}

type UserIDGenerator interface {
	NewUserID() (domain.UserID, error)
}
