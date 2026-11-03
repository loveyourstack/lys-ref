// Package testutil provides helpers for integration tests that run against the local test database.
// The database is created with "make testdb".
package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys/lyspgdb"
)

const configFileName = "ref_config.toml"

// findConfigPath walks up from the working directory to the module root and returns the path of the config file found there
func findConfigPath() (string, error) {

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, configFileName), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found in any parent directory")
		}
		dir = parent
	}
}

// LoadConfig loads ref_config.toml from the module root and points it at the test database.
// It fails the test (rather than skipping) if the config is unusable or could target a non-test database.
func LoadConfig(t testing.TB) *myapp.Config {
	t.Helper()

	path, err := findConfigPath()
	if err != nil {
		t.Fatalf("finding config file: %v", err)
	}

	conf := &myapp.Config{}
	if err = conf.LoadFromFile(path); err != nil {
		t.Fatalf("loading %s: %v", path, err)
	}

	if conf.General.Env != appenv.Dev {
		t.Fatalf("integration tests may only run when General.Env is %q, not %q", appenv.Dev, conf.General.Env)
	}
	if conf.TestDatabase.Database == "" {
		t.Fatal("testDatabase.database is not set in the config")
	}
	if conf.TestDatabase.Database == conf.Db.Database {
		t.Fatalf("testDatabase.database must differ from database (%q)", conf.Db.Database)
	}

	conf.Db.Database = conf.TestDatabase.Database

	return conf
}

// NewServerPool opens a pool on the test database as the server user, with the app's custom types registered.
// The pool is closed when the test ends.
func NewServerPool(t testing.TB, conf *myapp.Config) *pgxpool.Pool {
	t.Helper()

	pool, err := lyspgdb.GetPoolWithTypes(context.Background(), conf.Db, conf.DbServerUser, conf.General.AppName+" Test", myapp.TypesToRegister)
	if err != nil {
		t.Fatalf("connecting to test database %q as server user (run \"make testdb\" first): %v", conf.Db.Database, err)
	}
	t.Cleanup(pool.Close)

	return pool
}

// NewOwnerPool opens a pool on the test database as the owner user, for fixture setup and cleanup.
// The app's custom types are registered so that the pool can be used with the stores. The pool is closed when the test ends.
func NewOwnerPool(t testing.TB, conf *myapp.Config) *pgxpool.Pool {
	t.Helper()

	pool, err := lyspgdb.GetPoolWithTypes(context.Background(), conf.Db, conf.DbOwnerUser, conf.General.AppName+" Test (owner)", myapp.TypesToRegister)
	if err != nil {
		t.Fatalf("connecting to test database %q as owner user (run \"make testdb\" first): %v", conf.Db.Database, err)
	}
	t.Cleanup(pool.Close)

	return pool
}
