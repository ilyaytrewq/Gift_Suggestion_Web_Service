package usecase

import "errors"

var (
	ErrNilCandidateReader    = errors.New("candidate reader is nil")
	ErrNilRequestRepository  = errors.New("recommendation request repository is nil")
	ErrNilUserReader         = errors.New("user reader is nil")
	ErrNilWishlistReader     = errors.New("wishlist reader is nil")
	ErrNilGiftReader         = errors.New("gift reader is nil")
	ErrNilRankingGateway     = errors.New("ranking gateway is nil")
	ErrNilRequestIDGenerator = errors.New("recommendation request id generator is nil")
	ErrNilResultIDGenerator  = errors.New("recommendation result id generator is nil")
	ErrNilClock              = errors.New("clock is nil")

	ErrRankingUnavailable     = errors.New("ranking service is unavailable")
	ErrRankingTimedOut        = errors.New("ranking service timed out")
	ErrRankingNotImplemented  = errors.New("ranking service is not implemented")
	ErrInvalidRankingResponse = errors.New("ranking response is invalid")
)
