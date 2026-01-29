package registration

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email user.Email) (*user.User, error)
	Save(ctx context.Context, user *user.User) error
}

type IDGenerator interface {
	NewUserID() (user.UserID, error)
}
