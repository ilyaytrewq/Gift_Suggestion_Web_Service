package authsession

import (
	"testing"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewAuthSessionValid(t *testing.T) {
	t.Parallel()

	ttl := time.Hour
	uId := new(user.UserID(validUUID))
	session, err := NewAuthSession(uId, ttl)
	if err != nil {
		t.Fatalf("NewAuthSession() error=%v", err)
	}
	if session == nil {
		t.Fatal("NewAuthSession() returned nil session")
	}
	if session.AuthSessionID == nil {
		t.Fatal("AuthSessionID is empty")
	}
	if session.CSRFSecret == nil {
		t.Fatal("CSRFSecret is empty")
	}
	if err := session.AuthSessionID.IsValid(); err != nil {
		t.Fatalf("AuthSessionID.IsValid() error=%v", err)
	}
	if err := session.CSRFSecret.IsValid(); err != nil {
		t.Fatalf("CSRFSecret.IsValid() error=%v", err)
	}
	if got := session.ExpiresAt.Sub(session.CreatedAt); got != ttl {
		t.Fatalf("ExpiresAt- CreatedAt=%v, want %v", got, ttl)
	}
	if !session.LastUsedAt.Equal(session.CreatedAt) {
		t.Fatalf("LastUsedAt=%v, want %v", session.LastUsedAt, session.CreatedAt)
	}
}

func TestNewAuthSessionInvalidUserID(t *testing.T) {
	t.Parallel()

	_, err := NewAuthSession(new(user.UserID("bad-id")), time.Hour)
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestNewAuthSessionInvalidTTL(t *testing.T) {
	t.Parallel()

	_, err := NewAuthSession(new(user.UserID(validUUID)), 0)
	if err != ErrInvalidTTL {
		t.Fatalf("expected ErrInvalidTTL, got %v", err)
	}

	_, err = NewAuthSession(new(user.UserID(validUUID)), -time.Second)
	if err != ErrInvalidTTL {
		t.Fatalf("expected ErrInvalidTTL, got %v", err)
	}
}
