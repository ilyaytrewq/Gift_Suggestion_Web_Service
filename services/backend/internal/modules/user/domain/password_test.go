package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "valid",
			input: validTestPassword,
		},
		{
			name:    "too short",
			input:   "Aa1!",
			wantErr: ErrPasswordTooShort,
		},
		{
			name:    "too long",
			input:   strings.Repeat("Aa1!", 19),
			wantErr: ErrPasswordTooLong,
		},
		{
			name:    "no lowercase",
			input:   "PASSWORD1!",
			wantErr: ErrPasswordNoLower,
		},
		{
			name:    "no uppercase",
			input:   "password1!",
			wantErr: ErrPasswordNoUpper,
		},
		{
			name:    "no digit",
			input:   "Password!",
			wantErr: ErrPasswordNoDigit,
		},
		{
			name:    "no special",
			input:   "Password1",
			wantErr: ErrPasswordNoSpecial,
		},
		{
			name:    "contains space",
			input:   "Password 1!",
			wantErr: ErrPasswordContainsSpace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			password, err := newPassword(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("newPassword() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && password.value != validTestPassword {
				t.Fatalf("newPassword() value = %q, want %q", password.value, validTestPassword)
			}
		})
	}
}

func TestPasswordIsValid(t *testing.T) {
	t.Parallel()

	if !mustPassword(t, validTestPassword).IsValid() {
		t.Fatal("IsValid() = false, want true")
	}

	if (Password{value: "short"}).IsValid() {
		t.Fatal("IsValid() = true, want false")
	}
}
