package config

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestPostgresPoolIdleTimeIntegration(t *testing.T) {
	if os.Getenv("RUN_POSTGRES_POOL_INTEGRATION") != "1" {
		t.Skip("set RUN_POSTGRES_POOL_INTEGRATION=1 to run")
	}

	dsn := os.Getenv("POSTGRES_POOL_TEST_DSN")
	if dsn == "" {
		t.Fatal("POSTGRES_POOL_TEST_DSN is required")
	}

	db, err := sql.Open("postgres", WithApplicationName(dsn, "evogo-pool-test"))
	if err != nil {
		t.Fatalf("failed to open postgres connection: %v", err)
	}
	defer db.Close()

	ApplyDBPool(db, DBPoolConfig{
		Name:            "integration",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: time.Second,
	})

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}
	if _, err := db.Exec("SELECT 1"); err != nil {
		t.Fatalf("failed to execute test query: %v", err)
	}

	time.Sleep(2500 * time.Millisecond)

	stats := db.Stats()
	if stats.MaxIdleTimeClosed == 0 && stats.OpenConnections > 0 {
		t.Fatalf("expected idle-time cleaner to close idle connection, stats=%+v", stats)
	}
}
