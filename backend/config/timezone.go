package config

import (
	"fmt"
	"time"
	_ "time/tzdata" // LoadLocation works even without OS zoneinfo
)

// ApplyProcessTimezone makes the Go process' local zone match the database
// session zone (DB_TIMEZONE).
//
// Every timestamp column is a naive TIMESTAMP. pgx writes a time.Time as its
// wall clock in time.Local and the session reads it back as a wall clock in
// DB_TIMEZONE, so the two zones must be the same. The alpine container has no
// TZ set (time.Local = UTC), which made fresh rows read back 7 hours in the
// past ("7 jam lalu").
func ApplyProcessTimezone(name string) error {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return fmt.Errorf("load timezone %q: %w", name, err)
	}
	time.Local = loc
	return nil
}
