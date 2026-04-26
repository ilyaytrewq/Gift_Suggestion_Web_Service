package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type RegistrationRepository struct {
	db *sql.DB
}

func NewRegistrationRepository(db *sql.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) SaveUserWithVerificationToken(
	ctx context.Context,
	user *userdomain.User,
	token *authdomain.EmailVerificationToken,
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

	const insertUser = `
		INSERT INTO users (
			id, email, password_hash, role, display_name, created_at, updated_at, last_login_at, password_changed_at, email_verified_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	if _, err := tx.ExecContext(
		ctx,
		insertUser,
		user.ID().String(),
		user.Email().String(),
		user.PasswordHash().String(),
		string(user.Role()),
		nullString(user.DisplayName()),
		user.CreatedAt(),
		user.UpdatedAt(),
		nullTime(user.LastLoginAt()),
		nullTime(user.PasswordChangedAt()),
		nullTime(user.EmailVerifiedAt()),
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return userdomain.ErrUserExists
		}

		return err
	}

	//nolint:gosec // token_hash is a domain term, not a credential.
	const insertToken = `
		INSERT INTO email_verification_tokens (
			id, user_id, token_hash, expires_at, created_at, used_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := tx.ExecContext(
		ctx,
		insertToken,
		token.ID().String(),
		token.UserID().String(),
		token.TokenHash(),
		token.ExpiresAt(),
		token.CreatedAt(),
		nullTime(token.UsedAt()),
	); err != nil {
		return err
	}

	return tx.Commit()
}
