package idgen

import (
	"github.com/google/uuid"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type UUIDGenerator struct{}

func (UUIDGenerator) NewUserID() (userdomain.UserID, error) {
	return userdomain.NewUserID(uuid.NewString())
}

func (UUIDGenerator) NewSessionID() (authdomain.SessionID, error) {
	return authdomain.NewSessionID(uuid.NewString())
}

func (UUIDGenerator) NewPasswordResetTokenID() (authdomain.PasswordResetTokenID, error) {
	return authdomain.NewPasswordResetTokenID(uuid.NewString())
}
