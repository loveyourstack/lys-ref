package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys-ref/cmd"
	"github.com/loveyourstack/lys-ref/internal/myapp"
	"github.com/loveyourstack/lys-ref/internal/services/launchsvc"
	"github.com/loveyourstack/lys/lyspgdb"
)

type dbListenerApplication struct {
	*cmd.Application
	// ... additional fields specific to db listener
}

func main() {

	configFileName := "ref_config.toml"

	// mandatory flag if not using default
	configFilePath := flag.String("configFilePath", configFileName, "Path to the config file")

	flag.Parse()

	// load config from file
	conf := myapp.Config{}
	err := conf.LoadFromFile(*configFilePath)
	if err != nil {
		log.Fatalf("initialization: conf.LoadFromFile (%s) failed: %s", *configFilePath, err.Error())
	}

	// create context that listens for interrupt and termination signals to allow for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// create non-specific app
	app := cmd.NewApplication(&conf)

	// create db listener app
	lisApp := &dbListenerApplication{
		Application: app,
	}

	// connect to db and assign pool to srvApp
	lisApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbLisUser, lisApp.Config.General.AppName+" Lis", myapp.TypesToRegister)
	if err != nil {
		log.Fatalf("initialization: failed to create regular db connection pool: %s", err.Error())
	}
	defer lisApp.Db.Close()

	// attach services
	lisApp.LaunchSvc = launchsvc.NewService(lisApp.Db, lisApp.Logger)

	// acquire a connection from the pool for listening
	conn, err := lisApp.Db.Acquire(ctx)
	if err != nil {
		log.Fatalf("initialization: failed to acquire connection: %s", err.Error())
	}
	defer conn.Release()

	// listen to the channel name defined in system.notify_change_trigger()
	pgChanName := "change"
	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", pgx.Identifier{pgChanName}.Sanitize()))
	if err != nil {
		log.Fatalf("initialization: failed to listen to '%s' channel", pgChanName)
	}

	// display startup message
	lisApp.Logger.Info(fmt.Sprintf("listening for events on pg channel: %s", pgChanName))

	// wait for pg_notify events
	for {
		not, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			// exit normally on ctx cancellation
			if ctx.Err() != nil {
				return
			}
			lisApp.Logger.Error("conn.Conn().WaitForNotification failed", "error", err)
			os.Exit(1)
		}

		// parse payload
		changeP := changePayload{}
		err = json.Unmarshal([]byte(not.Payload), &changeP)
		if err != nil {
			lisApp.Logger.Error("json.Unmarshal failed", "error", err)
			continue
		}

		lisApp.Logger.Debug("event", "action", changeP.Action, "schema", changeP.Schema, "table", changeP.Table)

		switch changeP.Schema {
		case "digmark":
			switch changeP.Table {
			case "launcher_fb":
				switch changeP.Action {
				case "insert":
					go func() {
						if err := lisApp.LaunchSvc.RunFbPreparation(ctx); err != nil {
							lisApp.Logger.Error("lisApp.LaunchSvc.RunFbPreparation failed", "error", err)
						}
					}()
				case "update":
					go func() {
						if err := lisApp.LaunchSvc.RunFbPreparation(ctx); err != nil {
							lisApp.Logger.Error("lisApp.LaunchSvc.RunFbPreparation failed", "error", err)
						}
					}()

					// TODO - run queued items
				default:
					lisApp.Logger.Error("no handler for action", "schema", changeP.Schema, "table", changeP.Table, "action", changeP.Action)
				}
			case "launcher_gads":
				switch changeP.Action {
				case "insert":
					go func() {
						if err := lisApp.LaunchSvc.RunGAdsPreparation(ctx); err != nil {
							lisApp.Logger.Error("lisApp.LaunchSvc.RunGAdsPreparation failed", "error", err)
						}
					}()
				case "update":
					go func() {
						if err := lisApp.LaunchSvc.RunGAdsPreparation(ctx); err != nil {
							lisApp.Logger.Error("lisApp.LaunchSvc.RunGAdsPreparation failed", "error", err)
						}
					}()

					// TODO - run queued items
				default:
					lisApp.Logger.Error("no handler for action", "schema", changeP.Schema, "table", changeP.Table, "action", changeP.Action)
				}
			default:
				lisApp.Logger.Error("no handler for table", "schema", changeP.Schema, "table", changeP.Table)
			}
		default:
			lisApp.Logger.Error("no handler for schema", "schema", changeP.Schema)
		}

	}
}

// changePayload matches the payload defined in system.notify_change_trigger()
type changePayload struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Schema    string    `json:"schema"`
	Table     string    `json:"table"`
}
