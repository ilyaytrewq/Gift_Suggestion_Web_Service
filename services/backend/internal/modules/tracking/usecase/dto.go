package usecase

import (
	"time"

	trackingdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/domain"
)

type TrackEventInput struct {
	UserID                  string
	Type                    string
	ClientEventID           *string
	RecommendationRequestID *string
	WishlistID              *string
	GiftID                  *string
	OccurredAt              *time.Time
	Metadata                EventMetadataInput
}

type EventMetadataInput struct {
	Surface  *string
	Position *int
}

type EventMetadata struct {
	Surface  *string `json:"surface,omitempty"`
	Position *int    `json:"position,omitempty"`
}

type Event struct {
	ID                      string        `json:"id"`
	Type                    string        `json:"type"`
	Duplicate               bool          `json:"duplicate"`
	ClientEventID           *string       `json:"client_event_id,omitempty"`
	RecommendationRequestID *string       `json:"recommendation_request_id,omitempty"`
	WishlistID              *string       `json:"wishlist_id,omitempty"`
	GiftID                  *string       `json:"gift_id,omitempty"`
	Metadata                EventMetadata `json:"metadata"`
	OccurredAt              time.Time     `json:"occurred_at"`
	RecordedAt              time.Time     `json:"recorded_at"`
}

type TrackEventOutput struct {
	Event Event `json:"event"`
}

func newEvent(item trackingdomain.Event, duplicate bool) Event {
	return Event{
		ID:                      item.ID().String(),
		Type:                    string(item.EventType()),
		Duplicate:               duplicate,
		ClientEventID:           item.ClientEventID(),
		RecommendationRequestID: requestIDToStringPtr(item.RecommendationRequestID()),
		WishlistID:              wishlistIDToStringPtr(item.WishlistID()),
		GiftID:                  giftIDToStringPtr(item.GiftID()),
		Metadata: EventMetadata{
			Surface:  item.Metadata().Surface(),
			Position: item.Metadata().Position(),
		},
		OccurredAt: item.OccurredAt(),
		RecordedAt: item.CreatedAt(),
	}
}
