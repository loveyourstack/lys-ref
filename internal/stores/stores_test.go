package stores_test

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/netip"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/stores/core/corearraytype"
	"github.com/loveyourstack/lys-ref/internal/stores/core/coredefaultvalue"
	"github.com/loveyourstack/lys-ref/internal/stores/core/coremandatoryvalue"
	"github.com/loveyourstack/lys-ref/internal/stores/core/coreoptionalvalue"
	"github.com/loveyourstack/lys-ref/internal/stores/core/corevarianttype"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaign"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampaignopt"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampperf"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmcampperfagg"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmgencamp"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchfb"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmlaunchgads"
	"github.com/loveyourstack/lys-ref/internal/stores/digmark/dmvertical"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurr"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbcurrmd"
	"github.com/loveyourstack/lys-ref/internal/stores/ecb/ecbxrperfnorm"
	"github.com/loveyourstack/lys-ref/internal/stores/gemini/gemapicall"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geocountry"
	"github.com/loveyourstack/lys-ref/internal/stores/geo/geoocean"
	"github.com/loveyourstack/lys-ref/internal/stores/lysinc/lysincemployee"
	"github.com/loveyourstack/lys-ref/internal/stores/lysinc/lysinceventsinmonth"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procflow"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procpoint"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procrun"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procstep"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procsteplink"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubauthor"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubauthorarch"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubbook"
	"github.com/loveyourstack/lys-ref/internal/stores/publisher/pubbookarch"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppcompany"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppemployee"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppprodcategory"
	"github.com/loveyourstack/lys-ref/internal/stores/supplier/suppproduct"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysextdatasync"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysnotification"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcloginattempt"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysrpcsessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssession"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssessionhist"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssvcaccount"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysuser"
	"github.com/loveyourstack/lys-ref/internal/stores/tedb/tedbvatratesumm"
	"github.com/loveyourstack/lys-ref/internal/testutil"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

// smoke adapts any store's Select method, whatever its Model type, to a uniform func.
// An empty result is an error, since scan errors can't be detected without rows.
func smoke[M any](sel func(context.Context, lyspg.SelectParams) ([]M, lyspg.TotalCount, error), setFuncParamVals ...any) func(context.Context) error {
	return func(ctx context.Context) error {
		items, _, err := sel(ctx, lyspg.SelectParams{Limit: 1, SetFuncParamValues: setFuncParamVals})
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return errors.New("no rows returned: add test data")
		}
		return nil
	}
}

// withRows inserts a known test row before running run, and deletes it afterwards.
// deleteStmt must delete exactly the inserted row: it also runs before the insert to clear leftovers of an aborted run.
//
// This is used instead of the SQL test data for tables that other tests empty. The refsrv tests truncate
// system.session and system.login_attempt (the app persists its in-memory state there on shutdown),
// so rows seeded by "make testdb" would not survive a test run.
func withRows(pool *pgxpool.Pool, insert func(context.Context, pgx.Tx) error, deleteStmt string, run func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		if _, err = pool.Exec(ctx, deleteStmt); err != nil {
			return fmt.Errorf("clearing test row: %w", err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning tx: %w", err)
		}
		defer tx.Rollback(ctx)

		if err = insert(ctx, tx); err != nil {
			return fmt.Errorf("inserting test row: %w", err)
		}
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("committing test row: %w", err)
		}

		defer func() {
			if _, delErr := pool.Exec(ctx, deleteStmt); delErr != nil {
				err = errors.Join(err, fmt.Errorf("deleting test row: %w", delErr))
			}
		}()

		return run(ctx)
	}
}

const (
	smokeSessionToken = "smoke-test-token"
	smokeLoginAttempt = "198.51.100.1" // TEST-NET-2: not used by the refsrv tests
)

