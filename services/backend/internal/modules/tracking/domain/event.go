package domain

import (
	"strings"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
)

type EventType string

const (
	EventTypeRecommendationRequest EventType = "recommendation_request"
	EventTypeCardView              EventType = "card_view"
	EventTypeWishlistAdd           EventType = "wishlist_add"
	EventTypeOutboundClick         EventType = "outbound_click"
)

const maxClientEventIDLength = 64

type EventMetadata struct {
	surface  *string
	position *int
}

type Event struct {
	id                      EventID
	eventType               EventType
	userID                  userdomain.UserID
	recommendationRequestID *recommendationdomain.RequestID
	wishlistID              *wishlistdomain.WishlistID
	giftID                  *catalogdomain.GiftID
	clientEventID           *string
	metadata                EventMetadata
	occurredAt              time.Time
	createdAt               time.Time
}

func NewEventType(raw string) (EventType, error) {
	switch EventType(strings.TrimSpace(raw)) {
	case EventTypeRecommendationRequest:
		return EventTypeRecommendationRequest, nil
	case EventTypeCardView:
		return EventTypeCardView, nil
	case EventTypeWishlistAdd:
		return EventTypeWishlistAdd, nil
	case EventTypeOutboundClick:
		return EventTypeOutboundClick, nil
	default:
		return "", ErrInvalidEventType
	}
}

func NewEventMetadata(surface *string, position *int) (EventMetadata, error) {
	var normalizedSurface *string
	if surface != nil {
		value := strings.TrimSpace(*surface)
		if len(value) > 32 {
			return EventMetadata{}, ErrMetadataSurfaceTooLong
		}
		if value != "" {
			switch value {
			case "catalog", "recommendation", "wishlist", "direct":
				normalizedSurface = &value
			default:
				return EventMetadata{}, ErrInvalidMetadataSurface
			}
		}
	}

	if position != nil && (*position < 1 || *position > 100) {
		return EventMetadata{}, ErrInvalidMetadataPosition
	}

	return EventMetadata{
		surface:  cloneStringPtr(normalizedSurface),
		position: cloneIntPtr(position),
	}, nil
}

func NewEvent(
	id EventID,
	eventType EventType,
	userID userdomain.UserID,
	recommendationRequestID *recommendationdomain.RequestID,
	wishlistID *wishlistdomain.WishlistID,
	giftID *catalogdomain.GiftID,
	clientEventID *string,
	metadata EventMetadata,
	occurredAt time.Time,
	now time.Time,
) (Event, error) {
	normalizedClientEventID, err := normalizeClientEventID(clientEventID)
	if err != nil {
		return Event{}, err
	}

	if occurredAt.IsZero() {
		return Event{}, ErrOccurredAtZero
	}

	if err := validateEventRefs(eventType, recommendationRequestID, wishlistID, giftID); err != nil {
		return Event{}, err
	}

	return Event{
		id:                      id,
		eventType:               eventType,
		userID:                  userID,
		recommendationRequestID: cloneRecommendationRequestID(recommendationRequestID),
		wishlistID:              cloneWishlistID(wishlistID),
		giftID:                  cloneGiftID(giftID),
		clientEventID:           cloneStringPtr(normalizedClientEventID),
		metadata:                metadata,
		occurredAt:              occurredAt.UTC(),
		createdAt:               now.UTC(),
	}, nil
}

func RestoreEvent(
	id string,
	eventType string,
	userID string,
	recommendationRequestID *string,
	wishlistID *string,
	giftID *string,
	clientEventID *string,
	metadata EventMetadata,
	occurredAt time.Time,
	createdAt time.Time,
) (Event, error) {
	eventID, err := NewEventID(id)
	if err != nil {
		return Event{}, err
	}

	parsedEventType, err := NewEventType(eventType)
	if err != nil {
		return Event{}, err
	}

	parsedUserID, err := userdomain.NewUserID(userID)
	if err != nil {
		return Event{}, err
	}

	var parsedRecommendationRequestID *recommendationdomain.RequestID
	if recommendationRequestID != nil && strings.TrimSpace(*recommendationRequestID) != "" {
		value, err := recommendationdomain.NewRequestID(strings.TrimSpace(*recommendationRequestID))
		if err != nil {
			return Event{}, err
		}
		parsedRecommendationRequestID = &value
	}

	var parsedWishlistID *wishlistdomain.WishlistID
	if wishlistID != nil && strings.TrimSpace(*wishlistID) != "" {
		value, err := wishlistdomain.NewWishlistID(strings.TrimSpace(*wishlistID))
		if err != nil {
			return Event{}, err
		}
		parsedWishlistID = &value
	}

	var parsedGiftID *catalogdomain.GiftID
	if giftID != nil && strings.TrimSpace(*giftID) != "" {
		value, err := catalogdomain.NewGiftID(strings.TrimSpace(*giftID))
		if err != nil {
			return Event{}, err
		}
		parsedGiftID = &value
	}

	normalizedClientEventID, err := normalizeClientEventID(clientEventID)
	if err != nil {
		return Event{}, err
	}

	if err := validateEventRefs(parsedEventType, parsedRecommendationRequestID, parsedWishlistID, parsedGiftID); err != nil {
		return Event{}, err
	}

	return Event{
		id:                      eventID,
		eventType:               parsedEventType,
		userID:                  parsedUserID,
		recommendationRequestID: cloneRecommendationRequestID(parsedRecommendationRequestID),
		wishlistID:              cloneWishlistID(parsedWishlistID),
		giftID:                  cloneGiftID(parsedGiftID),
		clientEventID:           cloneStringPtr(normalizedClientEventID),
		metadata:                metadata,
		occurredAt:              occurredAt.UTC(),
		createdAt:               createdAt.UTC(),
	}, nil
}

