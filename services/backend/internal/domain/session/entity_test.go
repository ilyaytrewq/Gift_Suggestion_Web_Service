package session

import (
	"testing"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/shared"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/snapshot"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewSessionValid(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	snap, err := snapshot.NewSnapshot(shared.NewYear, shared.Mother, shared.Money(100), []shared.TagID{"a"}, shared.Age12)
	if err != nil {
		t.Fatalf("NewSnapshot() error=%v", err)
	}

	session, err := NewSession(SessionID(validUUID), &ownerID, *snap)
	if err != nil {
		t.Fatalf("NewSession() error=%v", err)
	}
	if session == nil {
		t.Fatal("NewSession() returned nil session")
	}
}

func TestNewSessionInvalidID(t *testing.T) {
	t.Parallel()

	snap, err := snapshot.NewSnapshot(shared.NewYear, shared.Mother, shared.Money(100), nil, shared.Age12)
	if err != nil {
		t.Fatalf("NewSnapshot() error=%v", err)
	}

	_, err = NewSession(SessionID(""), nil, *snap)
	if err != ErrSessionIDEmpty {
		t.Fatalf("expected ErrSessionIDEmpty, got %v", err)
	}

	_, err = NewSession(SessionID("bad-id"), nil, *snap)
	if err != ErrInvalidSessionID {
		t.Fatalf("expected ErrInvalidSessionID, got %v", err)
	}
}

func TestNewSessionInvalidOwnerUserID(t *testing.T) {
	t.Parallel()

	snap, err := snapshot.NewSnapshot(shared.NewYear, shared.Mother, shared.Money(100), nil, shared.Age12)
	if err != nil {
		t.Fatalf("NewSnapshot() error=%v", err)
	}

	invalidOwner := user.UserID("bad-id")
	_, err = NewSession(SessionID(validUUID), &invalidOwner, *snap)
	if err != ErrInvalidOwnerUserID {
		t.Fatalf("expected ErrInvalidOwnerUserID, got %v", err)
	}
}

func TestNewSessionNilOwnerAllowed(t *testing.T) {
	t.Parallel()

	snap, err := snapshot.NewSnapshot(shared.NewYear, shared.Mother, shared.Money(100), nil, shared.Age12)
	if err != nil {
		t.Fatalf("NewSnapshot() error=%v", err)
	}

	_, err = NewSession(SessionID(validUUID), nil, *snap)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNewSessionInvalidSnapshot(t *testing.T) {
	t.Parallel()

	snap := snapshot.Snapshot{
		Occasion: shared.Occasion(" "),
		Relation: shared.Mother,
		Budget:   shared.Money(100),
		Age:      shared.Age12,
	}
	_, err := NewSession(SessionID(validUUID), nil, snap)
	if err != ErrInvalidSnapshot {
		t.Fatalf("expected ErrInvalidSnapshot, got %v", err)
	}
}
