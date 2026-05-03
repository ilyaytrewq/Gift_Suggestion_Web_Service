package mlgrpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
	rankingv1 "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/mlgrpc/gen/ranking/v1"
)

var ErrRankingNotImplemented = errors.New("ml ranking rpc is not implemented yet")

type QueryContext struct {
	Occasion        string
	Relationship    string
	Interests       []string
	BudgetCents     int64
	RecipientAge    *int
	RecipientGender string // "male", "female", "other", or "" for unspecified
}

type Candidate struct {
	ID              string
	CategoryID      *string
	CategoryName    string
	PriceCents      int64
	AgeRestriction  *int
	Title           string
	Description     string
	AlreadyInWishlist bool
}

type Explanation struct {
	Code string
	Text string
}

type RankRequest struct {
	SelectionID string
	UserID      string
	TopN        int
	Query       QueryContext
	Candidates  []Candidate
}

type RankedItem struct {
	CandidateID             string
	Score                   float64
	Explanations            []Explanation
	AlternativeCandidateIDs []string
}

type RankResponse struct {
	Items        []RankedItem
	ModelVersion string
}

type Client struct {
	conn          *grpc.ClientConn
	health        grpc_health_v1.HealthClient
	rankingClient rankingv1.RankingServiceClient
	config        config.MLConfig
}

func NewClient(ctx context.Context, cfg config.MLConfig) (*Client, error) {
	client := &Client{config: cfg}

	if !cfg.Enabled {
		return client, nil
	}

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client.conn = conn
	client.health = grpc_health_v1.NewHealthClient(conn)
	client.rankingClient = rankingv1.NewRankingServiceClient(conn)

	healthCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
	defer cancel()
	if err := client.HealthCheck(healthCtx); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}

	return client, nil
}

func (c *Client) Enabled() bool {
	if c == nil {
		return false
	}
	return c.config.Enabled
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) HealthCheck(ctx context.Context) error {
	if c == nil || !c.Enabled() {
		return nil
	}
	_, err := c.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	return err
}

func (c *Client) RankCandidates(ctx context.Context, req RankRequest) (RankResponse, error) {
	if c == nil || !c.Enabled() || c.rankingClient == nil {
		return RankResponse{}, ErrRankingNotImplemented
	}

	protoReq := toProtoRankRequest(req)
	protoResp, err := c.rankingClient.Rank(ctx, protoReq)
	if err != nil {
		return RankResponse{}, err
	}

	return fromProtoRankResponse(protoResp), nil
}

func toProtoRankRequest(req RankRequest) *rankingv1.RankRequest {
	query := &rankingv1.QueryContext{
		Occasion:        req.Query.Occasion,
		Relationship:    req.Query.Relationship,
		Interests:       append([]string(nil), req.Query.Interests...),
		BudgetCents:     req.Query.BudgetCents,
		RecipientGender: req.Query.RecipientGender,
	}
	if req.Query.RecipientAge != nil {
		age := int32(*req.Query.RecipientAge)
		query.RecipientAge = &age
	}

	candidates := make([]*rankingv1.Candidate, 0, len(req.Candidates))
	for _, c := range req.Candidates {
		proto := &rankingv1.Candidate{
			Id:              c.ID,
			CategoryName:    c.CategoryName,
			PriceCents:      c.PriceCents,
			Title:           c.Title,
			Description:     c.Description,
			AlreadyInWishlist: c.AlreadyInWishlist,
		}
		if c.CategoryID != nil {
			proto.CategoryId = c.CategoryID
		}
		if c.AgeRestriction != nil {
			age := int32(*c.AgeRestriction)
			proto.AgeRestriction = &age
		}
		candidates = append(candidates, proto)
	}

	return &rankingv1.RankRequest{
		SelectionId: req.SelectionID,
		UserId:      req.UserID,
		TopN:        int32(req.TopN),
		Query:       query,
		Candidates:  candidates,
	}
}

func fromProtoRankResponse(resp *rankingv1.RankResponse) RankResponse {
	if resp == nil {
		return RankResponse{}
	}

	items := make([]RankedItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		explanations := make([]Explanation, 0, len(item.Explanations))
		for _, e := range item.Explanations {
			explanations = append(explanations, Explanation{Code: e.Code, Text: e.Text})
		}
		items = append(items, RankedItem{
			CandidateID:             item.CandidateId,
			Score:                   item.Score,
			Explanations:            explanations,
			AlternativeCandidateIDs: append([]string(nil), item.AlternativeCandidateIds...),
		})
	}

	return RankResponse{
		Items:        items,
		ModelVersion: resp.ModelVersion,
	}
}
