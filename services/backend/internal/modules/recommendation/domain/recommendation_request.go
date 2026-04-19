package domain

import (
	"strings"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
)

type RequestStatus string

const (
	StatusPending               RequestStatus = "pending"
	StatusRunning               RequestStatus = "running"
	StatusCompleted             RequestStatus = "completed"
	StatusCompletedWithFallback RequestStatus = "completed_with_fallback"
	StatusCompletedEmpty        RequestStatus = "completed_empty"
	StatusFailed                RequestStatus = "failed"
)

type RankingSource string

const (
	RankingSourceNone     RankingSource = "none"
	RankingSourceML       RankingSource = "ml"
	RankingSourceFallback RankingSource = "fallback"
)

type RecommendationRequest struct {
	id                          RequestID
	requestedByUserID           *userdomain.UserID
	questionnaire               Questionnaire
	status                      RequestStatus
	rankingSource               RankingSource
	candidateCountBeforeFilters int
	candidateCountAfterFilters  int
	returnedPrimaryCount        int
	returnedAlternativeCount    int
	fallbackReasonCode          *string
	failureCode                 *string
	failureMessage              *string
	startedAt                   time.Time
	finishedAt                  *time.Time
	createdAt                   time.Time
	updatedAt                   time.Time
}

func NewRecommendationRequest(
	id RequestID,
	requestedByUserID *userdomain.UserID,
	questionnaire Questionnaire,
	now time.Time,
) RecommendationRequest {
	normalizedNow := now.UTC()
	return RecommendationRequest{
		id:                id,
		requestedByUserID: cloneUserID(requestedByUserID),
		questionnaire:     questionnaire,
		status:            StatusRunning,
		rankingSource:     RankingSourceNone,
		startedAt:         normalizedNow,
		createdAt:         normalizedNow,
		updatedAt:         normalizedNow,
	}
}

