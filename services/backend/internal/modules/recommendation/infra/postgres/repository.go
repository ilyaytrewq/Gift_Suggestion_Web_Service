package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type Repository struct {
	db *sql.DB
}

type questionnaireSnapshot struct {
	Occasion             string   `json:"occasion,omitempty"`
	Relationship         string   `json:"relationship,omitempty"`
	RecipientAge         *int     `json:"recipient_age,omitempty"`
	BudgetMax            string   `json:"budget_max"`
	PreferredCategoryIDs []string `json:"preferred_category_ids,omitempty"`
	Interests            []string `json:"interests,omitempty"`
	TopN                 int      `json:"top_n"`
	UseWishlistContext   bool     `json:"use_wishlist_context"`
}

type hardFiltersSnapshot struct {
	BudgetMax            string   `json:"budget_max"`
	RecipientAge         *int     `json:"recipient_age,omitempty"`
	PreferredCategoryIDs []string `json:"preferred_category_ids,omitempty"`
	UseWishlistContext   bool     `json:"use_wishlist_context"`
}

type explanationSnapshot struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SelectCandidates(
	ctx context.Context,
	selection recommendationusecase.CandidateSelection,
) (candidates []catalogdomain.Gift, total int, err error) {
	whereSQL, args := buildCandidateWhere(selection)

	countQuery := `
		SELECT COUNT(*)
		FROM gifts g
	` + whereSQL
	if err = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, selection.Limit)

	//nolint:gosec // WHERE fragments and placeholder indexes are built only from validated internal filters.
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
		ORDER BY g.created_at DESC, g.id DESC
		LIMIT $` + fmt.Sprintf("%d", len(queryArgs))

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

	candidates = make([]catalogdomain.Gift, 0, selection.Limit)
	for rows.Next() {
		var gift catalogdomain.Gift
		gift, err = scanGift(rows)
		if err != nil {
			return nil, 0, err
		}

		candidates = append(candidates, gift)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return candidates, total, nil
}

func (r *Repository) CreateRequest(ctx context.Context, request *recommendationdomain.RecommendationRequest) error {
	questionnaireJSON, err := marshalQuestionnaire(request.Questionnaire())
	if err != nil {
		return err
	}

	hardFiltersJSON, err := marshalHardFilters(request.Questionnaire())
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO recommendation_requests (
			id,
			requested_by_user_id,
			status,
			ranking_source,
			criteria_version,
			questionnaire,
			hard_filters,
			requested_top_n,
			candidate_count_before_filters,
			candidate_count_after_filters,
			returned_primary_count,
			returned_alternative_count,
			fallback_reason_code,
			failure_code,
			failure_message,
			started_at,
			finished_at,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, 'v1', $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		request.ID().String(),
		userIDValue(request.RequestedByUserID()),
		string(request.Status()),
		string(request.RankingSource()),
		questionnaireJSON,
		hardFiltersJSON,
		request.Questionnaire().TopN(),
		request.CandidateCountBeforeFilters(),
		request.CandidateCountAfterFilters(),
		request.ReturnedPrimaryCount(),
		request.ReturnedAlternativeCount(),
		nullStringPtr(request.FallbackReasonCode()),
		nullStringPtr(request.FailureCode()),
		nullStringPtr(request.FailureMessage()),
		request.StartedAt(),
		nullTime(request.FinishedAt()),
		request.CreatedAt(),
		request.UpdatedAt(),
	)
	return err
}

func (r *Repository) CompleteRequest(
	ctx context.Context,
	request *recommendationdomain.RecommendationRequest,
	results []recommendationdomain.RecommendationResult,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	const updateQuery = `
		UPDATE recommendation_requests
		SET
			status = $2,
			ranking_source = $3,
			candidate_count_before_filters = $4,
			candidate_count_after_filters = $5,
			returned_primary_count = $6,
			returned_alternative_count = $7,
			fallback_reason_code = $8,
			failure_code = $9,
			failure_message = $10,
			finished_at = $11,
			updated_at = $12
		WHERE id = $1
	`

	result, err := tx.ExecContext(
		ctx,
		updateQuery,
		request.ID().String(),
		string(request.Status()),
		string(request.RankingSource()),
		request.CandidateCountBeforeFilters(),
		request.CandidateCountAfterFilters(),
		request.ReturnedPrimaryCount(),
		request.ReturnedAlternativeCount(),
		nullStringPtr(request.FallbackReasonCode()),
		nullStringPtr(request.FailureCode()),
		nullStringPtr(request.FailureMessage()),
		nullTime(request.FinishedAt()),
		request.UpdatedAt(),
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM recommendation_results WHERE request_id = $1`, request.ID().String()); err != nil {
		return err
	}

	const insertQuery = `
		INSERT INTO recommendation_results (
			id,
			request_id,
			gift_id,
			slot_position,
			result_kind,
			alternative_rank,
			ranking_source,
			score,
			explanations,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, item := range results {
		explanationsJSON, err := marshalExplanations(item.Explanations())
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(
			ctx,
			insertQuery,
			item.ID().String(),
			item.RequestID().String(),
			item.GiftID().String(),
			item.SlotPosition(),
			string(item.ResultKind()),
			nullInt(item.AlternativeRank()),
			string(item.RankingSource()),
			nullFloat64(item.Score()),
			explanationsJSON,
			item.CreatedAt(),
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (r *Repository) FailRequest(ctx context.Context, request *recommendationdomain.RecommendationRequest) error {
	const query = `
		UPDATE recommendation_requests
		SET
			status = $2,
			ranking_source = $3,
			failure_code = $4,
			failure_message = $5,
			finished_at = $6,
			updated_at = $7
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		request.ID().String(),
		string(request.Status()),
		string(request.RankingSource()),
		nullStringPtr(request.FailureCode()),
		nullStringPtr(request.FailureMessage()),
		nullTime(request.FinishedAt()),
		request.UpdatedAt(),
	)
	return err
}

