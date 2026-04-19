package http

import "time"

type trackEventRequest struct {
	Type                    string              `json:"type"`
	ClientEventID           *string             `json:"client_event_id,omitempty"`
	RecommendationRequestID *string             `json:"recommendation_request_id,omitempty"`
	WishlistID              *string             `json:"wishlist_id,omitempty"`
	GiftID                  *string             `json:"gift_id,omitempty"`
	OccurredAt              *time.Time          `json:"occurred_at,omitempty"`
	Metadata                *trackEventMetadata `json:"metadata,omitempty"`
}

type trackEventMetadata struct {
	Surface  *string `json:"surface,omitempty"`
	Position *int    `json:"position,omitempty"`
}
