package usecase

import (
	"context"
	"testing"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testWishlistUserID      = "550e8400-e29b-41d4-a716-446655440100"
	testWishlistOtherUserID = "550e8400-e29b-41d4-a716-446655440101"
	testWishlistID          = "550e8400-e29b-41d4-a716-446655440102"
	testWishlistItemID      = "550e8400-e29b-41d4-a716-446655440103"
	testWishlistGiftID      = "550e8400-e29b-41d4-a716-446655440104"
	testWishlistStoreLink   = "https://example.com/gifts/lego"
	testWishlistPrice       = "129.99"
	testWishlistDescription = "Creative building set"
)

func TestServiceGetWishlistCreatesPersonalWishlistOnFirstRead(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	repo := &fakeWishlistRepo{}
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: repo,
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: now},
	})

	output, err := service.GetWishlist(context.Background(), GetWishlistInput{
		UserID: testWishlistUserID,
	})
	if err != nil {
		t.Fatalf("GetWishlist() error = %v", err)
	}

	if output.Wishlist.ID != testWishlistID {
		t.Fatalf("GetWishlist() id = %q, want %q", output.Wishlist.ID, testWishlistID)
	}
	if output.Wishlist.Name != personalWishlistName {
		t.Fatalf("GetWishlist() name = %q, want %q", output.Wishlist.Name, personalWishlistName)
	}
	if repo.createdWishlist == nil {
		t.Fatal("GetWishlist() did not create personal wishlist")
	}
}

