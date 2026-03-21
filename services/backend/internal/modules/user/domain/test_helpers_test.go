package domain

import "testing"

const (
	validTestUserID   = "550e8400-e29b-41d4-a716-446655440000"
	validTestEmail    = "test@example.com"
	validTestPassword = "ValidPass1!"
)

func mustUserID(t *testing.T, raw string) UserID {
	t.Helper()

	id, err := newUserID(raw)
	if err != nil {
		t.Fatalf("newUserID(%q) error = %v", raw, err)
	}

	return id
}

func mustEmail(t *testing.T, raw string) Email {
	t.Helper()

	email, err := newEmail(raw)
	if err != nil {
		t.Fatalf("newEmail(%q) error = %v", raw, err)
	}

	return email
}

func mustPassword(t *testing.T, raw string) Password {
	t.Helper()

	password, err := newPassword(raw)
	if err != nil {
		t.Fatalf("newPassword(%q) error = %v", raw, err)
	}

	return password
}