func RestoreRecommendationRequest(
	id string,
	requestedByUserID *string,
	questionnaire Questionnaire,
	status string,
	rankingSource string,
	candidateCountBeforeFilters int,
	candidateCountAfterFilters int,
	returnedPrimaryCount int,
	returnedAlternativeCount int,
	fallbackReasonCode *string,
	failureCode *string,
	failureMessage *string,
	startedAt time.Time,
	finishedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (RecommendationRequest, error) {
	requestID, err := NewRequestID(id)
	if err != nil {
		return RecommendationRequest{}, err
	}

	var requestedBy *userdomain.UserID
	if requestedByUserID != nil && strings.TrimSpace(*requestedByUserID) != "" {
		value, err := userdomain.NewUserID(strings.TrimSpace(*requestedByUserID))
		if err != nil {
			return RecommendationRequest{}, err
		}

		requestedBy = &value
	}

	requestStatus, err := parseRequestStatus(status)
	if err != nil {
		return RecommendationRequest{}, err
	}

	source, err := parseRankingSource(rankingSource)
	if err != nil {
		return RecommendationRequest{}, err
	}

	return RecommendationRequest{
		id:                          requestID,
		requestedByUserID:           requestedBy,
		questionnaire:               questionnaire,
		status:                      requestStatus,
		rankingSource:               source,
		candidateCountBeforeFilters: candidateCountBeforeFilters,
		candidateCountAfterFilters:  candidateCountAfterFilters,
		returnedPrimaryCount:        returnedPrimaryCount,
		returnedAlternativeCount:    returnedAlternativeCount,
		fallbackReasonCode:          cloneStringPtr(fallbackReasonCode),
		failureCode:                 cloneStringPtr(failureCode),
		failureMessage:              cloneStringPtr(failureMessage),
		startedAt:                   startedAt.UTC(),
		finishedAt:                  cloneTimePtr(finishedAt),
		createdAt:                   createdAt.UTC(),
		updatedAt:                   updatedAt.UTC(),
	}, nil
}

func (r *RecommendationRequest) Complete(
	source RankingSource,
	fallbackUsed bool,
	fallbackReasonCode *string,
	candidateCountBeforeFilters int,
	candidateCountAfterFilters int,
	returnedPrimaryCount int,
	returnedAlternativeCount int,
	finishedAt time.Time,
) error {
	if source != RankingSourceML && source != RankingSourceFallback {
		return ErrInvalidRankingSource
	}

	r.status = StatusCompleted
	if fallbackUsed {
		r.status = StatusCompletedWithFallback
	}
	r.rankingSource = source
	r.candidateCountBeforeFilters = candidateCountBeforeFilters
	r.candidateCountAfterFilters = candidateCountAfterFilters
	r.returnedPrimaryCount = returnedPrimaryCount
	r.returnedAlternativeCount = returnedAlternativeCount
	r.fallbackReasonCode = cloneStringPtr(fallbackReasonCode)
	r.failureCode = nil
	r.failureMessage = nil

	completedAt := finishedAt.UTC()
	r.finishedAt = &completedAt
	r.updatedAt = completedAt

	return nil
}

func (r *RecommendationRequest) CompleteEmpty(
	fallbackReasonCode *string,
	candidateCountBeforeFilters int,
	candidateCountAfterFilters int,
	finishedAt time.Time,
) {
	r.status = StatusCompletedEmpty
	r.rankingSource = RankingSourceNone
	r.candidateCountBeforeFilters = candidateCountBeforeFilters
	r.candidateCountAfterFilters = candidateCountAfterFilters
	r.returnedPrimaryCount = 0
	r.returnedAlternativeCount = 0
	r.fallbackReasonCode = cloneStringPtr(fallbackReasonCode)
	r.failureCode = nil
	r.failureMessage = nil

	completedAt := finishedAt.UTC()
	r.finishedAt = &completedAt
	r.updatedAt = completedAt
}

func (r *RecommendationRequest) Fail(code, message string, finishedAt time.Time) {
	r.status = StatusFailed
	r.rankingSource = RankingSourceNone
	r.failureCode = cloneStringPtr(&code)
	r.failureMessage = cloneStringPtr(&message)

	completedAt := finishedAt.UTC()
	r.finishedAt = &completedAt
	r.updatedAt = completedAt
}

func (r RecommendationRequest) ID() RequestID {
	return r.id
}

func (r RecommendationRequest) RequestedByUserID() *userdomain.UserID {
	return cloneUserID(r.requestedByUserID)
}

func (r RecommendationRequest) Questionnaire() Questionnaire {
	return r.questionnaire
}

func (r RecommendationRequest) Status() RequestStatus {
	return r.status
}

func (r RecommendationRequest) RankingSource() RankingSource {
	return r.rankingSource
}

func (r RecommendationRequest) CandidateCountBeforeFilters() int {
	return r.candidateCountBeforeFilters
}

func (r RecommendationRequest) CandidateCountAfterFilters() int {
	return r.candidateCountAfterFilters
}

func (r RecommendationRequest) ReturnedPrimaryCount() int {
	return r.returnedPrimaryCount
}

func (r RecommendationRequest) ReturnedAlternativeCount() int {
	return r.returnedAlternativeCount
}

func (r RecommendationRequest) FallbackReasonCode() *string {
	return cloneStringPtr(r.fallbackReasonCode)
}

func (r RecommendationRequest) FailureCode() *string {
	return cloneStringPtr(r.failureCode)
}

func (r RecommendationRequest) FailureMessage() *string {
	return cloneStringPtr(r.failureMessage)
}

func (r RecommendationRequest) StartedAt() time.Time {
	return r.startedAt
}

func (r RecommendationRequest) FinishedAt() *time.Time {
	return cloneTimePtr(r.finishedAt)
}

func (r RecommendationRequest) CreatedAt() time.Time {
	return r.createdAt
}

func (r RecommendationRequest) UpdatedAt() time.Time {
	return r.updatedAt
}

func parseRequestStatus(raw string) (RequestStatus, error) {
	switch RequestStatus(raw) {
	case StatusPending, StatusRunning, StatusCompleted, StatusCompletedWithFallback, StatusCompletedEmpty, StatusFailed:
		return RequestStatus(raw), nil
	default:
		return "", ErrInvalidRequestStatus
	}
}

func parseRankingSource(raw string) (RankingSource, error) {
	switch RankingSource(raw) {
	case RankingSourceNone, RankingSourceML, RankingSourceFallback:
		return RankingSource(raw), nil
	default:
		return "", ErrInvalidRankingSource
	}
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := value.UTC()
	return &cloned
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func cloneUserID(value *userdomain.UserID) *userdomain.UserID {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
