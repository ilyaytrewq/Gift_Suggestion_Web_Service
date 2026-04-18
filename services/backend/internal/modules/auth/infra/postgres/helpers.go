package postgres

import (
	"database/sql"
	"time"
)

func nullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{
		Time:  value.UTC(),
		Valid: true,
	}
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	timeValue := value.Time.UTC()
	return &timeValue
}