// TestStoresSelect runs each store's Select method with a smoke test.
// This checks the store's columns against the Postgres table or view as defined in /sql/ddl.
func TestStoresSelect(t *testing.T) {
	conf := testutil.LoadConfig(t)
	pool := testutil.NewOwnerPool(t, conf)

	sessionStore := syssession.Store{Db: pool}
	loginAttemptStore := sysloginattempt.Store{Db: pool}

	tests := []struct {
		name string
		run  func(context.Context) error
	}{
		{"corearraytype", smoke(corearraytype.Store{Db: pool}.Select)},
		{"coredefaultvalue", smoke(coredefaultvalue.Store{Db: pool}.Select)},
		{"coremandatoryvalue", smoke(coremandatoryvalue.Store{Db: pool}.Select)},
		{"coreoptionalvalue", smoke(coreoptionalvalue.Store{Db: pool}.Select)},
		{"corevarianttype", smoke(corevarianttype.Store{Db: pool}.Select)},

		{"dmcampaign", smoke(dmcampaign.Store{Db: pool}.Select)},
		{"dmcampaignopt", smoke(dmcampaignopt.Store{Db: pool}.Select, "Today")},
		{"dmcampperf", smoke(dmcampperf.Store{Db: pool}.Select)},
		{"dmcampperfagg", smoke(dmcampperfagg.Store{Db: pool}.Select)},
		{"dmgencamp", smoke(dmgencamp.Store{Db: pool}.Select)},
		//{"dmlaunch", smoke(dmlaunch.Store{Db: pool}.Select)}, // abstract Model: not expected to match table columns
		{"dmlaunchfb", smoke(dmlaunchfb.Store{Db: pool}.Select)},
		{"dmlaunchgads", smoke(dmlaunchgads.Store{Db: pool}.Select)},
		{"dmvertical", smoke(dmvertical.Store{Db: pool}.Select)},

		{"ecbcurr", smoke(ecbcurr.Store{Db: pool}.Select)},
		{"ecbcurrmd", smoke(ecbcurrmd.Store{Db: pool}.Select)},
		{"ecbxrperfnorm", smoke(ecbxrperfnorm.Store{Db: pool}.Select)},

		{"gemapicall", smoke(gemapicall.Store{Db: pool}.Select)},

		{"geocountry", smoke(geocountry.Store{Db: pool}.Select)},
		{"geoocean", smoke(geoocean.Store{Db: pool}.Select)},

		{"lysincemployee", smoke(lysincemployee.Store{Db: pool}.Select)},
		{"lysinceventsinmonth", smoke(lysinceventsinmonth.Store{Db: pool}.Select, "2026-06-01")},

		{"procflow", smoke(procflow.Store{Db: pool}.Select)},
		{"procpoint", smoke(procpoint.Store{Db: pool}.Select)},
		{"procrun", smoke(procrun.Store{Db: pool}.Select)},
		{"procstep", smoke(procstep.Store{Db: pool}.Select)},
		{"procsteplink", smoke(procsteplink.Store{Db: pool}.Select)},

		{"pubauthor", smoke(pubauthor.Store{Db: pool}.Select)},
		{"pubauthorarch", smoke(pubauthorarch.Store{Db: pool}.Select)},
		{"pubbook", smoke(pubbook.Store{Db: pool}.Select)},
		{"pubbookarch", smoke(pubbookarch.Store{Db: pool}.Select)},

		{"suppcompany", smoke(suppcompany.Store{Db: pool}.Select)},
		{"suppemployee", smoke(suppemployee.Store{Db: pool}.Select)},
		{"suppprodcategory", smoke(suppprodcategory.Store{Db: pool}.Select)},
		{"suppproduct", smoke(suppproduct.Store{Db: pool}.Select)},

		{"sysblockedip", smoke(sysblockedip.Store{Db: pool}.Select)},
		{"sysextdatasync", smoke(sysextdatasync.Store{Db: pool}.Select)},
		// session and login attempt rows are inserted by the test rather than seeded: see withRows
		{"sysloginattempt", withRows(pool,
			func(ctx context.Context, tx pgx.Tx) error {
				_, err := loginAttemptStore.BulkInsertTx(ctx, tx, []sysloginattempt.Input{{
					CreatedAt:   lystype.Datetime(time.Now()),
					Ip:          netip.MustParseAddr(smokeLoginAttempt),
					IsBlocked:   true,
					NumAttempts: 5,
				}})
				return err
			},
			fmt.Sprintf("DELETE FROM system.login_attempt WHERE ip = '%s'", smokeLoginAttempt),
			smoke(loginAttemptStore.Select))},
		{"sysnotification", smoke(sysnotification.Store{Db: pool}.Select)},
		{"sysrpcloginattempt", smoke(sysrpcloginattempt.Store{Db: pool}.Select)},
		{"sysrpcreq", smoke(sysrpcreq.Store{Db: pool}.Select)},
		{"sysrpcsession", smoke(sysrpcsession.Store{Db: pool}.Select)},
		{"sysrpcsessionhist", smoke(sysrpcsessionhist.Store{Db: pool}.Select)},
		{"syssession", withRows(pool,
			func(ctx context.Context, tx pgx.Tx) error {
				now := lystype.Datetime(time.Now())
				_, err := sessionStore.BulkInsertTx(ctx, tx, []syssession.Input{{
					AllowMultipleSessions: true,
					CreatedAt:             now,
					Email:                 "test@example.com",
					ExpiresAt:             lystype.Datetime(time.Now().Add(24 * time.Hour)),
					FamilyName:            "Doe",
					GivenName:             "John",
					GeoIpCountryIsoCode:   "US",
					GeoIpLocation:         "Test Location",
					Ip:                    netip.MustParseAddr("198.51.100.1"),
					LastAccessAt:          now,
					Roles:                 []string{"Viewer"},
					Token:                 smokeSessionToken,
					UserAgent:             "ua",
					UserFk:                1,
					UserName:              "test_user",
				}})
				return err
			},
			fmt.Sprintf("DELETE FROM system.session WHERE token = '%s'", smokeSessionToken),
			smoke(sessionStore.Select))},
		{"syssessionhist", smoke(syssessionhist.Store{Db: pool}.Select)},
		{"syssrvreq", smoke(syssrvreq.Store{Db: pool}.Select)},
		{"syssvcaccount", smoke(syssvcaccount.Store{Db: pool}.Select)},
		{"sysuser", smoke(sysuser.Store{Db: pool}.Select)},

		{"tedbvatratesumm", smoke(tedbvatratesumm.Store{Db: pool}.Select)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(t.Context()); err != nil {
				t.Errorf("Select failed: %v", err)
			}
		})
	}
}

