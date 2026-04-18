package usecase

import (
	"context"
	"time"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *userdomain.User) error
	GetByID(ctx context.Context, id userdomain.UserID) (*userdomain.User, error)
	GetByEmail(ctx context.Context, email userdomain.Email) (*userdomain.User, error)
	MarkLastLogin(ctx context.Context, id userdomain.UserID, at time.Time) error
}

type SessionRepository interface {
	Save(ctx context.Context, session *authdomain.Session) error
	GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*authdomain.Session, error)
	Update(ctx context.Context, session *authdomain.Session) error
}

type PasswordResetRepository interface {
	Save(ctx context.Context, token *authdomain.PasswordResetToken) error
}

type AccessTokenManager interface {
	IssueToken(actor Actor, issuedAt time.Time) (string, error)
	ParseToken(token string) (Actor, error)
	TokenTTL() time.Duration
}

type TokenGenerator interface {
	NewToken() (rawToken, tokenHash string, err error)
	Hash(rawToken string) string
}

type UserIDGenerator interface {
	NewUserID() (userdomain.UserID, error)
}

type SessionIDGenerator interface {
	NewSessionID() (authdomain.SessionID, error)
}

type PasswordResetTokenIDGenerator interface {
	NewPasswordResetTokenID() (authdomain.PasswordResetTokenID, error)
}

type Clock interface {
	Now() time.Time
}
