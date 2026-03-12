package wishlist

import "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/gift"

type SaveGiftInput struct {
	UserID string
	GiftID string
}

type DeleteGiftInput struct {
	UserID string
	GiftID string
}

type GetGiftsByUserIDInput struct {
	UserID string
}

type GetAllGiftsByUserIDOutput struct {
	Gifts []*gift.Gift
}
