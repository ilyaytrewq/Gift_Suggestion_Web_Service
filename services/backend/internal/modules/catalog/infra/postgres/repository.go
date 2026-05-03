package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	catalogusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/usecase"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListGifts(ctx context.Context, filter catalogusecase.GiftFilter) (gifts []catalogdomain.Gift, total int, err error) {
	whereSQL, args := buildGiftWhere(filter)

	countQuery := `
		SELECT COUNT(*)
		FROM gifts g
		LEFT JOIN categories c ON c.id = g.category_id
	` + whereSQL

	if err = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.Limit, filter.Offset)

	//nolint:gosec // ORDER BY fragments and placeholder indexes come only from internal enum mappings.
	query := `
		SELECT
			g.id,
			g.category_id,
			c.name,
			g.name,
			g.description,
			g.price::text,
			g.store_link,
			g.image,
			g.age_restriction,
			g.created_at,
			g.updated_at
		FROM gifts g
		LEFT JOIN categories c ON c.id = g.category_id
	` + whereSQL + `
		ORDER BY ` + giftOrderBy(filter.Sort) + `
		LIMIT $` + fmt.Sprintf("%d", len(queryArgs)-1) + `
		OFFSET $` + fmt.Sprintf("%d", len(queryArgs))

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	gifts = make([]catalogdomain.Gift, 0, filter.Limit)
	for rows.Next() {
		var gift catalogdomain.Gift
		gift, err = scanGift(rows)
		if err != nil {
			return nil, 0, err
		}

		gifts = append(gifts, gift)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return gifts, total, nil
}

