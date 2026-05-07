package domain

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id UserID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
	UpdateUserRole(ctx context.Context, user *User) error
	MarkLastLogin(ctx context.Context, id UserID, at time.Time) error
}
