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
	testWishlistGiftName    = "LEGO Set"
	testWishlistName        = "Birthday Ideas"
	testWishlistStoreLink   = "https://example.com/gifts/lego"
	testWishlistPrice       = "129.99"
	testWishlistDescription = "Creative building set"
)

func TestServiceCreateWishlistSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: fakeWishlistRepo{},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: now},
	})

	output, err := service.CreateWishlist(context.Background(), CreateWishlistInput{
		UserID: testWishlistUserID,
		Name:   "  Birthday Ideas  ",
	})
	if err != nil {
		t.Fatalf("CreateWishlist() error = %v", err)
	}

	if output.Wishlist.ID != testWishlistID {
		t.Fatalf("CreateWishlist() id = %q, want %q", output.Wishlist.ID, testWishlistID)
	}
	if output.Wishlist.Name != testWishlistName {
		t.Fatalf("CreateWishlist() name = %q, want %q", output.Wishlist.Name, testWishlistName)
	}
	if output.Wishlist.ItemCount != 0 {
		t.Fatalf("CreateWishlist() item count = %d, want 0", output.Wishlist.ItemCount)
	}
}

func TestServiceCreateWishlistDuplicateName(t *testing.T) {
	t.Parallel()

	service := mustWishlistService(t, wishlistServiceDeps{
		repo: fakeWishlistRepo{createErr: wishlistdomain.ErrWishlistAlreadyExists},
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
		Name:   testWishlistName,
	})
	if err == nil {
		t.Fatal("CreateWishlist() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindConflict {
		t.Fatalf("CreateWishlist() kind = %q, want %q", appErr.Kind(), apperrors.KindConflict)
	}
	if appErr.Code() != "wishlist_name_exists" {
		t.Fatalf("CreateWishlist() code = %q, want %q", appErr.Code(), "wishlist_name_exists")
	}
}

func TestServiceCreateWishlistMissingUser(t *testing.T) {
	t.Parallel()

	service := mustWishlistService(t, wishlistServiceDeps{
		repo:              fakeWishlistRepo{},
		userReader:        fakeUserReader{},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	_, err := service.CreateWishlist(context.Background(), CreateWishlistInput{
		UserID: testWishlistUserID,
		Name:   testWishlistName,
	})
	if err == nil {
		t.Fatal("CreateWishlist() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindNotFound {
		t.Fatalf("CreateWishlist() kind = %q, want %q", appErr.Kind(), apperrors.KindNotFound)
	}
	if appErr.Code() != "user_not_found" {
		t.Fatalf("CreateWishlist() code = %q, want %q", appErr.Code(), "user_not_found")
	}
}

func TestServiceGetWishlistMasksForeignAccess(t *testing.T) {
	t.Parallel()

	foreignWishlist := mustWishlist(t, testWishlistOtherUserID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: fakeWishlistRepo{wishlist: &foreignWishlist},
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
		repo: fakeWishlistRepo{wishlist: &existingWishlist},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	_, err := service.AddWishlistItem(context.Background(), AddWishlistItemInput{
		UserID:     testWishlistUserID,
		WishlistID: testWishlistID,
		GiftID:     testWishlistGiftID,
	})
	if err == nil {
		t.Fatal("AddWishlistItem() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "gift_not_found" {
		t.Fatalf("AddWishlistItem() code = %q, want %q", appErr.Code(), "gift_not_found")
	}
}

func TestServiceAddWishlistItemDuplicateGift(t *testing.T) {
	t.Parallel()

	existingWishlist := mustWishlist(t, testWishlistUserID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: fakeWishlistRepo{
			wishlist: &existingWishlist,
			addErr:   wishlistdomain.ErrWishlistItemExists,
		},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader: fakeGiftReader{
			gift: mustGift(t),
		},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	_, err := service.AddWishlistItem(context.Background(), AddWishlistItemInput{
		UserID:     testWishlistUserID,
		WishlistID: testWishlistID,
		GiftID:     testWishlistGiftID,
	})
	if err == nil {
		t.Fatal("AddWishlistItem() expected conflict error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindConflict {
		t.Fatalf("AddWishlistItem() kind = %q, want %q", appErr.Kind(), apperrors.KindConflict)
	}
}

func TestServiceGetWishlistEmptySuccess(t *testing.T) {
	t.Parallel()

	existingWishlist := mustWishlist(t, testWishlistUserID)
	service := mustWishlistService(t, wishlistServiceDeps{
		repo: fakeWishlistRepo{
			wishlist: &existingWishlist,
			items:    nil,
		},
		userReader: fakeUserReader{
			user: mustWishlistUser(t),
		},
		giftReader:        fakeGiftReader{},
		wishlistIDGen:     fakeWishlistIDGenerator{id: testWishlistID},
		wishlistItemIDGen: fakeWishlistItemIDGenerator{id: testWishlistItemID},
		clock:             fixedWishlistClock{now: time.Now().UTC()},
	})

	output, err := service.GetWishlist(context.Background(), GetWishlistInput{
		UserID:     testWishlistUserID,
		WishlistID: testWishlistID,
	})
	if err != nil {
		t.Fatalf("GetWishlist() error = %v", err)
	}

	if len(output.Wishlist.Items) != 0 {
		t.Fatalf("GetWishlist() items = %d, want 0", len(output.Wishlist.Items))
	}
}

type wishlistServiceDeps struct {
	repo              fakeWishlistRepo
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

	wishlist    *wishlistdomain.Wishlist
	listRecords []WishlistSummaryRecord
	listTotal   int
	items       []wishlistdomain.WishlistItem
}

func (r fakeWishlistRepo) CreateWishlist(context.Context, *wishlistdomain.Wishlist) error {
	return r.createErr
}

func (r fakeWishlistRepo) GetWishlistByID(context.Context, wishlistdomain.WishlistID) (*wishlistdomain.Wishlist, error) {
	return r.wishlist, nil
}

func (r fakeWishlistRepo) ListWishlistsByUser(context.Context, userdomain.UserID, int, int) ([]WishlistSummaryRecord, int, error) {
	return r.listRecords, r.listTotal, nil
}

func (r fakeWishlistRepo) ListWishlistItems(context.Context, wishlistdomain.WishlistID) ([]wishlistdomain.WishlistItem, error) {
	return r.items, nil
}

func (r fakeWishlistRepo) AddWishlistItem(context.Context, *wishlistdomain.WishlistItem) error {
	return r.addErr
}

func (r fakeWishlistRepo) RemoveWishlistItem(context.Context, wishlistdomain.WishlistID, catalogdomain.GiftID) error {
	return r.removeErr
}

func (r fakeWishlistRepo) DeleteWishlist(context.Context, wishlistdomain.WishlistID) error {
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

func mustWishlist(
	t *testing.T,
	userID string,
) wishlistdomain.Wishlist {
	t.Helper()

	wishlist, err := wishlistdomain.RestoreWishlist(
		testWishlistID,
		userID,
		testWishlistName,
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreWishlist() error = %v", err)
	}

	return wishlist
}

func mustGift(t *testing.T) *catalogdomain.Gift {
	t.Helper()

	gift, err := catalogdomain.RestoreGift(
		testWishlistGiftID,
		nil,
		nil,
		testWishlistGiftName,
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
