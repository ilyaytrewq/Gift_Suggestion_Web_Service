package domain

import "testing"

func TestNewEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "valid",
			input: validTestEmail,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: ErrEmailEmpty,
		},
		{
			name:    "invalid format",
			input:   "not-an-email",
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			email, err := newEmail(tt.input)
			if err != tt.wantErr {
				t.Fatalf("newEmail() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && email.value != validTestEmail {
				t.Fatalf("newEmail() value = %q, want %q", email.value, validTestEmail)
			}
		})
	}
}

func TestEmailIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		email   Email
		wantErr error
	}{
		{
			name:    "valid",
			email:   mustEmail(t, validTestEmail),
			wantErr: nil,
		},
		{
			name:    "empty",
			email:   Email{},
			wantErr: ErrEmailEmpty,
		},
		{
			name:    "invalid format",
			email:   Email{value: "not-an-email"},
			wantErr: ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.email.IsValid(); err != tt.wantErr {
				t.Fatalf("IsValid() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
