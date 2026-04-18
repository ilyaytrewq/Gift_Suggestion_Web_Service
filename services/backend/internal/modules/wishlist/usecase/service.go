package usecase

import (
	"context"
	"errors"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo              WishlistRepository
	userReader        UserReader
	giftReader        GiftReader
	wishlistIDGen     WishlistIDGenerator
	wishlistItemIDGen WishlistItemIDGenerator
	clock             Clock
}

func NewService(
	repo WishlistRepository,
	userReader UserReader,
	giftReader GiftReader,
	wishlistIDGen WishlistIDGenerator,
	wishlistItemIDGen WishlistItemIDGenerator,
	clock Clock,
) (*Service, error) {
	switch {
	case repo == nil:
		return nil, ErrNilWishlistRepository
	case userReader == nil:
		return nil, ErrNilUserReader
	case giftReader == nil:
		return nil, ErrNilGiftReader
	case wishlistIDGen == nil:
		return nil, ErrNilWishlistIDGenerator
	case wishlistItemIDGen == nil:
		return nil, ErrNilWishlistItemIDGenerator
	case clock == nil:
		return nil, ErrNilClock
	}

	return &Service{
		repo:              repo,
		userReader:        userReader,
		giftReader:        giftReader,
		wishlistIDGen:     wishlistIDGen,
		wishlistItemIDGen: wishlistItemIDGen,
		clock:             clock,
	}, nil
}

func (s *Service) CreateWishlist(ctx context.Context, input CreateWishlistInput) (CreateWishlistOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return CreateWishlistOutput{}, err
	}

	wishlistID, err := s.wishlistIDGen.NewWishlistID()
	if err != nil {
		return CreateWishlistOutput{}, err
	}

	wishlist, err := wishlistdomain.NewWishlist(wishlistID, userID, input.Name, s.clock.Now())
	if err != nil {
		return CreateWishlistOutput{}, mapWishlistValidationError(err)
	}

	if err := s.repo.CreateWishlist(ctx, &wishlist); err != nil {
		if errors.Is(err, wishlistdomain.ErrWishlistAlreadyExists) {
			return CreateWishlistOutput{}, apperrors.New(
				apperrors.KindConflict,
				"wishlist_name_exists",
				"wishlist name already exists",
			)
		}

		return CreateWishlistOutput{}, err
	}

	return CreateWishlistOutput{
		Wishlist: newWishlist(wishlist, nil),
	}, nil
}

func (s *Service) ListWishlists(ctx context.Context, input ListWishlistsInput) (ListWishlistsOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return ListWishlistsOutput{}, err
	}

	limit, err := normalizeLimit(input.Limit)
	if err != nil {
		return ListWishlistsOutput{}, err
	}
	offset, err := normalizeOffset(input.Offset)
	if err != nil {
		return ListWishlistsOutput{}, err
	}

	records, total, err := s.repo.ListWishlistsByUser(ctx, userID, limit, offset)
	if err != nil {
		return ListWishlistsOutput{}, err
	}

	items := make([]WishlistSummary, 0, len(records))
	for _, record := range records {
		items = append(items, newWishlistSummary(record))
	}

	return ListWishlistsOutput{
		Items: items,
		Page: Page{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}, nil
}

func (s *Service) GetWishlist(ctx context.Context, input GetWishlistInput) (GetWishlistOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return GetWishlistOutput{}, err
	}

	wishlist, err := s.loadOwnedWishlist(ctx, userID, input.WishlistID)
	if err != nil {
		return GetWishlistOutput{}, err
	}

	items, err := s.repo.ListWishlistItems(ctx, wishlist.ID())
	if err != nil {
		return GetWishlistOutput{}, err
	}

	outputItems := make([]WishlistItem, 0, len(items))
	for _, item := range items {
		gift, err := s.giftReader.GetGift(ctx, item.GiftID())
		if err != nil {
			return GetWishlistOutput{}, err
		}
		if gift == nil {
			continue
		}

		outputItems = append(outputItems, newWishlistItem(item, *gift))
	}

	return GetWishlistOutput{
		Wishlist: newWishlist(*wishlist, outputItems),
	}, nil
}

func (s *Service) AddWishlistItem(ctx context.Context, input AddWishlistItemInput) (AddWishlistItemOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return AddWishlistItemOutput{}, err
	}

	wishlist, err := s.loadOwnedWishlist(ctx, userID, input.WishlistID)
	if err != nil {
		return AddWishlistItemOutput{}, err
	}

	giftID, gift, err := s.loadGift(ctx, input.GiftID)
	if err != nil {
		return AddWishlistItemOutput{}, err
	}

	itemID, err := s.wishlistItemIDGen.NewWishlistItemID()
	if err != nil {
		return AddWishlistItemOutput{}, err
	}

	item := wishlistdomain.NewWishlistItem(itemID, wishlist.ID(), giftID, s.clock.Now())
	if err := s.repo.AddWishlistItem(ctx, &item); err != nil {
		if errors.Is(err, wishlistdomain.ErrWishlistItemExists) {
			return AddWishlistItemOutput{}, apperrors.New(
				apperrors.KindConflict,
				"wishlist_item_exists",
				"gift is already added to wishlist",
			)
		}

		return AddWishlistItemOutput{}, err
	}

	return AddWishlistItemOutput{
		Item: newWishlistItem(item, *gift),
	}, nil
}

