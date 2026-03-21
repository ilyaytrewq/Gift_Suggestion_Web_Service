package usecase

import (
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	UserID domain.UserID
	Email  domain.Email
	Role   domain.Role
}
