package usecase

import (
	"context"
	"time"
)

type Probe interface {
	Ping(ctx context.Context) error
}

type Clock interface {
	Now() time.Time
}

type Dependency struct {
	Name     string
	Required bool
	Enabled  bool
	Probe    Probe
}
