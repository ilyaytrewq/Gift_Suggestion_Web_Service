package usecase

import (
	userusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/usecase"
)

type Actor struct {
	UserID    string
	SessionID string
	Role      string
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type RequestPasswordResetInput struct {
	Email string
}

type AuthPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RegisterOutput struct {
	User userusecase.Profile `json:"user"`
}

type LoginOutput struct {
	User userusecase.Profile `json:"user"`
	Auth AuthPayload         `json:"auth"`
}

type RefreshOutput struct {
	Auth AuthPayload `json:"auth"`
}

type AcceptedOutput struct {
	Accepted bool `json:"accepted"`
}
