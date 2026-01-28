package session

import (
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/services/backend/internal/domain/snapshot"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/services/backend/internal/domain/user"
)

type SessionID string

type Session struct {
	ID          SessionID
	OwnerUserID *user.UserID
	SnapShot    snapshot.Snapshot
	CreatedAt   time.Time
}

func NewSession(id SessionID, ownerUserID *user.UserID, snap snapshot.Snapshot) (*Session, error) {
	if isBlank(string(id)) {
		return nil, ErrSessionIDEmpty
	}
	if !id.IsValid() {
		return nil, ErrInvalidSessionID
	}
	if ownerUserID != nil && !ownerUserID.IsValid() {
		return nil, ErrInvalidOwnerUserID
	}
	normalizedSnapshot, err := normalizeSnapshot(snap)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:          id,
		OwnerUserID: ownerUserID,
		SnapShot:    normalizedSnapshot,
		CreatedAt:   time.Now(),
	}, nil
}

func normalizeSnapshot(snap snapshot.Snapshot) (snapshot.Snapshot, error) {
	normalized, err := snapshot.NewSnapshot(snap.Occasion, snap.Relation, snap.Budget, snap.InterestTags, snap.Age)
	if err != nil {
		return snapshot.Snapshot{}, ErrInvalidSnapshot
	}
	return *normalized, nil
}
