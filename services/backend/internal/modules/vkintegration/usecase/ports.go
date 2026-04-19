package usecase

import (
	"context"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	vkintegrationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/domain"
)

type ConnectionRepository interface {
	GetByUserID(ctx context.Context, userID userdomain.UserID) (*vkintegrationdomain.Connection, error)
	Save(ctx context.Context, connection *vkintegrationdomain.Connection) error
	ListImportedInterests(ctx context.Context, connectionID vkintegrationdomain.ConnectionID) ([]vkintegrationdomain.ImportedInterest, error)
	ReplaceImportedInterests(ctx context.Context, connection *vkintegrationdomain.Connection, interests []vkintegrationdomain.ImportedInterest) error
}

type UserReader interface {
	GetByID(ctx context.Context, id userdomain.UserID) (*userdomain.User, error)
}

type TokenProtector interface {
	Configured() bool
	Seal(plain string) (string, error)
	Open(ciphertext string) (string, error)
}

type InterestImporter interface {
	ImportInterests(ctx context.Context, input ImportInterestsRequest) (ImportInterestsResult, error)
}

type ImportInterestsRequest struct {
	ProviderUserID string
	AccessToken    string
	Scopes         []string
}

type ImportInterestsResult struct {
	Interests []ImportedInterestRecord
}

type ImportedInterestRecord struct {
	Name        string
	SourceLabel string
	Position    int
}

type ConnectionIDGenerator interface {
	NewVKConnectionID() (vkintegrationdomain.ConnectionID, error)
}

type Clock interface {
	Now() time.Time
}
