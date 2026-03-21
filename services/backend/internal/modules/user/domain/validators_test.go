package domain

import "testing"

func TestIsBlank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty",
			input: "",
			want:  true,
		},
		{
			name:  "spaces only",
			input: " \n\t ",
			want:  true,
		},
		{
			name:  "non blank",
			input: " value ",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isBlank(tt.input); got != tt.want {
				t.Fatalf("isBlank() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidUserID(t *testing.T) {
	t.Parallel()

	if !isValidUserID(validTestUserID) {
		t.Fatal("isValidUserID() = false, want true")
	}

	if isValidUserID("bad-id") {
		t.Fatal("isValidUserID() = true, want false")
	}
}

func TestIsValidEmail(t *testing.T) {
	t.Parallel()

	if err := isValidEmail(validTestEmail); err != nil {
		t.Fatalf("isValidEmail() error = %v", err)
	}

	if err := isValidEmail("bad-email"); err == nil {
		t.Fatal("isValidEmail() error = nil, want non-nil")
	}
}

func TestIsValidPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid",
			password: validTestPassword,
		},
		{
			name:     "contains space",
			password: "Valid Pass1!",
			wantErr:  ErrPasswordContainsSpace,
		},
		{
			name:     "missing special",
			password: "ValidPass1",
			wantErr:  ErrPasswordNoSpecial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := isValidPassword(tt.password); err != tt.wantErr {
				t.Fatalf("isValidPassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
