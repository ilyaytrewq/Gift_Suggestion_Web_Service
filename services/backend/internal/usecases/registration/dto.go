package registration

import (
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	UserID user.UserID
	Email  user.Email
	Role   user.Role
}
