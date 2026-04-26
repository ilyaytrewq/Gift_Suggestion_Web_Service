package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
)

type EmailVerificationRepository struct {
	db *sql.DB
}

func NewEmailVerificationRepository(db *sql.DB) *EmailVerificationRepository {
	return &EmailVerificationRepository{db: db}
}

func (r *EmailVerificationRepository) Consume(ctx context.Context, tokenHash string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			_ = rollbackErr
		}
	}()

	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at, used_at
		FROM email_verification_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`

	var (
		id        string
		userID    string
		hash      string
		expiresAt time.Time
		createdAt time.Time
		usedAt    sql.NullTime
	)

	if err := tx.QueryRowContext(ctx, query, tokenHash).Scan(
		&id,
		&userID,
		&hash,
		&expiresAt,
		&createdAt,
		&usedAt,
	); errors.Is(err, sql.ErrNoRows) {
		return authdomain.ErrEmailVerificationTokenNotFound
	} else if err != nil {
		return err
	}

	token, err := authdomain.RestoreEmailVerificationToken(
		id,
		userID,
		hash,
		expiresAt,
		createdAt,
		nullTimePtr(usedAt),
	)
	if err != nil {
		return err
	}
	if token.IsUsed() {
		return authdomain.ErrEmailVerificationTokenUsed
	}
	if token.IsExpired(now) {
		return authdomain.ErrEmailVerificationTokenExpired
	}

	token.MarkUsed(now)

	const updateUser = `
		UPDATE users
		SET email_verified_at = COALESCE(email_verified_at, $2), updated_at = $2
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateUser, token.UserID().String(), now.UTC()); err != nil {
		return err
	}

	//nolint:gosec // token_hash is a domain term, not a credential.
	const updateToken = `
		UPDATE email_verification_tokens
		SET used_at = $2
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateToken, token.ID().String(), nullTime(token.UsedAt())); err != nil {
		return err
	}

	return tx.Commit()
}
