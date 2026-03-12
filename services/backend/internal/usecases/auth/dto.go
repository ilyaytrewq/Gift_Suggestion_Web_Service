package auth

import (
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/authsession"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	AuthSessionID *authsession.AuthSessionID
	CSRFSecret    *authsession.CSRFSecret
	UserID        *user.UserID
}

type IsAuthorizedInput struct {
	UserID        string
	AuthSessionID string
	CSRFSecret    string
}

type IsAuthorizedOutput struct{}

type LogoutInput struct {
	AuthSessionID string
	CSRFSecret    string
}