// storeNames holds the schemaName, tableName, viewName and setFuncName constants declared by a store package
type storeNames struct {
	schema, table, view, setFunc string
}

// parseStoreNames returns the name constants of each store package under the current directory, keyed by package path
func parseStoreNames(t *testing.T) map[string]storeNames {
	t.Helper()

	res := make(map[string]storeNames)
	fset := token.NewFileSet()

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		pkgPath := filepath.ToSlash(filepath.Dir(path))
		names := res[pkgPath]

		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, id := range vs.Names {
					if i >= len(vs.Values) {
						continue
					}
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					val, err := strconv.Unquote(lit.Value)
					if err != nil {
						return err
					}
					switch id.Name {
					case "schemaName":
						names.schema = val
					case "tableName":
						names.table = val
					case "viewName":
						names.view = val
					case "setFuncName":
						names.setFunc = val
					}
				}
			}
		}

		if names != (storeNames{}) {
			res[pkgPath] = names
		}
		return nil
	})
	if err != nil {
		t.Fatalf("parsing store packages: %v", err)
	}

	return res
}

// dbNameSet runs a query returning "schema.name" strings and returns them as a set
func dbNameSet(t *testing.T, pool *pgxpool.Pool, query string) map[string]bool {
	t.Helper()

	rows, err := pool.Query(t.Context(), query)
	if err != nil {
		t.Fatalf("querying names: %v", err)
	}
	defer rows.Close()

	res := make(map[string]bool)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			t.Fatalf("scanning name: %v", err)
		}
		res[name] = true
	}
	if err = rows.Err(); err != nil {
		t.Fatalf("reading names: %v", err)
	}

	return res
}

// TestStoreNames checks that the schemaName, tableName, viewName and setFuncName constants of every store package
// refer to an existing table, view or function in the database.
func TestStoreNames(t *testing.T) {
	conf := testutil.LoadConfig(t)
	pool := testutil.NewOwnerPool(t, conf)

	// information_schema.tables omits materialized views, so use pg_class
	relations := dbNameSet(t, pool, `
		SELECT n.nspname || '.' || c.relname
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind IN ('r', 'p', 'v', 'm', 'f')
		AND n.nspname NOT IN ('pg_catalog', 'information_schema')`)
	funcs := dbNameSet(t, pool, `
		SELECT n.nspname || '.' || p.proname
		FROM pg_proc p
		JOIN pg_namespace n ON n.oid = p.pronamespace
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')`)

	storeNamesByPkg := parseStoreNames(t)
	if len(storeNamesByPkg) == 0 {
		t.Fatal("no store packages found")
	}

	for pkgPath, names := range storeNamesByPkg {
		t.Run(pkgPath, func(t *testing.T) {
			if names.schema == "" {
				t.Fatal("schemaName not declared")
			}
			if names.table == "" && names.view == "" && names.setFunc == "" {
				t.Fatal("none of tableName, viewName or setFuncName declared")
			}
			checks := []struct {
				kind, name string
				existing   map[string]bool
			}{
				{"tableName", names.table, relations},
				{"viewName", names.view, relations},
				{"setFuncName", names.setFunc, funcs},
			}
			for _, c := range checks {
				if c.name != "" && !c.existing[names.schema+"."+c.name] {
					t.Errorf("%s %q not found in schema %q", c.kind, c.name, names.schema)
				}
			}
		})
	}
}
