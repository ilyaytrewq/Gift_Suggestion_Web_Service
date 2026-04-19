package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	trackingdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/domain"
)

type Repository struct {
	db *sql.DB
}

type metadataSnapshot struct {
	Surface  *string `json:"surface,omitempty"`
	Position *int    `json:"position,omitempty"`
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateEvent(
	ctx context.Context,
	event *trackingdomain.Event,
) (trackingdomain.Event, bool, error) {
	metadataJSON, err := marshalMetadata(event.Metadata())
	if err != nil {
		return trackingdomain.Event{}, false, err
	}

	const insertQuery = `
		INSERT INTO tracking_events (
			id,
			event_type,
			user_id,
			recommendation_request_id,
			wishlist_id,
			gift_id,
			client_event_id,
			metadata,
			occurred_at,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, client_event_id) WHERE client_event_id IS NOT NULL DO NOTHING
		RETURNING
			id,
			event_type,
			user_id,
			recommendation_request_id,
			wishlist_id,
			gift_id,
			client_event_id,
			metadata,
			occurred_at,
			created_at
	`

	var stored trackingdomain.Event
	row := r.db.QueryRowContext(
		ctx,
		insertQuery,
		event.ID().String(),
		string(event.EventType()),
		event.UserID().String(),
		requestIDValue(event.RecommendationRequestID()),
		wishlistIDValue(event.WishlistID()),
		giftIDValue(event.GiftID()),
		nullString(event.ClientEventID()),
		metadataJSON,
		event.OccurredAt(),
		event.CreatedAt(),
	)

	stored, err = scanEvent(row)
	if err == nil {
		return stored, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return trackingdomain.Event{}, false, err
	}

	clientEventID := event.ClientEventID()
	if clientEventID == nil {
		return trackingdomain.Event{}, false, sql.ErrNoRows
	}

	const selectExistingQuery = `
		SELECT
			id,
			event_type,
			user_id,
			recommendation_request_id,
			wishlist_id,
			gift_id,
			client_event_id,
			metadata,
			occurred_at,
			created_at
		FROM tracking_events
		WHERE user_id = $1 AND client_event_id = $2
	`

	stored, err = scanEvent(r.db.QueryRowContext(ctx, selectExistingQuery, event.UserID().String(), *clientEventID))
	if err != nil {
		return trackingdomain.Event{}, false, err
	}

	return stored, false, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEvent(scanner rowScanner) (trackingdomain.Event, error) {
	var (
		id                      string
		eventType               string
		userID                  string
		recommendationRequestID sql.NullString
		wishlistID              sql.NullString
		giftID                  sql.NullString
		clientEventID           sql.NullString
		metadataRaw             []byte
		occurredAt              sql.NullTime
		createdAt               sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&eventType,
		&userID,
		&recommendationRequestID,
		&wishlistID,
		&giftID,
		&clientEventID,
		&metadataRaw,
		&occurredAt,
		&createdAt,
	); err != nil {
		return trackingdomain.Event{}, err
	}

	metadata, err := unmarshalMetadata(metadataRaw)
	if err != nil {
		return trackingdomain.Event{}, err
	}

	return trackingdomain.RestoreEvent(
		id,
		eventType,
		userID,
		stringPtrFromNull(recommendationRequestID),
		stringPtrFromNull(wishlistID),
		stringPtrFromNull(giftID),
		stringPtrFromNull(clientEventID),
		metadata,
		occurredAt.Time,
		createdAt.Time,
	)
}

func marshalMetadata(metadata trackingdomain.EventMetadata) ([]byte, error) {
	return json.Marshal(metadataSnapshot{
		Surface:  metadata.Surface(),
		Position: metadata.Position(),
	})
}

func unmarshalMetadata(payload []byte) (trackingdomain.EventMetadata, error) {
	if len(payload) == 0 {
		return trackingdomain.NewEventMetadata(nil, nil)
	}

	var snapshot metadataSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return trackingdomain.EventMetadata{}, err
	}

	return trackingdomain.NewEventMetadata(snapshot.Surface, snapshot.Position)
}

func nullString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	current := value.String
	return &current
}

func requestIDValue(value interface{ String() string }) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func wishlistIDValue(value interface{ String() string }) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func giftIDValue(value interface{ String() string }) any {
	if value == nil {
		return nil
	}

	return value.String()
}
