package db

import (
	"os"
	"testing"
)

func TestDB_Connect_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 to run integration tests")
	}

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://cierge_user:cierge_pass@localhost:5432/subsync?sslmode=disable"
	}

	ctx := t.Context()
	database, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer database.Close()

	if err := database.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}
}
