package domain

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id           UserID
	email        Email
	passwordHash PasswordHash
	role         Role
	displayName  string
	createdAt    time.Time
	updatedAt    time.Time
	lastLoginAt  *time.Time
}

func NewUser(id, email, password, role string) (User, error) {
	uid, err := newUserID(id)
	if err != nil {
		return User{}, err
	}

	em, err := newEmail(email)
	if err != nil {
		return User{}, err
	}

	psw, err := newPassword(password)
	if err != nil {
		return User{}, err
	}

	rl, err := newRole(role)
	if err != nil {
		return User{}, err
	}

	passwordHash, err := newPasswordHash(psw)
	if err != nil {
		return User{}, err
	}

	now := time.Now().UTC()

	return User{
		id:           uid,
		email:        em,
		passwordHash: passwordHash,
		role:         rl,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func (u *User) ComparePassword(password Password) bool {
	if u == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword(u.passwordHash.value, []byte(password.value)) == nil
}

func (u *User) ComparePasswordString(password string) bool {
	if u == nil || isBlank(password) {
		return false
	}

	return bcrypt.CompareHashAndPassword(u.passwordHash.value, []byte(password)) == nil
}

func RestoreUser(
	id string,
	email string,
	passwordHash string,
	role string,
	displayName string,
	createdAt time.Time,
	updatedAt time.Time,
	lastLoginAt *time.Time,
) (User, error) {
	uid, err := newUserID(id)
	if err != nil {
		return User{}, err
	}

	em, err := newEmail(email)
	if err != nil {
		return User{}, err
	}

	hash, err := RestorePasswordHash(passwordHash)
	if err != nil {
		return User{}, err
	}

	rl, err := newRole(role)
	if err != nil {
		return User{}, err
	}

	normalizedDisplayName, err := normalizeDisplayName(displayName)
	if err != nil {
		return User{}, err
	}

	return User{
		id:           uid,
		email:        em,
		passwordHash: hash,
		role:         rl,
		displayName:  normalizedDisplayName,
		createdAt:    createdAt.UTC(),
		updatedAt:    updatedAt.UTC(),
		lastLoginAt:  cloneTimePtr(lastLoginAt),
	}, nil
}

func (u *User) UpdateDisplayName(displayName string, updatedAt time.Time) error {
	if u == nil {
		return ErrUserNotFound
	}

	normalizedDisplayName, err := normalizeDisplayName(displayName)
	if err != nil {
		return err
	}

	u.displayName = normalizedDisplayName
	u.updatedAt = updatedAt.UTC()

	return nil
}

func (u *User) MarkLoggedIn(at time.Time) {
	if u == nil {
		return
	}

	loggedInAt := at.UTC()
	u.lastLoginAt = &loggedInAt
	u.updatedAt = loggedInAt
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

func (u *User) DisplayName() string {
	return u.displayName
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

func (u *User) LastLoginAt() *time.Time {
	return cloneTimePtr(u.lastLoginAt)
}

func normalizeDisplayName(displayName string) (string, error) {
	if isBlank(displayName) {
		return "", nil
	}

	normalized := strings.TrimSpace(displayName)
	if len([]rune(normalized)) > 120 {
		return "", ErrDisplayNameTooLong
	}

	return normalized, nil
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := value.UTC()
	return &cloned
}
