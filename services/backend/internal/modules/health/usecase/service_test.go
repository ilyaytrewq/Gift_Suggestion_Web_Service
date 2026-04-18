package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/domain"
)

type stubClock struct {
	now time.Time
}

func (c stubClock) Now() time.Time {
	return c.now
}

type stubProbe struct {
	err error
}

func (p stubProbe) Ping(context.Context) error {
	return p.err
}

func TestNewServiceRequiresClock(t *testing.T) {
	t.Parallel()

	service, err := NewService(nil, nil)
	if !errors.Is(err, ErrNilClock) {
		t.Fatalf("expected ErrNilClock, got %v", err)
	}
	if service != nil {
		t.Fatal("expected nil service when clock is nil")
	}
}

func TestServiceLiveReturnsUp(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 18, 9, 0, 0, 0, time.UTC)
	service, err := NewService(stubClock{now: now}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report := service.Live(context.Background())

	if report.Status != domain.StatusUp {
		t.Fatalf("expected status up, got %s", report.Status)
	}
	if !report.CheckedAt.Equal(now) {
		t.Fatalf("expected checked_at %s, got %s", now, report.CheckedAt)
	}
}

func TestServiceReadyMarksOptionalDisabledDependencyAsSkipped(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 18, 9, 0, 0, 0, time.UTC)
	service, err := NewService(
		stubClock{now: now},
		[]Dependency{
			{
				Name:     "postgres",
				Required: true,
				Enabled:  true,
				Probe:    stubProbe{},
			},
			{
				Name:     "ml_service",
				Required: false,
				Enabled:  false,
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report := service.Ready(context.Background())

	if report.Status != domain.StatusUp {
		t.Fatalf("expected status up, got %s", report.Status)
	}
	if len(report.Components) != 2 {
		t.Fatalf("expected 2 components, got %d", len(report.Components))
	}
	if report.Components[1].Status != domain.StatusSkipped {
		t.Fatalf("expected skipped status for optional dependency, got %s", report.Components[1].Status)
	}
}

func TestServiceReadyReturnsDownWhenRequiredDependencyFails(t *testing.T) {
	t.Parallel()

	service, err := NewService(
		stubClock{now: time.Now()},
		[]Dependency{
			{
				Name:     "postgres",
				Required: true,
				Enabled:  true,
				Probe: stubProbe{
					err: errors.New("connection refused"),
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report := service.Ready(context.Background())

	if report.Status != domain.StatusDown {
		t.Fatalf("expected status down, got %s", report.Status)
	}
	if report.Components[0].Error == "" {
		t.Fatal("expected component error to be populated")
	}
}
