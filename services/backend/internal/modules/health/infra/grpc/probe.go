package grpc

import "context"

type healthChecker interface {
	HealthCheck(ctx context.Context) error
}

type Probe struct {
	client healthChecker
}

func NewProbe(client healthChecker) *Probe {
	return &Probe{client: client}
}

func (p *Probe) Ping(ctx context.Context) error {
	if p == nil || p.client == nil {
		return context.DeadlineExceeded
	}

	return p.client.HealthCheck(ctx)
}
