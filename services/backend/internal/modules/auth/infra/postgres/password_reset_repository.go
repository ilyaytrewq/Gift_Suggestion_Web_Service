package postgres

import (
	"context"
	"database/sql"

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
