package snapshot

import (
	"reflect"
	"testing"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/models/shared"
)

func TestNewSnapshotValid(t *testing.T) {
	t.Parallel()

	snap, err := NewSnapshot(shared.NewYear, shared.Mother, shared.Money(100), []shared.TagID{"b", "a", "a"}, shared.Age12)
	if err != nil {
		t.Fatalf("NewSnapshot() error=%v", err)
	}
	if snap == nil {
		t.Fatal("NewSnapshot() returned nil snapshot")
	}
	if !reflect.DeepEqual(snap.InterestTags, []shared.TagID{"a", "b"}) {
		t.Fatalf("InterestTags=%v, want %v", snap.InterestTags, []shared.TagID{"a", "b"})
	}
}

func TestNewSnapshotInvalidOccasion(t *testing.T) {
	t.Parallel()

	_, err := NewSnapshot(shared.Occasion(" "), shared.Mother, shared.Money(0), nil, shared.Age12)
	if err != ErrInvalidOccasion {
		t.Fatalf("expected ErrInvalidOccasion, got %v", err)
	}
}

func TestNewSnapshotInvalidRelation(t *testing.T) {
	t.Parallel()

	_, err := NewSnapshot(shared.NewYear, shared.Relation(" "), shared.Money(0), nil, shared.Age12)
	if err != ErrInvalidRelation {
		t.Fatalf("expected ErrInvalidRelation, got %v", err)
	}
}

func TestNewSnapshotInvalidBudget(t *testing.T) {
	t.Parallel()

	_, err := NewSnapshot(shared.NewYear, shared.Mother, shared.Money(-1), nil, shared.Age12)
	if err != ErrInvalidBudget {
		t.Fatalf("expected ErrInvalidBudget, got %v", err)
	}
}

func TestNewSnapshotInvalidAge(t *testing.T) {
	t.Parallel()

	_, err := NewSnapshot(shared.NewYear, shared.Mother, shared.Money(0), nil, shared.AgeLimit(15))
	if err != ErrInvalidAge {
		t.Fatalf("expected ErrInvalidAge, got %v", err)
	}
}

func TestNewSnapshotInvalidTag(t *testing.T) {
	t.Parallel()

	_, err := NewSnapshot(shared.NewYear, shared.Mother, shared.Money(0), []shared.TagID{" "}, shared.Age12)
	if err != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}