func (r *Repository) GetRequest(
	ctx context.Context,
	id recommendationdomain.RequestID,
) (*recommendationdomain.RecommendationRequest, error) {
	const query = `
		SELECT
			id,
			requested_by_user_id,
			status,
			ranking_source,
			questionnaire,
			candidate_count_before_filters,
			candidate_count_after_filters,
			returned_primary_count,
			returned_alternative_count,
			fallback_reason_code,
			failure_code,
			failure_message,
			started_at,
			finished_at,
			created_at,
			updated_at
		FROM recommendation_requests
		WHERE id = $1
	`

	var (
		requestID                   string
		requestedByUserID           sql.NullString
		status                      string
		rankingSource               string
		questionnaireJSON           []byte
		candidateCountBeforeFilters int
		candidateCountAfterFilters  int
		returnedPrimaryCount        int
		returnedAlternativeCount    int
		fallbackReasonCode          sql.NullString
		failureCode                 sql.NullString
		failureMessage              sql.NullString
		startedAt                   time.Time
		finishedAt                  sql.NullTime
		createdAt                   time.Time
		updatedAt                   time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&requestID,
		&requestedByUserID,
		&status,
		&rankingSource,
		&questionnaireJSON,
		&candidateCountBeforeFilters,
		&candidateCountAfterFilters,
		&returnedPrimaryCount,
		&returnedAlternativeCount,
		&fallbackReasonCode,
		&failureCode,
		&failureMessage,
		&startedAt,
		&finishedAt,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	questionnaire, err := unmarshalQuestionnaire(questionnaireJSON)
	if err != nil {
		return nil, err
	}

	request, err := recommendationdomain.RestoreRecommendationRequest(
		requestID,
		stringPtrFromNull(requestedByUserID),
		questionnaire,
		status,
		rankingSource,
		candidateCountBeforeFilters,
		candidateCountAfterFilters,
		returnedPrimaryCount,
		returnedAlternativeCount,
		stringPtrFromNull(fallbackReasonCode),
		stringPtrFromNull(failureCode),
		stringPtrFromNull(failureMessage),
		startedAt,
		timePtrFromNull(finishedAt),
		createdAt,
		updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *Repository) ListResults(
	ctx context.Context,
	requestID recommendationdomain.RequestID,
) (results []recommendationdomain.RecommendationResult, err error) {
	const query = `
		SELECT
			id,
			request_id,
			gift_id,
			slot_position,
			result_kind,
			alternative_rank,
			ranking_source,
			score,
			explanations,
			created_at
		FROM recommendation_results
		WHERE request_id = $1
		ORDER BY slot_position ASC,
			CASE WHEN result_kind = 'primary' THEN 0 ELSE 1 END ASC,
			alternative_rank ASC NULLS FIRST,
			id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, requestID.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	results = make([]recommendationdomain.RecommendationResult, 0)
	for rows.Next() {
		var (
			id              string
			requestIDRaw    string
			giftID          string
			slotPosition    int
			resultKind      string
			alternativeRank sql.NullInt64
			rankingSource   string
			score           sql.NullFloat64
			explanationsRaw []byte
			createdAt       time.Time
		)

		if err = rows.Scan(
			&id,
			&requestIDRaw,
			&giftID,
			&slotPosition,
			&resultKind,
			&alternativeRank,
			&rankingSource,
			&score,
			&explanationsRaw,
			&createdAt,
		); err != nil {
			return nil, err
		}

		explanations, err := unmarshalExplanations(explanationsRaw)
		if err != nil {
			return nil, err
		}

		item, err := recommendationdomain.RestoreRecommendationResult(
			id,
			requestIDRaw,
			giftID,
			slotPosition,
			resultKind,
			intPtrFromNullInt64(alternativeRank),
			rankingSource,
			float64PtrFromNull(score),
			explanations,
			createdAt,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
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

	if err := scanner.Scan(
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
	); err != nil {
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

func buildCandidateWhere(selection recommendationusecase.CandidateSelection) (string, []any) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3+len(selection.PreferredCategoryIDs))

	args = append(args, selection.BudgetMax.DecimalString())
	clauses = append(clauses, fmt.Sprintf("g.price <= $%d", len(args)))

	if selection.RecipientAge != nil {
		args = append(args, *selection.RecipientAge)
		clauses = append(clauses, fmt.Sprintf("(g.age_restriction IS NULL OR g.age_restriction <= $%d)", len(args)))
	}

	if len(selection.PreferredCategoryIDs) > 0 {
		categoryClauses := make([]string, 0, len(selection.PreferredCategoryIDs))
		for _, categoryID := range selection.PreferredCategoryIDs {
			args = append(args, categoryID.String())
			categoryClauses = append(categoryClauses, fmt.Sprintf("$%d", len(args)))
		}

		clauses = append(clauses, "g.category_id IN ("+strings.Join(categoryClauses, ", ")+")")
	}

	if len(clauses) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}

func marshalQuestionnaire(questionnaire recommendationdomain.Questionnaire) ([]byte, error) {
	payload := questionnaireSnapshot{
		Occasion:             questionnaire.Occasion(),
		Relationship:         questionnaire.Relationship(),
		RecipientAge:         questionnaire.RecipientAge(),
		BudgetMax:            questionnaire.BudgetMax().DecimalString(),
		PreferredCategoryIDs: categoryIDsToStrings(questionnaire.PreferredCategoryIDs()),
		Interests:            questionnaire.Interests(),
		TopN:                 questionnaire.TopN(),
		UseWishlistContext:   questionnaire.UseWishlistContext(),
	}

	return json.Marshal(payload)
}

func marshalHardFilters(questionnaire recommendationdomain.Questionnaire) ([]byte, error) {
	payload := hardFiltersSnapshot{
		BudgetMax:            questionnaire.BudgetMax().DecimalString(),
		RecipientAge:         questionnaire.RecipientAge(),
		PreferredCategoryIDs: categoryIDsToStrings(questionnaire.PreferredCategoryIDs()),
		UseWishlistContext:   questionnaire.UseWishlistContext(),
	}

	return json.Marshal(payload)
}

func marshalExplanations(explanations []recommendationdomain.Explanation) ([]byte, error) {
	payload := make([]explanationSnapshot, 0, len(explanations))
	for _, explanation := range explanations {
		payload = append(payload, explanationSnapshot{
			Code: explanation.Code(),
			Text: explanation.Text(),
		})
	}

	return json.Marshal(payload)
}

func unmarshalQuestionnaire(payload []byte) (recommendationdomain.Questionnaire, error) {
	var snapshot questionnaireSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return recommendationdomain.Questionnaire{}, err
	}

	return recommendationdomain.NewQuestionnaire(
		snapshot.Occasion,
		snapshot.Relationship,
		snapshot.RecipientAge,
		snapshot.BudgetMax,
		snapshot.PreferredCategoryIDs,
		snapshot.Interests,
		snapshot.TopN,
		snapshot.UseWishlistContext,
	)
}

func unmarshalExplanations(payload []byte) ([]recommendationdomain.Explanation, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var snapshot []explanationSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return nil, err
	}

	result := make([]recommendationdomain.Explanation, 0, len(snapshot))
	for _, item := range snapshot {
		explanation, err := recommendationdomain.NewExplanation(item.Code, item.Text)
		if err != nil {
			return nil, err
		}

		result = append(result, explanation)
	}

	return result, nil
}

func userIDValue(value *userdomain.UserID) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func nullStringPtr(value *string) sql.NullString {
	if value == nil || *value == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
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

func nullInt(value *int) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{
		Int64: int64(*value),
		Valid: true,
	}
}

func nullFloat64(value *float64) sql.NullFloat64 {
	if value == nil {
		return sql.NullFloat64{}
	}

	return sql.NullFloat64{
		Float64: *value,
		Valid:   true,
	}
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	result := value.String
	return &result
}

func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time.UTC()
	return &result
}

func intPtrFromNullInt64(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}

	result := int(value.Int64)
	return &result
}

func float64PtrFromNull(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	result := value.Float64
	return &result
}

func categoryIDsToStrings(ids []catalogdomain.CategoryID) []string {
	result := make([]string, len(ids))
	for index, id := range ids {
		result[index] = id.String()
	}

	return result
}
