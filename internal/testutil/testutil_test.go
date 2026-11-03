package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPools(t *testing.T) {
	conf := LoadConfig(t)

	tests := []struct {
		name string
		pool *pgxpool.Pool
	}{
		{name: "server", pool: NewServerPool(t, conf)},
		{name: "owner", pool: NewOwnerPool(t, conf)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dbName string
			if err := tt.pool.QueryRow(context.Background(), "SELECT current_database()").Scan(&dbName); err != nil {
				t.Fatalf("current_database query failed: %v", err)
			}
			if dbName != conf.TestDatabase.Database {
				t.Errorf("connected to %q, want %q", dbName, conf.TestDatabase.Database)
			}
		})
	}
}
