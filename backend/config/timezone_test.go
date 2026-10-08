package config

import (
	"os"
	"testing"
	"time"
)

func TestApplyProcessTimezone_SetsLocal(t *testing.T) {
	orig := time.Local
	t.Cleanup(func() { time.Local = orig })
	time.Local = time.UTC

	if err := ApplyProcessTimezone("Asia/Jakarta"); err != nil {
		t.Fatalf("ApplyProcessTimezone: %v", err)
	}

	if got := time.Local.String(); got != "Asia/Jakarta" {
		t.Fatalf("time.Local = %q, want Asia/Jakarta", got)
	}
	if _, offset := time.Now().Zone(); offset != 7*60*60 {
		t.Fatalf("local offset = %ds, want %ds", offset, 7*60*60)
	}
}

func TestApplyProcessTimezone_InvalidZoneKeepsLocal(t *testing.T) {
	orig := time.Local
	t.Cleanup(func() { time.Local = orig })
	time.Local = time.UTC

	if err := ApplyProcessTimezone("Mars/Olympus"); err == nil {
		t.Fatal("expected error for unknown zone")
	}
	if time.Local != time.UTC {
		t.Fatalf("time.Local changed to %s on error", time.Local)
	}
}

// TestOpenDatabase_TimestampRoundTrip reproduces the "notification 7 hours
// ago" bug: columns are naive TIMESTAMP, the session TimeZone is
// Asia/Jakarta, and the production container runs with UTC as its local
// zone. A time.Now() written by the backend must read back as the same
// instant. Needs a real Postgres: set TEST_PG_HOST (and optionally
// TEST_PG_PORT/USER/PASSWORD/DB).
func TestOpenDatabase_TimestampRoundTrip(t *testing.T) {
	host := os.Getenv("TEST_PG_HOST")
	if host == "" {
		t.Skip("TEST_PG_HOST not set")
	}
	orig := time.Local
	t.Cleanup(func() { time.Local = orig })
	time.Local = time.UTC // same as the alpine backend container

	db, err := OpenDatabase(&Config{
		DBHost:     host,
		DBPort:     envOr("TEST_PG_PORT", "5432"),
		DBUser:     envOr("TEST_PG_USER", "t"),
		DBPassword: envOr("TEST_PG_PASSWORD", "t"),
		DBName:     envOr("TEST_PG_DB", "t"),
		DBSSLMode:  "disable",
		DBTimezone: "Asia/Jakarta",
	})
	if err != nil {
		t.Fatalf("OpenDatabase: %v", err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })

	type tzRoundTrip struct {
		ID        int       `gorm:"column:id;primaryKey"`
		CreatedAt time.Time `gorm:"column:CreatedAt;autoCreateTime"`
	}
	const table = "tz_round_trip_test"
	db.Exec(`DROP TABLE IF EXISTS ` + table)
	if err := db.Exec(`CREATE TABLE ` + table + ` (id serial PRIMARY KEY, "CreatedAt" TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() { db.Exec(`DROP TABLE IF EXISTS ` + table) })

	written := time.Now()
	if err := db.Table(table).Create(&tzRoundTrip{CreatedAt: written}).Error; err != nil {
		t.Fatalf("insert from Go: %v", err)
	}
	if err := db.Exec(`INSERT INTO ` + table + ` DEFAULT VALUES`).Error; err != nil {
		t.Fatalf("insert with DB default: %v", err)
	}

	var rows []tzRoundTrip
	if err := db.Table(table).Order("id").Find(&rows).Error; err != nil {
		t.Fatalf("select: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	for i, r := range rows {
		if skew := written.Sub(r.CreatedAt); skew > time.Minute || skew < -time.Minute {
			t.Errorf("row %d: read back %s, written %s (skew %s)", i, r.CreatedAt.Format(time.RFC3339), written.Format(time.RFC3339), skew.Round(time.Minute))
		}
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
