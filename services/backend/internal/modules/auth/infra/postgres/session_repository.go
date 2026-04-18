package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Save(ctx context.Context, session *authdomain.Session) error {
	const query = `
		INSERT INTO auth_sessions (
			id, user_id, refresh_token_hash, expires_at, created_at, updated_at, last_used_at, revoked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID().String(),
		session.UserID().String(),
		session.RefreshTokenHash(),
		session.ExpiresAt(),
		session.CreatedAt(),
		session.UpdatedAt(),
		nullTime(session.LastUsedAt()),
		nullTime(session.RevokedAt()),
	)

	return err
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, tokenHash string) (*authdomain.Session, error) {
	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, created_at, updated_at, last_used_at, revoked_at
		FROM auth_sessions
		WHERE refresh_token_hash = $1
	`

	var (
		id               string
		userID           string
		refreshTokenHash string
		expiresAt        time.Time
		createdAt        time.Time
		updatedAt        time.Time
		lastUsedAt       sql.NullTime
		revokedAt        sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&id,
		&userID,
		&refreshTokenHash,
		&expiresAt,
		&createdAt,
		&updatedAt,
		&lastUsedAt,
		&revokedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, authdomain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	session, err := authdomain.RestoreSession(
		id,
		userID,
		refreshTokenHash,
		expiresAt,
		createdAt,
		updatedAt,
		nullTimePtr(lastUsedAt),
		nullTimePtr(revokedAt),
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) Update(ctx context.Context, session *authdomain.Session) error {
	const query = `
		UPDATE auth_sessions
		SET refresh_token_hash = $2, expires_at = $3, updated_at = $4, last_used_at = $5, revoked_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.ID().String(),
		session.RefreshTokenHash(),
		session.ExpiresAt(),
		session.UpdatedAt(),
		nullTime(session.LastUsedAt()),
		nullTime(session.RevokedAt()),
	)

	return err
}
