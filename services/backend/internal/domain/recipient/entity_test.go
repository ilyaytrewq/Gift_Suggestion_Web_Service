package recipient

import (
	"reflect"
	"testing"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/services/backend/internal/domain/shared"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/services/backend/internal/domain/user"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewRecipientProfileValid(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	occasion := shared.NewYear
	relation := shared.Mother
	profile, err := NewRecipientProfile(
		RecipientProfileID(validUUID),
		&ownerID,
		&occasion,
		&relation,
		shared.Age12,
		shared.Money(100),
		[]shared.TagID{"b", "a", "a"},
	)
	if err != nil {
		t.Fatalf("NewRecipientProfile() error=%v", err)
	}
	if profile == nil {
		t.Fatal("NewRecipientProfile() returned nil profile")
	}
	if got := profile.InterestTags; !reflect.DeepEqual(got, []shared.TagID{"a", "b"}) {
		t.Fatalf("InterestTags=%v, want %v", got, []shared.TagID{"a", "b"})
	}
}

func TestNewRecipientProfileInvalidID(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	_, err := NewRecipientProfile(RecipientProfileID(""), &ownerID, nil, nil, shared.Age12, shared.Money(0), nil)
	if err != ErrRecipientProfileIDEmpty {
		t.Fatalf("expected ErrRecipientProfileIDEmpty, got %v", err)
	}

	_, err = NewRecipientProfile(RecipientProfileID("bad-id"), &ownerID, nil, nil, shared.Age12, shared.Money(0), nil)
	if err != ErrInvalidRecipientProfileID {
		t.Fatalf("expected ErrInvalidRecipientProfileID, got %v", err)
	}
}

func TestNewRecipientProfileInvalidOwner(t *testing.T) {
	t.Parallel()

	invalidOwner := user.UserID("bad-id")
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), &invalidOwner, nil, nil, shared.Age12, shared.Money(0), nil)
	if err != ErrInvalidOwnerUserID {
		t.Fatalf("expected ErrInvalidOwnerUserID, got %v", err)
	}
}

func TestNewRecipientProfileNilOwnerAllowed(t *testing.T) {
	t.Parallel()

	occasion := shared.NewYear
	relation := shared.Mother
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), nil, &occasion, &relation, shared.Age12, shared.Money(0), nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestNewRecipientProfileInvalidOccasion(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	occasion := shared.Occasion(" ")
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), &ownerID, &occasion, nil, shared.Age12, shared.Money(0), nil)
	if err != ErrInvalidOccasion {
		t.Fatalf("expected ErrInvalidOccasion, got %v", err)
	}
}

func TestNewRecipientProfileInvalidRelation(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	relation := shared.Relation(" ")
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), &ownerID, nil, &relation, shared.Age12, shared.Money(0), nil)
	if err != ErrInvalidRelation {
		t.Fatalf("expected ErrInvalidRelation, got %v", err)
	}
}

func TestNewRecipientProfileInvalidAge(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), &ownerID, nil, nil, shared.AgeLimit(15), shared.Money(0), nil)
	if err != ErrInvalidAge {
		t.Fatalf("expected ErrInvalidAge, got %v", err)
	}
}

func TestNewRecipientProfileInvalidBudget(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), &ownerID, nil, nil, shared.Age12, shared.Money(-1), nil)
	if err != ErrInvalidBudget {
		t.Fatalf("expected ErrInvalidBudget, got %v", err)
	}
}

func TestNewRecipientProfileInvalidTag(t *testing.T) {
	t.Parallel()

	ownerID := user.UserID(validUUID)
	_, err := NewRecipientProfile(RecipientProfileID(validUUID), &ownerID, nil, nil, shared.Age12, shared.Money(0), []shared.TagID{" "})
	if err != ErrInvalidInterestTag {
		t.Fatalf("expected ErrInvalidInterestTag, got %v", err)
	}
}
