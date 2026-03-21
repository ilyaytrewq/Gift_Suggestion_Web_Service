package domain

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordHash(t *testing.T) {
	t.Parallel()

	password := mustPassword(t, validTestPassword)

	hash, err := newPasswordHash(password)
	if err != nil {
		t.Fatalf("newPasswordHash() error = %v", err)
	}

	if len(hash.value) == 0 {
		t.Fatal("newPasswordHash() returned empty hash")
	}

	if string(hash.value) == password.value {
		t.Fatal("newPasswordHash() returned raw password instead of hash")
	}

	if err = bcrypt.CompareHashAndPassword(hash.value, []byte(password.value)); err != nil {
		t.Fatalf("CompareHashAndPassword() error = %v", err)
	}
}
