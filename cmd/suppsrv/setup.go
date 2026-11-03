package main

import (
	"context"
	"fmt"
	"time"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/enums/appenv"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/stores/system/sysblockedip"
	"github.com/loveyourstack/lys-ref/internal/stores/system/syssrvreq"
	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lysauth"
	"github.com/loveyourstack/lys/lyspgdb"
)

// newBaseApplication builds the in-memory part of the application.
func newBaseApplication(conf *myapp.Config) (srvApp *httpServerApplication, err error) {

	srvApp = &httpServerApplication{
		Application: cmd.NewApplication(conf),
		PostOptions: lysapi.FillPostOptions(lysapi.PostOptions{}), // use defaults

		RateLimits: lysauth.NewAppRateLimits(2, 5, 1*time.Hour), // 2 rps with burst of 5

		UseRealIpHeader: conf.General.Env == appenv.Prod,
	}

	srvApp.GetOptions, err = lysapi.FillGetOptions(lysapi.GetOptions{}) // use defaults
	if err != nil {
		return nil, fmt.Errorf("lysapi.FillGetOptions failed: %w", err)
	}

	srvApp.LoginAttempts = lysauth.NewAppLoginAttempts(5) // block after 5 failed attempts

	return srvApp, nil
}

// newHttpServerApplication builds the application and connects everything it depends on.
// On error, anything already opened is closed. On success, the caller must call Close.
func newHttpServerApplication(ctx context.Context, conf *myapp.Config) (srvApp *httpServerApplication, err error) {

	// build the in-memory part of the application first
	srvApp, err = newBaseApplication(conf)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			srvApp.Close()
			srvApp = nil
		}
	}()

	// RLS connection (main connection for queries, with company_id setting applied for RLS)
	srvApp.Db, err = lyspgdb.GetPoolWithCtxSetting[int64](ctx, conf.Db, conf.DbServerUser, conf.General.AppName, "app.company_id", CompanyCtxKey, srvApp.Logger)
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPoolWithCtxSetting failed: %w", err)
	}

	// owner db connection for queries that need to bypass RLS
	srvApp.OwnerDb, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" (owner)")
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPool (owner) failed: %w", err)
	}

	srvApp.BlockedIPStore = sysblockedip.Store{Db: srvApp.OwnerDb}
	srvApp.SrvLogStore = syssrvreq.Store{Db: srvApp.OwnerDb}

	return srvApp, nil
}

// Close releases the database pools. It is safe to call on a partially built app.
func (srvApp *httpServerApplication) Close() {
	if srvApp.OwnerDb != nil {
		srvApp.OwnerDb.Close()
	}
	if srvApp.Db != nil {
		srvApp.Db.Close()
	}
}
