package domain

import (
	"errors"
	"testing"
)

func TestNewRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Role
		wantErr error
	}{
		{
			name:  "admin",
			input: string(UserRoleAdmin),
			want:  UserRoleAdmin,
		},
		{
			name:  "user",
			input: string(UserRoleUser),
			want:  UserRoleUser,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: ErrRoleEmpty,
		},
		{
			name:    "invalid",
			input:   "boss",
			wantErr: ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			role, err := newRole(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("newRole() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && role != tt.want {
				t.Fatalf("newRole() role = %q, want %q", role, tt.want)
			}
		})
	}
}
