package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
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
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {

	configFileName := "ref_config.toml"

	// mandatory flag if not using default
	configFilePath := flag.String("configFilePath", configFileName, "Path to the config file")

	flag.Parse()

	// load config from file
	conf := myapp.Config{}
	err := conf.LoadFromFile(*configFilePath)
	if err != nil {
		return fmt.Errorf("initialization: conf.LoadFromFile (%s) failed: %w", *configFilePath, err)
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

	// connect to db and assign pool to lisApp
	lisApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbLisUser, lisApp.Config.General.AppName+" Lis", myapp.TypesToRegister)
	if err != nil {
		return fmt.Errorf("initialization: failed to create db connection pool: %w", err)
	}
	defer lisApp.Db.Close()

	// attach services
	lisApp.LaunchSvc = launchsvc.NewService(lisApp.Db, lisApp.Logger)

	// create prep runners
	fbPrepRunner := newPreparationRunner(ctx, lisApp.LaunchSvc.RunFbPreparation, lisApp.Logger)
	gadsPrepRunner := newPreparationRunner(ctx, lisApp.LaunchSvc.RunGAdsPreparation, lisApp.Logger)

	// acquire a connection from the pool for listening
	conn, err := lisApp.Db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("initialization: failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// listen to the channel name defined in system.notify_change_trigger()
	pgChanName := "change"
	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", pgx.Identifier{pgChanName}.Sanitize()))
	if err != nil {
		return fmt.Errorf("initialization: failed to listen to '%s' channel: %w", pgChanName, err)
	}

	// trigger prep runners once at startup in case listener was down when changes were made
	fbPrepRunner.trigger()
	gadsPrepRunner.trigger()

	// display startup message
	lisApp.Logger.Info(fmt.Sprintf("listening for events on pg channel: %s", pgChanName))

	// wait for notifications and handle them
	err = waitForNotifications(ctx, conn.Conn(), fbPrepRunner, gadsPrepRunner, lisApp.Logger)

	// wait for runners to complete before exit
	fbPrepRunner.wait()
	gadsPrepRunner.wait()

	if err != nil {
		lisApp.Logger.Error("reflis shutdown: waitForNotifications failed", "error", err)
		return err
	}

	lisApp.Logger.Info("reflis shutdown signal received")
	return nil
}

func waitForNotifications(ctx context.Context, conn *pgx.Conn, fbPrepRunner, gadsPrepRunner *preparationRunner,
	logger *slog.Logger) (err error) {

	// wait for pg_notify events
	for {
		not, err := conn.WaitForNotification(ctx)
		if err != nil {
			// break without err on ctx cancellation
			if ctx.Err() != nil {
				break
			}
			// return error on other failures
			return fmt.Errorf("conn.WaitForNotification failed: %w", err)
		}

		// parse payload
		changeP := changePayload{}
		err = json.Unmarshal([]byte(not.Payload), &changeP)
		if err != nil {
			// log and continue
			logger.Error("json.Unmarshal failed", "error", err)
			continue
		}

		logger.Debug("event", "action", changeP.Action, "schema", changeP.Schema, "table", changeP.Table)

		// dispatch event based on schema, table and action
		switch changeP.Schema {
		case "digmark":
			switch changeP.Table {
			case "launcher_fb":
				switch changeP.Action {
				case "insert":
					fbPrepRunner.trigger()
				case "update":
					fbPrepRunner.trigger()

					// TODO - run queued items
				default:
					logger.Error("no handler for action", "schema", changeP.Schema, "table", changeP.Table, "action", changeP.Action)
				}
			case "launcher_gads":
				switch changeP.Action {
				case "insert":
					gadsPrepRunner.trigger()
				case "update":
					gadsPrepRunner.trigger()

					// TODO - run queued items
				default:
					logger.Error("no handler for action", "schema", changeP.Schema, "table", changeP.Table, "action", changeP.Action)
				}
			default:
				logger.Error("no handler for table", "schema", changeP.Schema, "table", changeP.Table)
			}
		default:
			logger.Error("no handler for schema", "schema", changeP.Schema)
		}
	}

	return nil
}

// changePayload matches the payload defined in system.notify_change_trigger()
type changePayload struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Schema    string    `json:"schema"`
	Table     string    `json:"table"`
}
