package wishlist

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/gift"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/user"
	"github.com/pkg/errors"
)

type WishlistUseCase struct {
	wishlistRepo WishlistRepository
	userRepo     UserRepository
	giftRepo     GiftRepository
}

func NewWishlistUseCase(wishlistRepo WishlistRepository, userRepo UserRepository, giftRepo GiftRepository) (*WishlistUseCase, error) {
	if wishlistRepo == nil {
		return nil, ErrNilWishlistRepository
	}
	if giftRepo == nil {
		return nil, ErrNilGiftRepository
	}
	if userRepo == nil {
		return nil, ErrNilUserRepository
	}
	return &WishlistUseCase{
		wishlistRepo: wishlistRepo,
		userRepo:     userRepo,
		giftRepo:     giftRepo,
	}, nil
}

func (wl *WishlistUseCase) SaveGift(ctx context.Context, input SaveGiftInput) error {
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to create user id")
	}

	giftID, err := gift.NewGiftID(input.GiftID)
	if err != nil {
		return errors.Wrap(err, "failed to create gift id")
	}

	if _, err = wl.userRepo.FindByID(ctx, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return errors.Wrap(err, "failed to find user")
	}

	if _, err = wl.giftRepo.FindByID(ctx, giftID); err != nil {
		if errors.Is(err, ErrGiftNotFound) {
			return ErrGiftNotFound
		}
		return errors.Wrap(err, "failed to find gift")
	}

	if err = wl.wishlistRepo.Save(ctx, userID, giftID); err != nil {
		return errors.Wrap(err, "failed to save gift")
	}

	return nil
}

func (wl *WishlistUseCase) GetAllGiftsByUserID(ctx context.Context, input GetGiftsByUserIDInput) ([]*gift.Gift, error) {
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user id")
	}

	if _, err := wl.userRepo.FindByID(ctx, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, errors.Wrap(err, "failed to find user")
	}

	gifts, err := wl.wishlistRepo.GetAll(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find gifts")
	}

	return gifts, nil
}

func (wl *WishlistUseCase) Delete(ctx context.Context, input DeleteGiftInput) error {
	userID, err := user.NewUserID(input.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to create user id")
	}

	if _, err = wl.userRepo.FindByID(ctx, userID); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return errors.Wrap(err, "failed to find user")
	}

	giftID, err := gift.NewGiftID(input.GiftID)
	if err != nil {
		return errors.Wrap(err, "failed to create gift id")
	}

	if err = wl.wishlistRepo.Delete(ctx, userID, giftID); err != nil {
		return errors.Wrap(err, "failed to delete gift")
	}

	return nil
}
