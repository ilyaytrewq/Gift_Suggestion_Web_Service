package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
	wishlistusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/usecase"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWishlist(ctx context.Context, wishlist *wishlistdomain.Wishlist) error {
	const query = `
		INSERT INTO wishlists (id, user_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		wishlist.ID().String(),
		wishlist.UserID().String(),
		wishlist.Name(),
		wishlist.CreatedAt(),
		wishlist.UpdatedAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return wishlistdomain.ErrWishlistAlreadyExists
		}

		return err
	}

	return nil
}

func (r *Repository) GetWishlistByID(ctx context.Context, id wishlistdomain.WishlistID) (*wishlistdomain.Wishlist, error) {
	const query = `
		SELECT id, user_id, name, created_at, updated_at
		FROM wishlists
		WHERE id = $1
	`

	var (
		wishlistID string
		userID     string
		name       string
		createdAt  sql.NullTime
		updatedAt  sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&wishlistID,
		&userID,
		&name,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	wishlist, err := wishlistdomain.RestoreWishlist(wishlistID, userID, name, createdAt.Time, updatedAt.Time)
	if err != nil {
		return nil, err
	}

	return &wishlist, nil
}

func (r *Repository) ListWishlistsByUser(
	ctx context.Context,
	userID userdomain.UserID,
	limit int,
	offset int,
) (records []wishlistusecase.WishlistSummaryRecord, total int, err error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM wishlists
		WHERE user_id = $1
	`

	if err = r.db.QueryRowContext(ctx, countQuery, userID.String()).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT
			w.id,
			w.user_id,
			w.name,
			w.created_at,
			w.updated_at,
			COUNT(wi.id) AS item_count
		FROM wishlists w
		LEFT JOIN wishlist_items wi ON wi.wishlist_id = w.id
		WHERE w.user_id = $1
		GROUP BY w.id, w.user_id, w.name, w.created_at, w.updated_at
		ORDER BY w.created_at DESC, w.id DESC
		LIMIT $2
		OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	records = make([]wishlistusecase.WishlistSummaryRecord, 0, limit)
	for rows.Next() {
		var (
			wishlistID string
			ownerID    string
			name       string
			createdAt  sql.NullTime
			updatedAt  sql.NullTime
			itemCount  int
		)

		if err = rows.Scan(&wishlistID, &ownerID, &name, &createdAt, &updatedAt, &itemCount); err != nil {
			return nil, 0, err
		}

		wishlist, err := wishlistdomain.RestoreWishlist(wishlistID, ownerID, name, createdAt.Time, updatedAt.Time)
		if err != nil {
			return nil, 0, err
		}

		records = append(records, wishlistusecase.WishlistSummaryRecord{
			Wishlist:  wishlist,
			ItemCount: itemCount,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *Repository) ListWishlistItems(
	ctx context.Context,
	wishlistID wishlistdomain.WishlistID,
) (items []wishlistdomain.WishlistItem, err error) {
	const query = `
		SELECT id, wishlist_id, gift_id, created_at
		FROM wishlist_items
		WHERE wishlist_id = $1
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, wishlistID.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	items = make([]wishlistdomain.WishlistItem, 0)
	for rows.Next() {
		var (
			itemID    string
			listID    string
			giftID    string
			createdAt sql.NullTime
		)

		if err = rows.Scan(&itemID, &listID, &giftID, &createdAt); err != nil {
			return nil, err
		}

		item, err := wishlistdomain.RestoreWishlistItem(itemID, listID, giftID, createdAt.Time)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) AddWishlistItem(ctx context.Context, item *wishlistdomain.WishlistItem) error {
	const query = `
		INSERT INTO wishlist_items (id, wishlist_id, gift_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		item.ID().String(),
		item.WishlistID().String(),
		item.GiftID().String(),
		item.CreatedAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return wishlistdomain.ErrWishlistItemExists
		}

		return err
	}

	return nil
}

func (r *Repository) RemoveWishlistItem(
	ctx context.Context,
	wishlistID wishlistdomain.WishlistID,
	giftID catalogdomain.GiftID,
) error {
	const query = `
		DELETE FROM wishlist_items
		WHERE wishlist_id = $1 AND gift_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, wishlistID.String(), giftID.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return wishlistdomain.ErrWishlistItemNotFound
	}

	return nil
}

func (r *Repository) DeleteWishlist(ctx context.Context, id wishlistdomain.WishlistID) error {
	const query = `
		DELETE FROM wishlists
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return wishlistdomain.ErrWishlistNotFound
	}

	return nil
}
