package models

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMain starts an ephemeral PostgreSQL container, applies migrations, then
// runs all model tests against it. Docker daemon must be running.
//
// gobuffalo/suite is intentionally not used here: it creates its own
// pop.Connect() call which bypasses the container-backed DB set up below.
// Plain tests use the package-level DB variable directly.
func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("ohm_test"),
		postgres.WithUsername("ohm"),
		postgres.WithPassword("ohm"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}
	defer pg.Terminate(ctx) //nolint:errcheck

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("get connection string: %v", err)
	}

	conn, err := pop.NewConnection(&pop.ConnectionDetails{URL: connStr})
	if err != nil {
		log.Fatalf("create pop connection: %v", err)
	}
	if err := conn.Open(); err != nil {
		log.Fatalf("open pop connection: %v", err)
	}
	DB = conn // override package-level DB; all tests use this connection

	// Test binary runs from webapp/models/ — migrations are two levels up.
	migrator, err := pop.NewFileMigrator("../../db/migrations", conn)
	if err != nil {
		log.Fatalf("create migrator: %v", err)
	}
	if err := migrator.Up(); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	os.Exit(m.Run())
}

// truncateAll clears all non-migration tables between tests.
func truncateAll(t *testing.T) {
	t.Helper()
	if err := DB.TruncateAll(); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
