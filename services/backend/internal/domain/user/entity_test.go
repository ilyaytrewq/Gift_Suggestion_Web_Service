package user

import (
	"strings"
	"testing"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewUserValid(t *testing.T) {
	t.Parallel()

	u, err := NewUser(UserID(validUUID), Email("test@example.com"), Password("password123"), UserRoleUser)
	if err != nil {
		t.Fatalf("NewUser() error=%v", err)
	}
	if u == nil {
		t.Fatal("NewUser() returned nil user")
	}
	if !u.ComparePassword(Password("password123")) {
		t.Fatal("ComparePassword() expected true")
	}
	if u.ComparePassword(Password("wrongpass")) {
		t.Fatal("ComparePassword() expected false")
	}
}

func TestNewUserInvalidID(t *testing.T) {
	t.Parallel()

	_, err := NewUser(UserID(""), Email("test@example.com"), Password("password123"), UserRoleUser)
	if err != ErrUserIDEmpty {
		t.Fatalf("expected ErrUserIDEmpty, got %v", err)
	}

	_, err = NewUser(UserID("not-uuid"), Email("test@example.com"), Password("password123"), UserRoleUser)
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestNewUserInvalidEmail(t *testing.T) {
	t.Parallel()

	_, err := NewUser(UserID(validUUID), Email(""), Password("password123"), UserRoleUser)
	if err != ErrEmailEmpty {
		t.Fatalf("expected ErrEmailEmpty, got %v", err)
	}

	_, err = NewUser(UserID(validUUID), Email("not-an-email"), Password("password123"), UserRoleUser)
	if err != ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestNewUserInvalidPassword(t *testing.T) {
	t.Parallel()

	_, err := NewUser(UserID(validUUID), Email("test@example.com"), Password("short"), UserRoleUser)
	if err != ErrPasswordTooShort {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}

	longPassword := strings.Repeat("a", maxPasswordLen+1)
	_, err = NewUser(UserID(validUUID), Email("test@example.com"), Password(longPassword), UserRoleUser)
	if err != ErrPasswordTooLong {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
}

func TestNewUserInvalidRole(t *testing.T) {
	t.Parallel()

	_, err := NewUser(UserID(validUUID), Email("test@example.com"), Password("password123"), Role(""))
	if err != ErrRoleEmpty {
		t.Fatalf("expected ErrRoleEmpty, got %v", err)
	}

	_, err = NewUser(UserID(validUUID), Email("test@example.com"), Password("password123"), Role("boss"))
	if err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestComparePasswordNilReceiver(t *testing.T) {
	t.Parallel()

	var u *User
	if u.ComparePassword(Password("password123")) {
		t.Fatal("ComparePassword() expected false on nil receiver")
	}
}
