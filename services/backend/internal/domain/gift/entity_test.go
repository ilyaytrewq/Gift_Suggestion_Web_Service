package gift

import (
	"reflect"
	"testing"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/shared"
)

const (
	validUUID = "550e8400-e29b-41d4-a716-446655440000"
)

func TestNewGiftValid(t *testing.T) {
	t.Parallel()

	urls := []string{
		"HTTPS://Example.com:443/path#frag",
		"https://example.com/path",
		"http://example.com:80/path",
	}
	tags := []shared.TagID{"t2", "t1", "t1"}

	gift, err := NewGift(GiftID(validUUID), "Gift", shared.CategoryID("category"), shared.Money(100), urls, tags, shared.Age12)
	if err != nil {
		t.Fatalf("NewGift() error=%v", err)
	}
	if gift == nil {
		t.Fatal("NewGift() returned nil gift")
	}

	wantURLs := []string{"https://example.com/path", "http://example.com/path"}
	if !reflect.DeepEqual(gift.shopURLs, wantURLs) {
		t.Fatalf("shopURLs=%v, want %v", gift.shopURLs, wantURLs)
	}

	wantTags := []shared.TagID{"t1", "t2"}
	if !reflect.DeepEqual(gift.tags, wantTags) {
		t.Fatalf("tags=%v, want %v", gift.tags, wantTags)
	}
}

func TestNewGiftInvalidID(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID("bad-id"), "Gift", shared.CategoryID("category"), shared.Money(100), []string{"https://example.com"}, nil, shared.Age12)
	if err != ErrInvalidGiftID {
		t.Fatalf("expected ErrInvalidGiftID, got %v", err)
	}
}

func TestNewGiftInvalidTitle(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "", shared.CategoryID("category"), shared.Money(100), []string{"https://example.com"}, nil, shared.Age12)
	if err != ErrInvalidTitle {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestNewGiftInvalidCategory(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared.CategoryID(""), shared.Money(100), []string{"https://example.com"}, nil, shared.Age12)
	if err != ErrInvalidCategory {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestNewGiftInvalidURL(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared.CategoryID("category"), shared.Money(100), []string{"ftp://example.com"}, nil, shared.Age12)
	if err != ErrInvalidShopURL {
		t.Fatalf("expected ErrInvalidShopURL, got %v", err)
	}
}

func TestNewGiftInvalidTag(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared.CategoryID("category"), shared.Money(100), []string{"https://example.com"}, []shared.TagID{""}, shared.Age12)
	if err != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestNewGiftInvalidPrice(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared.CategoryID("category"), shared.Money(-1), []string{"https://example.com"}, nil, shared.Age12)
	if err != shared.ErrInvalidPrice {
		t.Fatalf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestNewGiftInvalidAge(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared.CategoryID("category"), shared.Money(100), []string{"https://example.com"}, nil, shared.AgeLimit(15))
	if err != shared.ErrInvalidAgeLimit {
		t.Fatalf("expected ErrInvalidAgeLimit, got %v", err)
	}
}
