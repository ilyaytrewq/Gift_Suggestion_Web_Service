package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, user *userdomain.User) error {
	const query = `
		INSERT INTO users (
			id, email, password_hash, role, display_name, created_at, updated_at, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID().String(),
		user.Email().String(),
		user.PasswordHash().String(),
		string(user.Role()),
		nullString(user.DisplayName()),
		user.CreatedAt(),
		user.UpdatedAt(),
		nullTime(user.LastLoginAt()),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return userdomain.ErrUserExists
		}

		return err
	}

	return nil
}

func (r *Repository) GetByID(ctx context.Context, id userdomain.UserID) (*userdomain.User, error) {
	const query = `
		SELECT id, email, password_hash, role, display_name, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	return r.getOne(ctx, query, id.String())
}

func (r *Repository) GetByEmail(ctx context.Context, email userdomain.Email) (*userdomain.User, error) {
	const query = `
		SELECT id, email, password_hash, role, display_name, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	return r.getOne(ctx, query, email.String())
}

func (r *Repository) UpdateProfile(ctx context.Context, user *userdomain.User) error {
	const query = `
		UPDATE users
		SET display_name = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.ID().String(),
		nullString(user.DisplayName()),
		user.UpdatedAt(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return userdomain.ErrUserNotFound
	}

	return nil
}

func (r *Repository) MarkLastLogin(ctx context.Context, id userdomain.UserID, at time.Time) error {
	const query = `
		UPDATE users
		SET last_login_at = $2, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String(), at.UTC())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return userdomain.ErrUserNotFound
	}

	return nil
}

func (r *Repository) getOne(ctx context.Context, query string, arg any) (*userdomain.User, error) {
	var (
		id           string
		email        string
		passwordHash string
		role         string
		displayName  sql.NullString
		createdAt    time.Time
		updatedAt    time.Time
		lastLoginAt  sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&id,
		&email,
		&passwordHash,
		&role,
		&displayName,
		&createdAt,
		&updatedAt,
		&lastLoginAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	user, err := userdomain.RestoreUser(
		id,
		email,
		passwordHash,
		role,
		displayName.String,
		createdAt,
		updatedAt,
		nullTimePtr(lastLoginAt),
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{
		String: value,
		Valid:  value != "",
	}
}

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{
		Time:  value.UTC(),
		Valid: true,
	}
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	timeValue := value.Time.UTC()
	return &timeValue
}
