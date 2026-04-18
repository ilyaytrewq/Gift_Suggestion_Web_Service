package domain

import "time"

type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusSkipped Status = "skipped"
)

type Component struct {
	Name     string `json:"name"`
	Status   Status `json:"status"`
	Required bool   `json:"required"`
	Error    string `json:"error,omitempty"`
}

type Report struct {
	Status     Status      `json:"status"`
	CheckedAt  time.Time   `json:"checked_at"`
	Components []Component `json:"components,omitempty"`
}

func NewLiveReport(now time.Time) Report {
	return Report{
		Status:    StatusUp,
		CheckedAt: now.UTC(),
	}
}

func NewReadinessReport(now time.Time, components []Component) Report {
	report := Report{
		Status:     StatusUp,
		CheckedAt:  now.UTC(),
		Components: append([]Component(nil), components...),
	}

	for _, component := range components {
		if component.Required && component.Status == StatusDown {
			report.Status = StatusDown
			break
		}
	}

	return report
}
