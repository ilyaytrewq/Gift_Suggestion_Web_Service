package id_generators

import (
	"github.com/google/uuid"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/usecases/registration"
)

type UserIDGenerator struct{}

func (UserIDGenerator) NewUserID() (user.UserID, error) {
	return user.NewUserID(uuid.NewString())
}

var _ registration.UserIDGenerator = UserIDGenerator{}
