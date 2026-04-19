package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/mlgrpc"
)

func TestGatewayMapsRankRequest(t *testing.T) {
	t.Parallel()

	client := &stubRankClient{
		response: mlgrpc.RankResponse{
			Items: []mlgrpc.RankedItem{
				{CandidateID: testGatewayGift1.ID().String(), Score: 0.9},
			},
		},
	}
	gateway := NewGateway(client, config.MLConfig{RequestTimeout: testGatewayTimeout, MaxRetries: 0})

	_, err := gateway.Rank(context.Background(), gatewayRankInput())
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	if client.request.SelectionID != testGatewayRequestID {
		t.Fatalf("SelectionID = %q, want %q", client.request.SelectionID, testGatewayRequestID)
	}
	if client.request.TopN != 2 {
		t.Fatalf("TopN = %d, want 2", client.request.TopN)
	}
	if client.request.Query.BudgetCents != 10000 {
		t.Fatalf("BudgetCents = %d, want 10000", client.request.Query.BudgetCents)
	}
	if len(client.request.Candidates) != 2 {
		t.Fatalf("Candidates = %d, want 2", len(client.request.Candidates))
	}
}

func TestGatewayMapsRankResponse(t *testing.T) {
	t.Parallel()

	explanation := mlgrpc.Explanation{Code: "ml_reason", Text: "Top match"}
	client := &stubRankClient{
		response: mlgrpc.RankResponse{
			Items: []mlgrpc.RankedItem{
				{
					CandidateID:             testGatewayGift2.ID().String(),
					Score:                   0.91,
					Explanations:            []mlgrpc.Explanation{explanation},
					AlternativeCandidateIDs: []string{testGatewayGift1.ID().String()},
				},
			},
		},
	}
	gateway := NewGateway(client, config.MLConfig{RequestTimeout: testGatewayTimeout, MaxRetries: 0})

	output, err := gateway.Rank(context.Background(), gatewayRankInput())
	if err != nil {
		t.Fatalf("Rank() error = %v", err)
	}

	if len(output.Items) != 1 {
		t.Fatalf("Items = %d, want 1", len(output.Items))
	}
	if output.Items[0].GiftID.String() != testGatewayGift2.ID().String() {
		t.Fatalf("GiftID = %q, want %q", output.Items[0].GiftID.String(), testGatewayGift2.ID().String())
	}
	if len(output.Items[0].Explanations) != 1 {
		t.Fatalf("Explanations = %d, want 1", len(output.Items[0].Explanations))
	}
	if len(output.Items[0].AlternativeGiftIDs) != 1 {
		t.Fatalf("AlternativeGiftIDs = %d, want 1", len(output.Items[0].AlternativeGiftIDs))
	}
}

func TestGatewayMapsUnavailableAndDeadlineErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "deadline",
			err:     status.Error(codes.DeadlineExceeded, "timeout"),
			wantErr: recommendationusecase.ErrRankingTimedOut,
		},
		{
			name:    "unavailable",
			err:     status.Error(codes.Unavailable, "down"),
			wantErr: recommendationusecase.ErrRankingUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gateway := NewGateway(&stubRankClient{err: tt.err}, config.MLConfig{RequestTimeout: testGatewayTimeout, MaxRetries: 0})
			_, err := gateway.Rank(context.Background(), gatewayRankInput())
			if err == nil {
				t.Fatal("Rank() expected error")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Rank() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGatewayRejectsMalformedResponse(t *testing.T) {
	t.Parallel()

	client := &stubRankClient{
		response: mlgrpc.RankResponse{
			Items: []mlgrpc.RankedItem{
				{CandidateID: "550e8400-e29b-41d4-a716-446655449999", Score: 0.9},
			},
		},
	}
	gateway := NewGateway(client, config.MLConfig{RequestTimeout: testGatewayTimeout, MaxRetries: 0})

	_, err := gateway.Rank(context.Background(), gatewayRankInput())
	if !errors.Is(err, recommendationusecase.ErrInvalidRankingResponse) {
		t.Fatalf("Rank() error = %v, want %v", err, recommendationusecase.ErrInvalidRankingResponse)
	}
}

var (
	testGatewayTimeout = 2500 * time.Millisecond
	testGatewayGift1   = mustGatewayGift("550e8400-e29b-41d4-a716-446655443001", "Gift One", "50.00")
	testGatewayGift2   = mustGatewayGift("550e8400-e29b-41d4-a716-446655443002", "Gift Two", "70.00")
)

const (
	testGatewayRequestID = "550e8400-e29b-41d4-a716-446655443000"
	testGatewayUserID    = "550e8400-e29b-41d4-a716-446655443010"
)

type stubRankClient struct {
	request  mlgrpc.RankRequest
	response mlgrpc.RankResponse
	err      error
}

func (c *stubRankClient) RankCandidates(_ context.Context, request mlgrpc.RankRequest) (mlgrpc.RankResponse, error) {
	c.request = request
	return c.response, c.err
}

func gatewayRankInput() recommendationusecase.RankInput {
	recipientAge := 16
	budget, _ := catalogdomain.NewPrice("100.00")
	return recommendationusecase.RankInput{
		RequestID:    testGatewayRequestID,
		UserID:       testGatewayUserID,
		Occasion:     "birthday",
		Relationship: "friend",
		RecipientAge: &recipientAge,
		BudgetMax:    budget,
		Interests:    []string{"books"},
		TopN:         2,
		Candidates: []recommendationusecase.RankCandidate{
			{Gift: testGatewayGift1},
			{Gift: testGatewayGift2},
		},
	}
}

func mustGatewayGift(id, name, price string) catalogdomain.Gift {
	categoryID := "550e8400-e29b-41d4-a716-446655443100"
	categoryName := "Books"
	gift, err := catalogdomain.RestoreGift(
		id,
		&categoryID,
		&categoryName,
		name,
		"Description",
		price,
		"https://example.com/"+id,
		nil,
		nil,
		testGatewayNow(),
		testGatewayNow(),
	)
	if err != nil {
		panic(err)
	}

	return gift
}

func testGatewayNow() time.Time {
	return time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
}
