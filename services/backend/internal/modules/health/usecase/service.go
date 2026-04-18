package usecase

import (
	"context"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/health/domain"
)

type Service struct {
	clock        Clock
	dependencies []Dependency
}

func NewService(clock Clock, dependencies []Dependency) (*Service, error) {
	if clock == nil {
		return nil, ErrNilClock
	}

	return &Service{
		clock:        clock,
		dependencies: append([]Dependency(nil), dependencies...),
	}, nil
}

func (s *Service) Live(context.Context) domain.Report {
	return domain.NewLiveReport(s.clock.Now())
}

func (s *Service) Ready(ctx context.Context) domain.Report {
	components := make([]domain.Component, 0, len(s.dependencies))

	for _, dependency := range s.dependencies {
		component := domain.Component{
			Name:     dependency.Name,
			Required: dependency.Required,
		}

		switch {
		case !dependency.Enabled:
			component.Status = domain.StatusSkipped
		case dependency.Probe == nil:
			component.Status = domain.StatusDown
			component.Error = "probe is not configured"
		default:
			if err := dependency.Probe.Ping(ctx); err != nil {
				component.Status = domain.StatusDown
				component.Error = err.Error()
			} else {
				component.Status = domain.StatusUp
			}
		}

		components = append(components, component)
	}

	return domain.NewReadinessReport(s.clock.Now(), components)
}