func (s *Service) RemoveWishlistItem(ctx context.Context, input RemoveWishlistItemInput) (RemoveWishlistItemOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return RemoveWishlistItemOutput{}, err
	}

	wishlist, err := s.loadOwnedWishlist(ctx, userID, input.WishlistID)
	if err != nil {
		return RemoveWishlistItemOutput{}, err
	}

	giftID, err := catalogdomain.NewGiftID(input.GiftID)
	if err != nil {
		return RemoveWishlistItemOutput{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_gift_id",
			"gift id is invalid",
			err,
		)
	}

	if err := s.repo.RemoveWishlistItem(ctx, wishlist.ID(), giftID); err != nil {
		if errors.Is(err, wishlistdomain.ErrWishlistItemNotFound) {
			return RemoveWishlistItemOutput{}, apperrors.New(
				apperrors.KindNotFound,
				"wishlist_item_not_found",
				"wishlist item not found",
			)
		}

		return RemoveWishlistItemOutput{}, err
	}

	return RemoveWishlistItemOutput{Removed: true}, nil
}

func (s *Service) DeleteWishlist(ctx context.Context, input DeleteWishlistInput) (DeleteWishlistOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return DeleteWishlistOutput{}, err
	}

	wishlist, err := s.loadOwnedWishlist(ctx, userID, input.WishlistID)
	if err != nil {
		return DeleteWishlistOutput{}, err
	}

	if err := s.repo.DeleteWishlist(ctx, wishlist.ID()); err != nil {
		if errors.Is(err, wishlistdomain.ErrWishlistNotFound) {
			return DeleteWishlistOutput{}, apperrors.New(
				apperrors.KindNotFound,
				"wishlist_not_found",
				"wishlist not found",
			)
		}

		return DeleteWishlistOutput{}, err
	}

	return DeleteWishlistOutput{Deleted: true}, nil
}

func (s *Service) ensureUserExists(ctx context.Context, rawUserID string) (userdomain.UserID, error) {
	userID, err := userdomain.NewUserID(rawUserID)
	if err != nil {
		return userdomain.UserID{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_user_id",
			"user id is invalid",
			err,
		)
	}

	user, err := s.userReader.GetByID(ctx, userID)
	if err != nil {
		return userdomain.UserID{}, err
	}
	if user == nil {
		return userdomain.UserID{}, apperrors.New(
			apperrors.KindNotFound,
			"user_not_found",
			"user not found",
		)
	}

	return userID, nil
}

func (s *Service) loadOwnedWishlist(
	ctx context.Context,
	userID userdomain.UserID,
	rawWishlistID string,
) (*wishlistdomain.Wishlist, error) {
	wishlistID, err := wishlistdomain.NewWishlistID(rawWishlistID)
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_wishlist_id",
			"wishlist id is invalid",
			err,
		)
	}

	wishlist, err := s.repo.GetWishlistByID(ctx, wishlistID)
	if err != nil {
		return nil, err
	}
	if wishlist == nil || wishlist.UserID().String() != userID.String() {
		return nil, apperrors.New(
			apperrors.KindNotFound,
			"wishlist_not_found",
			"wishlist not found",
		)
	}

	return wishlist, nil
}

func (s *Service) loadGift(
	ctx context.Context,
	rawGiftID string,
) (catalogdomain.GiftID, *catalogdomain.Gift, error) {
	giftID, err := catalogdomain.NewGiftID(rawGiftID)
	if err != nil {
		return catalogdomain.GiftID{}, nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_gift_id",
			"gift id is invalid",
			err,
		)
	}

	gift, err := s.giftReader.GetGift(ctx, giftID)
	if err != nil {
		return catalogdomain.GiftID{}, nil, err
	}
	if gift == nil {
		return catalogdomain.GiftID{}, nil, apperrors.New(
			apperrors.KindNotFound,
			"gift_not_found",
			"gift not found",
		)
	}

	return giftID, gift, nil
}

func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultLimit, nil
	}
	if limit < 1 || limit > maxLimit {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_limit",
			"limit must be between 1 and 100",
		)
	}

	return limit, nil
}

func normalizeOffset(offset int) (int, error) {
	if offset < 0 {
		return 0, apperrors.New(
			apperrors.KindValidation,
			"invalid_offset",
			"offset must be greater than or equal to zero",
		)
	}

	return offset, nil
}

func mapWishlistValidationError(err error) error {
	switch {
	case errors.Is(err, wishlistdomain.ErrWishlistNameEmpty), errors.Is(err, wishlistdomain.ErrWishlistNameTooLong):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_wishlist_name",
			"wishlist name is invalid",
			err,
		)
	default:
		return err
	}
}
