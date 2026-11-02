package main

import (
	"context"
	"fmt"

	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/launchsvc"
	"github.com/loveyourstack/lys/lyspgdb"
)

type dbListenerApplication struct {
	*cmd.Application
}

// newDbListenerApplication builds the application and connects everything it depends on.
// On error, anything already opened is closed. On success, the caller must call Close.
func newDbListenerApplication(ctx context.Context, conf *myapp.Config) (lisApp *dbListenerApplication, err error) {

	lisApp = &dbListenerApplication{
		Application: cmd.NewApplication(conf),
	}

	defer func() {
		if err != nil {
			lisApp.Close()
			lisApp = nil
		}
	}()

	lisApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbLisUser, conf.General.AppName+" Lis", myapp.TypesToRegister)
	if err != nil {
		return nil, fmt.Errorf("lyspgdb.GetPoolWithTypes failed: %w", err)
	}

	lisApp.LaunchSvc, err = launchsvc.NewService(lisApp.Db, conf.Reflis.PrepBatchSize, lisApp.Logger)
	if err != nil {
		return nil, fmt.Errorf("launchsvc.NewService failed: %w", err)
	}

	return lisApp, nil
}

// Close releases the db pool. It is safe to call on a partially built app.
func (lisApp *dbListenerApplication) Close() {
	if lisApp.Db != nil {
		lisApp.Db.Close()
	}
}
