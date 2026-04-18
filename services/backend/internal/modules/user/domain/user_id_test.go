package domain

import (
	"errors"
	"testing"
)

func TestNewUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "valid",
			input: validTestUserID,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: ErrUserIDEmpty,
		},
		{
			name:    "invalid format",
			input:   "not-a-uuid",
			wantErr: ErrInvalidUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id, err := newUserID(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("newUserID() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && id.value.String() != validTestUserID {
				t.Fatalf("newUserID() value = %s, want %s", id.value.String(), validTestUserID)
			}
		})
	}
}

func TestUserIDIsValid(t *testing.T) {
	t.Parallel()

	id := mustUserID(t, validTestUserID)
	if err := id.IsValid(); err != nil {
		t.Fatalf("IsValid() error = %v", err)
	}
}
