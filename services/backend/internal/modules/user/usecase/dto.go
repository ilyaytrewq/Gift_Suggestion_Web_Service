package usecase

import (
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type Profile struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	DisplayName string     `json:"display_name"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type UpdateProfileInput struct {
	UserID      string
	DisplayName string
}

func newProfile(user *domain.User) Profile {
	return Profile{
		ID:          user.ID().String(),
		Email:       user.Email().String(),
		Role:        string(user.Role()),
		DisplayName: user.DisplayName(),
		CreatedAt:   user.CreatedAt(),
		UpdatedAt:   user.UpdatedAt(),
		LastLoginAt: user.LastLoginAt(),
	}
}
