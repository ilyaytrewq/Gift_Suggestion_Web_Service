package usecase

import (
	"context"
	"errors"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	trackingdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type Service struct {
	repo                        EventRepository
	userReader                  UserReader
	giftReader                  GiftReader
	wishlistReader              WishlistReader
	recommendationRequestReader RecommendationRequestReader
	eventIDGen                  EventIDGenerator
	clock                       Clock
}

func NewService(
	repo EventRepository,
	userReader UserReader,
	giftReader GiftReader,
	wishlistReader WishlistReader,
	recommendationRequestReader RecommendationRequestReader,
	eventIDGen EventIDGenerator,
	clock Clock,
) (*Service, error) {
	switch {
	case repo == nil:
		return nil, ErrNilEventRepository
	case userReader == nil:
		return nil, ErrNilUserReader
	case giftReader == nil:
		return nil, ErrNilGiftReader
	case wishlistReader == nil:
		return nil, ErrNilWishlistReader
	case recommendationRequestReader == nil:
		return nil, ErrNilRecommendationRequestReader
	case eventIDGen == nil:
		return nil, ErrNilEventIDGenerator
	case clock == nil:
		return nil, ErrNilClock
	}

	return &Service{
		repo:                        repo,
		userReader:                  userReader,
		giftReader:                  giftReader,
		wishlistReader:              wishlistReader,
		recommendationRequestReader: recommendationRequestReader,
		eventIDGen:                  eventIDGen,
		clock:                       clock,
	}, nil
}

func (s *Service) TrackEvent(ctx context.Context, input TrackEventInput) (TrackEventOutput, error) {
	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return TrackEventOutput{}, err
	}

	eventType, err := trackingdomain.NewEventType(input.Type)
	if err != nil {
		return TrackEventOutput{}, mapEventValidationError(err)
	}

	recommendationRequestID, err := parseRecommendationRequestID(input.RecommendationRequestID)
	if err != nil {
		return TrackEventOutput{}, err
	}
	wishlistID, err := parseWishlistID(input.WishlistID)
	if err != nil {
		return TrackEventOutput{}, err
	}
	giftID, err := parseGiftID(input.GiftID)
	if err != nil {
		return TrackEventOutput{}, err
	}

	if err := s.ensureRecommendationRequestAllowed(ctx, userID, recommendationRequestID); err != nil {
		return TrackEventOutput{}, err
	}
	if err := s.ensureWishlistAllowed(ctx, userID, wishlistID); err != nil {
		return TrackEventOutput{}, err
	}
	if err := s.ensureGiftExists(ctx, giftID); err != nil {
		return TrackEventOutput{}, err
	}

	metadata, err := trackingdomain.NewEventMetadata(input.Metadata.Surface, input.Metadata.Position)
	if err != nil {
		return TrackEventOutput{}, mapEventValidationError(err)
	}

	occurredAt := s.clock.Now()
	if input.OccurredAt != nil {
		occurredAt = input.OccurredAt.UTC()
	}

	eventID, err := s.eventIDGen.NewTrackingEventID()
	if err != nil {
		return TrackEventOutput{}, err
	}

	event, err := trackingdomain.NewEvent(
		eventID,
		eventType,
		userID,
		recommendationRequestID,
		wishlistID,
		giftID,
		input.ClientEventID,
		metadata,
		occurredAt,
		s.clock.Now(),
	)
	if err != nil {
		return TrackEventOutput{}, mapEventValidationError(err)
	}

	stored, created, err := s.repo.CreateEvent(ctx, &event)
	if err != nil {
		return TrackEventOutput{}, err
	}
	if !created && !event.SameClientPayload(stored) {
		return TrackEventOutput{}, apperrors.New(
			apperrors.KindConflict,
			"tracking_event_id_reused",
			"client_event_id is already used for another event",
		)
	}

	return TrackEventOutput{
		Event: newEvent(stored, !created),
	}, nil
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

func (s *Service) ensureRecommendationRequestAllowed(
	ctx context.Context,
	userID userdomain.UserID,
	requestID *recommendationdomain.RequestID,
) error {
	if requestID == nil {
		return nil
	}

	request, err := s.recommendationRequestReader.GetRequest(ctx, *requestID)
	if err != nil {
		return err
	}
	if request == nil || request.RequestedByUserID() == nil || request.RequestedByUserID().String() != userID.String() {
		return apperrors.New(
			apperrors.KindNotFound,
			"recommendation_request_not_found",
			"recommendation request not found",
		)
	}

	return nil
}

func (s *Service) ensureWishlistAllowed(ctx context.Context, userID userdomain.UserID, wishlistID *wishlistdomain.WishlistID) error {
	if wishlistID == nil {
		return nil
	}

	wishlist, err := s.wishlistReader.GetWishlistByID(ctx, *wishlistID)
	if err != nil {
		return err
	}
	if wishlist == nil || wishlist.UserID().String() != userID.String() {
		return apperrors.New(
			apperrors.KindNotFound,
			"wishlist_not_found",
			"wishlist not found",
		)
	}

	return nil
}

func (s *Service) ensureGiftExists(ctx context.Context, giftID *catalogdomain.GiftID) error {
	if giftID == nil {
		return nil
	}

	gift, err := s.giftReader.GetGift(ctx, *giftID)
	if err != nil {
		return err
	}
	if gift == nil {
		return apperrors.New(
			apperrors.KindNotFound,
			"gift_not_found",
			"gift not found",
		)
	}

	return nil
}

func parseRecommendationRequestID(raw *string) (*recommendationdomain.RequestID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	value, err := recommendationdomain.NewRequestID(*raw)
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_recommendation_request_id",
			"recommendation request id is invalid",
			err,
		)
	}

	return &value, nil
}

