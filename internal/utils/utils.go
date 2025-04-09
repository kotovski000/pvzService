package utils

import "database/sql"
import "time"

func NullableTime(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}
