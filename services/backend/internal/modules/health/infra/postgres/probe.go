package postgres

import (
	"context"
	"database/sql"
)

type Probe struct {
	db *sql.DB
}

func NewProbe(db *sql.DB) *Probe {
	return &Probe{db: db}
}

func (p *Probe) Ping(ctx context.Context) error {
	if p == nil || p.db == nil {
		return sql.ErrConnDone
	}

	return p.db.PingContext(ctx)
}
