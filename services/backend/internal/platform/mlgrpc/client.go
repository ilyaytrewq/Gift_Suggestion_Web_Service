package mlgrpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

var ErrRankingNotImplemented = errors.New("ml ranking rpc is not implemented yet")

type RankRequest struct {
	SelectionID  string
	CandidateIDs []string
}

type RankResponse struct {
	RankedCandidateIDs []string
}

type Client struct {
	conn   *grpc.ClientConn
	health grpc_health_v1.HealthClient
	config config.MLConfig
}

func NewClient(ctx context.Context, cfg config.MLConfig) (*Client, error) {
	client := &Client{
		config: cfg,
	}

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

func (c *Client) RankCandidates(context.Context, RankRequest) (RankResponse, error) {
	return RankResponse{}, ErrRankingNotImplemented
}