func (r *Repository) GetGift(ctx context.Context, id catalogdomain.GiftID) (*catalogdomain.Gift, error) {
	const query = `
		SELECT
			g.id,
			g.category_id,
			c.name,
			g.name,
			g.description,
			g.price::text,
			g.store_link,
			g.image,
			g.age_restriction,
			g.created_at,
			g.updated_at
		FROM gifts g
		LEFT JOIN categories c ON c.id = g.category_id
		WHERE g.id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id.String())
	gift, err := scanGift(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &gift, nil
}

func (r *Repository) ListCategories(ctx context.Context, filter catalogusecase.CategoryFilter) (categories []catalogdomain.Category, total int, err error) {
	whereSQL, args := buildCategoryWhere(filter)

	countQuery := `SELECT COUNT(*) FROM categories` + whereSQL
	if err = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, filter.Limit, filter.Offset)

	//nolint:gosec // ORDER BY fragments and placeholder indexes come only from internal enum mappings.
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
	` + whereSQL + `
		ORDER BY ` + categoryOrderBy(filter.Sort) + `
		LIMIT $` + fmt.Sprintf("%d", len(queryArgs)-1) + `
		OFFSET $` + fmt.Sprintf("%d", len(queryArgs))

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	categories = make([]catalogdomain.Category, 0, filter.Limit)
	for rows.Next() {
		var (
			id        string
			name      string
			createdAt sql.NullTime
			updatedAt sql.NullTime
		)

		if err = rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}

		var category catalogdomain.Category
		category, err = catalogdomain.RestoreCategory(id, name, createdAt.Time, updatedAt.Time)
		if err != nil {
			return nil, 0, err
		}

		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *Repository) ListSimilarGifts(ctx context.Context, giftID catalogdomain.GiftID, categoryID *catalogdomain.CategoryID, priceCents int64, limit int) ([]catalogdomain.Gift, error) {
	// Same category + price within ±25%, excluding the source gift, ordered by closeness.
	low := priceCents * 75 / 100
	high := priceCents * 125 / 100

	args := []any{giftID.String(), low, high, limit}
	var whereCategory string
	if categoryID != nil {
		args = append([]any{giftID.String(), categoryID.String(), low, high, limit}, args[4:]...)
		args = []any{giftID.String(), categoryID.String(), low, high, limit}
		whereCategory = "AND g.category_id = $2"
	}

	var query string
	if categoryID != nil {
		query = `
			SELECT g.id, g.category_id, c.name, g.name, g.description, g.price::text,
			       g.store_link, g.image, g.age_restriction, g.created_at, g.updated_at
			FROM gifts g
			LEFT JOIN categories c ON c.id = g.category_id
			WHERE g.id <> $1 ` + whereCategory + `
			  AND g.price BETWEEN $3 AND $4
			ORDER BY ABS(g.price - $3) ASC, g.created_at DESC
			LIMIT $5
		`
	} else {
		query = `
			SELECT g.id, g.category_id, c.name, g.name, g.description, g.price::text,
			       g.store_link, g.image, g.age_restriction, g.created_at, g.updated_at
			FROM gifts g
			LEFT JOIN categories c ON c.id = g.category_id
			WHERE g.id <> $1
			  AND g.price BETWEEN $2 AND $3
			ORDER BY ABS(g.price - $2) ASC, g.created_at DESC
			LIMIT $4
		`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var gifts []catalogdomain.Gift
	for rows.Next() {
		gift, err := scanGift(rows)
		if err != nil {
			return nil, err
		}
		gifts = append(gifts, gift)
	}

	return gifts, rows.Err()
}

func (r *Repository) ListOffersByGiftID(ctx context.Context, giftID catalogdomain.GiftID) ([]catalogdomain.Offer, error) {
	const query = `
		SELECT id, store_name, store_url, price_cents, currency, available, created_at
		FROM gift_offers
		WHERE gift_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, giftID.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var offers []catalogdomain.Offer
	for rows.Next() {
		var (
			id         string
			storeName  string
			storeURL   string
			priceCents int64
			currency   string
			available  bool
			createdAt  sql.NullTime
		)

		if err = rows.Scan(&id, &storeName, &storeURL, &priceCents, &currency, &available, &createdAt); err != nil {
			return nil, err
		}

		offer, err := catalogdomain.RestoreOffer(id, giftID, storeName, storeURL, priceCents, currency, available, createdAt.Time)
		if err != nil {
			return nil, err
		}

		offers = append(offers, offer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGift(scanner rowScanner) (catalogdomain.Gift, error) {
	var (
		id             string
		categoryID     sql.NullString
		categoryName   sql.NullString
		name           string
		description    string
		price          string
		storeLink      string
		image          sql.NullString
		ageRestriction sql.NullInt16
		createdAt      sql.NullTime
		updatedAt      sql.NullTime
	)

	err := scanner.Scan(
		&id,
		&categoryID,
		&categoryName,
		&name,
		&description,
		&price,
		&storeLink,
		&image,
		&ageRestriction,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return catalogdomain.Gift{}, err
	}

	var categoryIDPtr *string
	if categoryID.Valid {
		value := categoryID.String
		categoryIDPtr = &value
	}

	var categoryNamePtr *string
	if categoryName.Valid {
		value := categoryName.String
		categoryNamePtr = &value
	}

	var imagePtr *string
	if image.Valid {
		value := image.String
		imagePtr = &value
	}

	var ageRestrictionPtr *int
	if ageRestriction.Valid {
		value := int(ageRestriction.Int16)
		ageRestrictionPtr = &value
	}

	return catalogdomain.RestoreGift(
		id,
		categoryIDPtr,
		categoryNamePtr,
		name,
		description,
		price,
		storeLink,
		imagePtr,
		ageRestrictionPtr,
		createdAt.Time,
		updatedAt.Time,
	)
}

func buildGiftWhere(filter catalogusecase.GiftFilter) (string, []any) {
	clauses := make([]string, 0, 6)
	args := make([]any, 0, 6)

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		position := len(args)
		clauses = append(clauses, fmt.Sprintf("(g.name ILIKE $%d OR g.description ILIKE $%d)", position, position))
	}
	if filter.CategoryID != nil {
		args = append(args, filter.CategoryID.String())
		clauses = append(clauses, fmt.Sprintf("g.category_id = $%d", len(args)))
	}
	if filter.MinPrice != nil {
		args = append(args, filter.MinPrice.DecimalString())
		clauses = append(clauses, fmt.Sprintf("g.price >= $%d", len(args)))
	}
	if filter.MaxPrice != nil {
		args = append(args, filter.MaxPrice.DecimalString())
		clauses = append(clauses, fmt.Sprintf("g.price <= $%d", len(args)))
	}
	if filter.AgeRestriction != nil {
		args = append(args, filter.AgeRestriction.Int())
		clauses = append(clauses, fmt.Sprintf("g.age_restriction = $%d", len(args)))
	}
	if filter.HasImage != nil {
		if *filter.HasImage {
			clauses = append(clauses, "g.image IS NOT NULL AND g.image <> ''")
		} else {
			clauses = append(clauses, "(g.image IS NULL OR g.image = '')")
		}
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func buildCategoryWhere(filter catalogusecase.CategoryFilter) (string, []any) {
	var clauses []string
	var args []any

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		clauses = append(clauses, fmt.Sprintf("name ILIKE $%d", len(args)))
	}
	if filter.HasGifts != nil && *filter.HasGifts {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM gifts g WHERE g.category_id = categories.id)")
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func giftOrderBy(sort catalogusecase.GiftSort) string {
	switch sort {
	case catalogusecase.GiftSortNameAsc:
		return "g.name ASC, g.id ASC"
	case catalogusecase.GiftSortNameDesc:
		return "g.name DESC, g.id DESC"
	case catalogusecase.GiftSortPriceAsc:
		return "g.price ASC, g.id ASC"
	case catalogusecase.GiftSortPriceDesc:
		return "g.price DESC, g.id DESC"
	default:
		return "g.created_at DESC, g.id DESC"
	}
}

func categoryOrderBy(sort catalogusecase.CategorySort) string {
	switch sort {
	case catalogusecase.CategorySortNameDesc:
		return "name DESC, id DESC"
	case catalogusecase.CategorySortNewest:
		return "created_at DESC, id DESC"
	default:
		return "name ASC, id ASC"
	}
}
