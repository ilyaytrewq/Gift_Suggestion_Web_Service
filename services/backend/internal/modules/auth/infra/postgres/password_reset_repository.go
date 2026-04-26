package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
)

type PasswordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Save(ctx context.Context, token *authdomain.PasswordResetToken) error {
	const query = `
		INSERT INTO password_reset_tokens (
			id, user_id, token_hash, expires_at, created_at, used_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID().String(),
		token.UserID().String(),
		token.TokenHash(),
		token.ExpiresAt(),
		token.CreatedAt(),
		nullTime(token.UsedAt()),
	)

	return err
}

func (r *PasswordResetRepository) Consume(
	ctx context.Context,
	tokenHash string,
	newPasswordHash string,
	now time.Time,
) error {
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
		FROM password_reset_tokens
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
		return authdomain.ErrPasswordResetTokenNotFound
	} else if err != nil {
		return err
	}

	token, err := authdomain.RestorePasswordResetToken(
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
		return authdomain.ErrPasswordResetTokenUsed
	}
	if token.IsExpired(now) {
		return authdomain.ErrPasswordResetTokenExpired
	}

	token.MarkUsed(now)

	const updateUser = `
		UPDATE users
		SET password_hash = $2, password_changed_at = $3, updated_at = $3
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateUser, token.UserID().String(), newPasswordHash, now.UTC()); err != nil {
		return err
	}

	//nolint:gosec // token_hash is a domain term, not a credential.
	const updateToken = `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateToken, token.ID().String(), nullTime(token.UsedAt())); err != nil {
		return err
	}

	return tx.Commit()
}
