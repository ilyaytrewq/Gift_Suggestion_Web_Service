package user

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type (
	UserID       string
	Email        string
	Password     string
	PasswordHash []byte
	Role         string
)

const (
	UserRoleAdmin Role = "admin"
	UserRoleUser  Role = "user"
)

type User struct {
	id           UserID
	email        Email
	passwordHash PasswordHash
	role         Role
	createdAt    time.Time
	updatedAt    time.Time
}

func NewUser(id UserID, email Email, password Password, role Role) (*User, error) {
	uid, err := NewUserID(string(id))
	if err != nil {
		return nil, err
	}
	em, err := NewEmail(string(email))
	if err != nil {
		return nil, err
	}
	psw, err := NewPassword(string(password))
	if err != nil {
		return nil, err
	}
	rl, err := NewRole(string(role))
	if err != nil {
		return nil, err
	}

	passwordHash, err := newPasswordHash(psw)
	if err != nil {
		return nil, err
	}

	return &User{
		id:           uid,
		email:        em,
		passwordHash: passwordHash,
		role:         rl,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
	}, nil
}

func NewUserID(id string) (UserID, error) {
	if isBlank(id) {
		return "", ErrUserIDEmpty
	}
	if !isValidUserID(id) {
		return "", ErrInvalidUserID
	}
	return UserID(id), nil
}

func (id UserID) IsValid() bool {
	if id == "" {
		return false
	}
	if !isValidUserID(string(id)) {
		return false
	}
	return true
}

func NewEmail(email string) (Email, error) {
	if isBlank(email) {
		return "", ErrEmailEmpty
	}
	if !isValidEmail(email) {
		return "", ErrInvalidEmail
	}

	return Email(email), nil
}

func NewPassword(password string) (Password, error) {
	if isBlank(password) {
		return "", ErrPasswordEmpty
	}
	if len(password) < minPasswordLen {
		return "", ErrPasswordTooShort
	}
	if len(password) > maxPasswordLen {
		return "", ErrPasswordTooLong
	}

	return Password(password), nil
}

func newPasswordHash(password Password) (PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(string(password)), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate password hash")
	}

	return PasswordHash(hash), nil
}

func (u *User) ComparePassword(password Password) bool {
	if u == nil {
		return false
	}
	return bcrypt.CompareHashAndPassword(u.passwordHash, []byte(string(password))) == nil
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Role() Role {
	return u.role
}

func (u *User) PasswordHash() PasswordHash {
	return u.passwordHash
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func NewRole(role string) (Role, error) {
	if isBlank(role) {
		return "", ErrRoleEmpty
	}
	switch role {
	case string(UserRoleAdmin):
		return UserRoleAdmin, nil
	case string(UserRoleUser):
		return UserRoleUser, nil
	default:
		return "", ErrInvalidRole
	}
}
