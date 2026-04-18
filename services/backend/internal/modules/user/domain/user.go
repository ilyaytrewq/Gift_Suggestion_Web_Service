package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id           UserID
	email        Email
	passwordHash PasswordHash
	role         Role
	createdAt    time.Time
	updatedAt    time.Time
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

	return User{
		id:           uid,
		email:        em,
		passwordHash: passwordHash,
		role:         rl,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
	}, nil
}

func (u *User) ComparePassword(password Password) bool {
	if u == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword(u.passwordHash.value, []byte(password.value)) == nil
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