func parseWishlistID(raw *string) (*wishlistdomain.WishlistID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	value, err := wishlistdomain.NewWishlistID(*raw)
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_wishlist_id",
			"wishlist id is invalid",
			err,
		)
	}

	return &value, nil
}

func parseGiftID(raw *string) (*catalogdomain.GiftID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	value, err := catalogdomain.NewGiftID(*raw)
	if err != nil {
		return nil, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_gift_id",
			"gift id is invalid",
			err,
		)
	}

	return &value, nil
}

func mapEventValidationError(err error) error {
	switch {
	case errors.Is(err, trackingdomain.ErrInvalidEventType):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_event_type",
			"event type is invalid",
			err,
		)
	case errors.Is(err, trackingdomain.ErrRecommendationRequestIDRequired),
		errors.Is(err, trackingdomain.ErrGiftIDRequired),
		errors.Is(err, trackingdomain.ErrWishlistIDRequired),
		errors.Is(err, trackingdomain.ErrGiftIDForbidden),
		errors.Is(err, trackingdomain.ErrWishlistIDForbidden),
		errors.Is(err, trackingdomain.ErrOccurredAtZero):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_event_payload",
			"event payload is invalid",
			err,
		)
	case errors.Is(err, trackingdomain.ErrClientEventIDTooLong):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_client_event_id",
			"client event id is invalid",
			err,
		)
	case errors.Is(err, trackingdomain.ErrMetadataSurfaceTooLong),
		errors.Is(err, trackingdomain.ErrInvalidMetadataSurface),
		errors.Is(err, trackingdomain.ErrInvalidMetadataPosition):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_event_metadata",
			"event metadata is invalid",
			err,
		)
	default:
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_tracking_event",
			"tracking event is invalid",
			err,
		)
	}
}

func requestIDToStringPtr(value *recommendationdomain.RequestID) *string {
	if value == nil {
		return nil
	}

	current := value.String()
	return &current
}

func wishlistIDToStringPtr(value *wishlistdomain.WishlistID) *string {
	if value == nil {
		return nil
	}

	current := value.String()
	return &current
}

func giftIDToStringPtr(value *catalogdomain.GiftID) *string {
	if value == nil {
		return nil
	}

	current := value.String()
	return &current
}