func (e Event) ID() EventID {
	return e.id
}

func (e Event) EventType() EventType {
	return e.eventType
}

func (e Event) UserID() userdomain.UserID {
	return e.userID
}

func (e Event) RecommendationRequestID() *recommendationdomain.RequestID {
	return cloneRecommendationRequestID(e.recommendationRequestID)
}

func (e Event) WishlistID() *wishlistdomain.WishlistID {
	return cloneWishlistID(e.wishlistID)
}

func (e Event) GiftID() *catalogdomain.GiftID {
	return cloneGiftID(e.giftID)
}

func (e Event) ClientEventID() *string {
	return cloneStringPtr(e.clientEventID)
}

func (e Event) Metadata() EventMetadata {
	return e.metadata
}

func (e Event) OccurredAt() time.Time {
	return e.occurredAt
}

func (e Event) CreatedAt() time.Time {
	return e.createdAt
}

func (e Event) SameClientPayload(other Event) bool {
	return e.eventType == other.eventType &&
		e.userID.String() == other.userID.String() &&
		recommendationRequestIDEq(e.recommendationRequestID, other.recommendationRequestID) &&
		wishlistIDEq(e.wishlistID, other.wishlistID) &&
		giftIDEq(e.giftID, other.giftID) &&
		stringPtrEq(e.clientEventID, other.clientEventID) &&
		e.metadata.Equal(other.metadata) &&
		e.occurredAt.Equal(other.occurredAt)
}

func (m EventMetadata) Surface() *string {
	return cloneStringPtr(m.surface)
}

func (m EventMetadata) Position() *int {
	return cloneIntPtr(m.position)
}

func (m EventMetadata) Equal(other EventMetadata) bool {
	return stringPtrEq(m.surface, other.surface) && intPtrEq(m.position, other.position)
}

func validateEventRefs(
	eventType EventType,
	recommendationRequestID *recommendationdomain.RequestID,
	wishlistID *wishlistdomain.WishlistID,
	giftID *catalogdomain.GiftID,
) error {
	switch eventType {
	case EventTypeRecommendationRequest:
		if recommendationRequestID == nil {
			return ErrRecommendationRequestIDRequired
		}
		if giftID != nil {
			return ErrGiftIDForbidden
		}
		if wishlistID != nil {
			return ErrWishlistIDForbidden
		}
	case EventTypeCardView, EventTypeOutboundClick:
		if giftID == nil {
			return ErrGiftIDRequired
		}
		if wishlistID != nil {
			return ErrWishlistIDForbidden
		}
	case EventTypeWishlistAdd:
		if giftID == nil {
			return ErrGiftIDRequired
		}
		if wishlistID == nil {
			return ErrWishlistIDRequired
		}
	default:
		return ErrInvalidEventType
	}

	return nil
}

func normalizeClientEventID(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, nil
	}
	if len(normalized) > maxClientEventIDLength {
		return nil, ErrClientEventIDTooLong
	}

	return &normalized, nil
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	current := *value
	return &current
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}

	current := *value
	return &current
}

func cloneRecommendationRequestID(value *recommendationdomain.RequestID) *recommendationdomain.RequestID {
	if value == nil {
		return nil
	}

	current := *value
	return &current
}

func cloneWishlistID(value *wishlistdomain.WishlistID) *wishlistdomain.WishlistID {
	if value == nil {
		return nil
	}

	current := *value
	return &current
}

func cloneGiftID(value *catalogdomain.GiftID) *catalogdomain.GiftID {
	if value == nil {
		return nil
	}

	current := *value
	return &current
}

func stringPtrEq(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return *left == *right
}

func intPtrEq(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return *left == *right
}

func recommendationRequestIDEq(left, right *recommendationdomain.RequestID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return left.String() == right.String()
}

func wishlistIDEq(left, right *wishlistdomain.WishlistID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return left.String() == right.String()
}

func giftIDEq(left, right *catalogdomain.GiftID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return left.String() == right.String()
}
