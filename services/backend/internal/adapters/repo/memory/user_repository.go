package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/usecases/registration"
)

var ErrNilUser = errors.New("user is nil")

type UserRepository struct {
	mu      sync.RWMutex
	byEmail map[user.Email]*user.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{byEmail: make(map[user.Email]*user.User)}
}

func (r *UserRepository) GetByEmail(_ context.Context, email user.Email) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	usr, ok := r.byEmail[email]
	if !ok {
		return nil, nil
	}

	return usr, nil
}

func (r *UserRepository) Save(_ context.Context, usr *user.User) error {
	if usr == nil {
		return ErrNilUser
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byEmail[usr.Email()] = usr
	return nil
}

var _ registration.UserRepository = (*UserRepository)(nil)