func TestServiceCreateWishlistRejectsSecondWishlist(t *testing.T) {
	t.Parallel()

	existingWishlist := mustWishlist(t, testWishlistUserID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: &fakeWishlistRepo{wishlistByUser: &existingWishlist},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	_, err := service.CreateWishlist(context.Background(), CreateWishlistInput{
		UserID: testWishlistUserID,
		Name:   "Что угодно",
	})
	if err == nil {
		t.Fatal("CreateWishlist() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindConflict {
		t.Fatalf("CreateWishlist() kind = %q, want %q", appErr.Kind(), apperrors.KindConflict)
	}
	if appErr.Code() != "wishlist_already_exists" {
		t.Fatalf("CreateWishlist() code = %q, want %q", appErr.Code(), "wishlist_already_exists")
	}
}

func TestServiceGetWishlistMasksForeignAccessByID(t *testing.T) {
	t.Parallel()

	foreignWishlist := mustWishlist(t, testWishlistOtherUserID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: &fakeWishlistRepo{wishlistByID: &foreignWishlist},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	_, err := service.GetWishlist(context.Background(), GetWishlistInput{
		UserID:     testWishlistUserID,
		WishlistID: testWishlistID,
	})
	if err == nil {
		t.Fatal("GetWishlist() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindNotFound {
		t.Fatalf("GetWishlist() kind = %q, want %q", appErr.Kind(), apperrors.KindNotFound)
	}
	if appErr.Code() != "wishlist_not_found" {
		t.Fatalf("GetWishlist() code = %q, want %q", appErr.Code(), "wishlist_not_found")
	}
}

func TestServiceAddWishlistItemMissingGift(t *testing.T) {
	t.Parallel()

	existingWishlist := mustWishlist(t, testWishlistUserID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: &fakeWishlistRepo{wishlistByUser: &existingWishlist},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	_, err := service.AddWishlistItem(context.Background(), AddWishlistItemInput{
		UserID: testWishlistUserID,
		GiftID: testWishlistGiftID,
	})
	if err == nil {
		t.Fatal("AddWishlistItem() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "gift_not_found" {
		t.Fatalf("AddWishlistItem() code = %q, want %q", appErr.Code(), "gift_not_found")
	}
}

func TestServiceAddWishlistItemDuplicateGiftIsIdempotent(t *testing.T) {
	t.Parallel()

	existingWishlist := mustWishlist(t, testWishlistUserID)
	existingItem := mustWishlistItem(t, existingWishlist.ID().String(), testWishlistGiftID, testWishlistItemID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: &fakeWishlistRepo{
			wishlistByUser: &existingWishlist,
			addErr:         wishlistdomain.ErrWishlistItemExists,
			itemByGift:     &existingItem,
		},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader: fakeGiftReader{
			gift: mustGift(t),
		},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: "550e8400-e29b-41d4-a716-446655440999"},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	output, err := service.AddWishlistItem(context.Background(), AddWishlistItemInput{
		UserID: testWishlistUserID,
		GiftID: testWishlistGiftID,
	})
	if err != nil {
		t.Fatalf("AddWishlistItem() error = %v", err)
	}

	if !output.AlreadyInWishlist {
		t.Fatal("AddWishlistItem() already_in_wishlist = false, want true")
	}
	if output.Item.ID != testWishlistItemID {
		t.Fatalf("AddWishlistItem() item id = %q, want %q", output.Item.ID, testWishlistItemID)
	}
}

func TestServiceRemoveWishlistItemReturnsFalseWhenWishlistMissing(t *testing.T) {
	t.Parallel()

	service := mustWishlistService(t, wishlistServiceDeps{
		repo: &fakeWishlistRepo{},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	output, err := service.RemoveWishlistItem(context.Background(), RemoveWishlistItemInput{
		UserID: testWishlistUserID,
		GiftID: testWishlistGiftID,
	})
	if err != nil {
		t.Fatalf("RemoveWishlistItem() error = %v", err)
	}
	if output.Removed {
		t.Fatal("RemoveWishlistItem() removed = true, want false")
	}
}

func TestServiceDeleteWishlistReturnsFalseWhenWishlistMissing(t *testing.T) {
	t.Parallel()

	service := mustWishlistService(t, wishlistServiceDeps{
		repo: &fakeWishlistRepo{},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	output, err := service.DeleteWishlist(context.Background(), DeleteWishlistInput{
		UserID: testWishlistUserID,
	})
	if err != nil {
		t.Fatalf("DeleteWishlist() error = %v", err)
	}
	if output.Deleted {
		t.Fatal("DeleteWishlist() deleted = true, want false")
	}
}

type wishlistServiceDeps struct {
	repo              *fakeWishlistRepo
	userReader        fakeUserReader
	giftReader        fakeGiftReader
	wishlistIDGen     fakeWishlistIDGenerator
	wishlistItemIDGen fakeWishlistItemIDGenerator
	clock             fixedWishlistClock
}

func mustWishlistService(t *testing.T, deps wishlistServiceDeps) *Service {
	t.Helper()

	service, err := NewService(
		deps.repo,
		deps.userReader,
		deps.giftReader,
		deps.wishlistIDGen,
		deps.wishlistItemIDGen,
		deps.clock,
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

type fakeWishlistRepo struct {
	createErr error
	addErr    error
	removeErr error
	deleteErr error

	createdWishlist *wishlistdomain.Wishlist
	wishlistByID    *wishlistdomain.Wishlist
	wishlistByUser  *wishlistdomain.Wishlist
	itemByGift      *wishlistdomain.WishlistItem
	items           []wishlistdomain.WishlistItem
	listRecords     []WishlistSummaryRecord
	listTotal       int
}

func (r *fakeWishlistRepo) CreateWishlist(_ context.Context, wishlist *wishlistdomain.Wishlist) error {
	if r.createErr != nil {
		return r.createErr
	}

	created := *wishlist
	r.createdWishlist = &created
	r.wishlistByUser = &created
	return nil
}

func (r *fakeWishlistRepo) GetWishlistByID(context.Context, wishlistdomain.WishlistID) (*wishlistdomain.Wishlist, error) {
	return r.wishlistByID, nil
}

func (r *fakeWishlistRepo) GetWishlistByUserID(context.Context, userdomain.UserID) (*wishlistdomain.Wishlist, error) {
	return r.wishlistByUser, nil
}

func (r *fakeWishlistRepo) ListWishlistsByUser(context.Context, userdomain.UserID, int, int) ([]WishlistSummaryRecord, int, error) {
	return r.listRecords, r.listTotal, nil
}

func (r *fakeWishlistRepo) ListWishlistItems(context.Context, wishlistdomain.WishlistID) ([]wishlistdomain.WishlistItem, error) {
	return r.items, nil
}

func (r *fakeWishlistRepo) GetWishlistItemByGiftID(context.Context, wishlistdomain.WishlistID, catalogdomain.GiftID) (*wishlistdomain.WishlistItem, error) {
	return r.itemByGift, nil
}

func (r *fakeWishlistRepo) AddWishlistItem(context.Context, *wishlistdomain.WishlistItem) error {
	return r.addErr
}

func (r *fakeWishlistRepo) RemoveWishlistItem(context.Context, wishlistdomain.WishlistID, catalogdomain.GiftID) error {
	return r.removeErr
}

func (r *fakeWishlistRepo) DeleteWishlist(context.Context, wishlistdomain.WishlistID) error {
	return r.deleteErr
}

type fakeUserReader struct {
	user *userdomain.User
}

func (r fakeUserReader) GetByID(context.Context, userdomain.UserID) (*userdomain.User, error) {
	return r.user, nil
}

type fakeGiftReader struct {
	gift *catalogdomain.Gift
}

func (r fakeGiftReader) GetGift(context.Context, catalogdomain.GiftID) (*catalogdomain.Gift, error) {
	return r.gift, nil
}

type fakeWishlistIDGenerator struct {
	id string
}

func (g fakeWishlistIDGenerator) NewWishlistID() (wishlistdomain.WishlistID, error) {
	return wishlistdomain.NewWishlistID(g.id)
}

type fakeWishlistItemIDGenerator struct {
	id string
}

func (g fakeWishlistItemIDGenerator) NewWishlistItemID() (wishlistdomain.WishlistItemID, error) {
	return wishlistdomain.NewWishlistItemID(g.id)
}

type fixedWishlistClock struct {
	now time.Time
}

func (c fixedWishlistClock) Now() time.Time {
	return c.now.UTC()
}

func mustWishlistUser(t *testing.T) *userdomain.User {
	t.Helper()

	user, err := userdomain.NewUser(
		testWishlistUserID,
		"wishlist@example.com",
		"ValidPass1!",
		string(userdomain.UserRoleUser),
	)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	return &user
}

func mustWishlist(t *testing.T, userID string) wishlistdomain.Wishlist {
	t.Helper()

	wishlist, err := wishlistdomain.RestoreWishlist(
		testWishlistID,
		userID,
		personalWishlistName,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreWishlist() error = %v", err)
	}

	return wishlist
}

func mustWishlistItem(t *testing.T, wishlistID, giftID, itemID string) wishlistdomain.WishlistItem {
	t.Helper()

	item, err := wishlistdomain.RestoreWishlistItem(
		itemID,
		wishlistID,
		giftID,
		time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreWishlistItem() error = %v", err)
	}

	return item
}

func mustGift(t *testing.T) *catalogdomain.Gift {
	t.Helper()

	gift, err := catalogdomain.RestoreGift(
		testWishlistGiftID,
		nil,
		nil,
		"LEGO Set",
		testWishlistDescription,
		testWishlistPrice,
		testWishlistStoreLink,
		nil,
		nil,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreGift() error = %v", err)
	}

	return &gift
}
