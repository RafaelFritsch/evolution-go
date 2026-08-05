package config

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestParseEnvIntDefaultAndOverride(t *testing.T) {
	t.Setenv("TEST_POOL_INT", "")
	if got := parseEnvInt("TEST_POOL_INT", 10); got != 10 {
		t.Fatalf("expected default 10, got %d", got)
	}

	t.Setenv("TEST_POOL_INT", "7")
	if got := parseEnvInt("TEST_POOL_INT", 10); got != 7 {
		t.Fatalf("expected override 7, got %d", got)
	}

	t.Setenv("TEST_POOL_INT", "-1")
	if got := parseEnvInt("TEST_POOL_INT", 10); got != 10 {
		t.Fatalf("expected default for invalid value, got %d", got)
	}
}

func TestParseEnvDurationSupportsDurationAndSeconds(t *testing.T) {
	t.Setenv("TEST_POOL_DURATION", "")
	if got := parseEnvDuration("TEST_POOL_DURATION", 5*time.Second); got != 5*time.Second {
		t.Fatalf("expected default 5s, got %s", got)
	}

	t.Setenv("TEST_POOL_DURATION", "250ms")
	if got := parseEnvDuration("TEST_POOL_DURATION", 5*time.Second); got != 250*time.Millisecond {
		t.Fatalf("expected 250ms, got %s", got)
	}

	t.Setenv("TEST_POOL_DURATION", "3")
	if got := parseEnvDuration("TEST_POOL_DURATION", 5*time.Second); got != 3*time.Second {
		t.Fatalf("expected 3s, got %s", got)
	}
}

func TestApplyDBPoolSetsMaxOpenConnections(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	ApplyDBPool(db, DBPoolConfig{
		Name:            "test",
		MaxOpenConns:    4,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	})

	if got := db.Stats().MaxOpenConnections; got != 4 {
		t.Fatalf("expected max open connections 4, got %d", got)
	}
}

func TestWithApplicationNameAddsOrPreservesName(t *testing.T) {
	urlDSN := WithApplicationName("postgresql://user:pass@localhost:5432/app?sslmode=disable", "evogo-users")
	if urlDSN == "" || urlDSN == "postgresql://user:pass@localhost:5432/app?sslmode=disable" {
		t.Fatalf("expected application_name to be added to url DSN, got %q", urlDSN)
	}

	kvDSN := WithApplicationName("host=localhost user=postgres dbname=app sslmode=disable", "evogo-auth")
	if kvDSN != "host=localhost user=postgres dbname=app sslmode=disable application_name=evogo-auth" {
		t.Fatalf("unexpected key/value DSN: %q", kvDSN)
	}
}
