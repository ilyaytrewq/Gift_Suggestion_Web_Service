package domain

import (
	"strings"
	"testing"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	user, err := NewUser(validTestUserID, validTestEmail, validTestPassword, string(UserRoleUser))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if got := user.ID().value.String(); got != validTestUserID {
		t.Fatalf("ID() = %q, want %q", got, validTestUserID)
	}

	if got := user.Email().value; got != validTestEmail {
		t.Fatalf("Email() = %q, want %q", got, validTestEmail)
	}

	if got := user.Role(); got != UserRoleUser {
		t.Fatalf("Role() = %q, want %q", got, UserRoleUser)
	}

	if len(user.PasswordHash().value) == 0 {
		t.Fatal("PasswordHash() returned empty value")
	}

	if user.CreatedAt().IsZero() {
		t.Fatal("CreatedAt() returned zero time")
	}

	if user.UpdatedAt().IsZero() {
		t.Fatal("UpdatedAt() returned zero time")
	}

	if !user.ComparePassword(mustPassword(t, validTestPassword)) {
		t.Fatal("ComparePassword() = false, want true")
	}

	if user.ComparePassword(mustPassword(t, "OtherPass1!")) {
		t.Fatal("ComparePassword() = true, want false")
	}
}

func TestNewUserInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		id       string
		email    string
		password string
		role     string
		wantErr  error
	}{
		{
			name:     "empty id",
			id:       "",
			email:    validTestEmail,
			password: validTestPassword,
			role:     string(UserRoleUser),
			wantErr:  ErrUserIDEmpty,
		},
		{
			name:     "invalid email",
			id:       validTestUserID,
			email:    "not-an-email",
			password: validTestPassword,
			role:     string(UserRoleUser),
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "invalid password",
			id:       validTestUserID,
			email:    validTestEmail,
			password: "short",
			role:     string(UserRoleUser),
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "invalid role",
			id:       validTestUserID,
			email:    validTestEmail,
			password: validTestPassword,
			role:     "boss",
			wantErr:  ErrInvalidRole,
		},
		{
			name:     "password too long",
			id:       validTestUserID,
			email:    validTestEmail,
			password: strings.Repeat("Aa1!", 19),
			role:     string(UserRoleUser),
			wantErr:  ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewUser(tt.id, tt.email, tt.password, tt.role)
			if err != tt.wantErr {
				t.Fatalf("NewUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserComparePasswordNilReceiver(t *testing.T) {
	t.Parallel()

	var user *User
	if user.ComparePassword(mustPassword(t, validTestPassword)) {
		t.Fatal("ComparePassword() = true, want false for nil receiver")
	}
}
