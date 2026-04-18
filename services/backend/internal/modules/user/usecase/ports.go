package usecase

import (
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type Repository = domain.Repository

type Clock interface {
	Now() time.Time
}
