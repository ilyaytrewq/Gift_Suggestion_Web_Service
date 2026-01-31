package wishlist

import (
	"strings"
	"testing"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/gift"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestNewWishlistItemValid(t *testing.T) {
	t.Parallel()

	item, err := NewWishlistItem(user.UserID(validUUID), gift.GiftID(validUUID), "note")
	if err != nil {
		t.Fatalf("NewWishlistItem() error=%v", err)
	}
	if item == nil {
		t.Fatal("NewWishlistItem() returned nil item")
	}
}

func TestNewWishlistItemInvalidUserID(t *testing.T) {
	t.Parallel()

	_, err := NewWishlistItem(user.UserID("bad-id"), gift.GiftID(validUUID), "note")
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestNewWishlistItemInvalidGiftID(t *testing.T) {
	t.Parallel()

	_, err := NewWishlistItem(user.UserID(validUUID), gift.GiftID("bad-id"), "note")
	if err != ErrInvalidGiftID {
		t.Fatalf("expected ErrInvalidGiftID, got %v", err)
	}
}

func TestNewWishlistItemNoteTooLong(t *testing.T) {
	t.Parallel()

	note := strings.Repeat("a", maxNoteLen+1)
	_, err := NewWishlistItem(user.UserID(validUUID), gift.GiftID(validUUID), note)
	if err != ErrNoteTooLong {
		t.Fatalf("expected ErrNoteTooLong, got %v", err)
	}
}
