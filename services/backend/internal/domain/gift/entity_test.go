package gift

import (
	"reflect"
	"testing"

	shared2 "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/models/shared"
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
	tags := []shared2.TagID{"t2", "t1", "t1"}

	gift, err := NewGift(GiftID(validUUID), "Gift", shared2.CategoryID("category"), shared2.Money(100), urls, tags, shared2.Age12)
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

	wantTags := []shared2.TagID{"t1", "t2"}
	if !reflect.DeepEqual(gift.tags, wantTags) {
		t.Fatalf("tags=%v, want %v", gift.tags, wantTags)
	}
}

func TestNewGiftInvalidID(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID("bad-id"), "Gift", shared2.CategoryID("category"), shared2.Money(100), []string{"https://example.com"}, nil, shared2.Age12)
	if err != ErrInvalidGiftID {
		t.Fatalf("expected ErrInvalidGiftID, got %v", err)
	}
}

func TestNewGiftInvalidTitle(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "", shared2.CategoryID("category"), shared2.Money(100), []string{"https://example.com"}, nil, shared2.Age12)
	if err != ErrInvalidTitle {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestNewGiftInvalidCategory(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared2.CategoryID(""), shared2.Money(100), []string{"https://example.com"}, nil, shared2.Age12)
	if err != ErrInvalidCategory {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestNewGiftInvalidURL(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared2.CategoryID("category"), shared2.Money(100), []string{"ftp://example.com"}, nil, shared2.Age12)
	if err != ErrInvalidShopURL {
		t.Fatalf("expected ErrInvalidShopURL, got %v", err)
	}
}

func TestNewGiftInvalidTag(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared2.CategoryID("category"), shared2.Money(100), []string{"https://example.com"}, []shared2.TagID{""}, shared2.Age12)
	if err != ErrInvalidTag {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}

func TestNewGiftInvalidPrice(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared2.CategoryID("category"), shared2.Money(-1), []string{"https://example.com"}, nil, shared2.Age12)
	if err != shared2.ErrInvalidPrice {
		t.Fatalf("expected ErrInvalidPrice, got %v", err)
	}
}

func TestNewGiftInvalidAge(t *testing.T) {
	t.Parallel()

	_, err := NewGift(GiftID(validUUID), "Gift", shared2.CategoryID("category"), shared2.Money(100), []string{"https://example.com"}, nil, shared2.AgeLimit(15))
	if err != shared2.ErrInvalidAgeLimit {
		t.Fatalf("expected ErrInvalidAgeLimit, got %v", err)
	}
}
