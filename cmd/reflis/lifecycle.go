package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lysexec"
)

func (lisApp *dbListenerApplication) serve(ctx context.Context, stop context.CancelFunc) error {

	// create prep runners
	fbPrepRunner := newPreparationRunner(ctx, lisApp.LaunchSvc.RunFbPreparation, lisApp.Logger)
	gadsPrepRunner := newPreparationRunner(ctx, lisApp.LaunchSvc.RunGAdsPreparation, lisApp.Logger)

	// create processing runners
	fbProcessingRunner := newProcessingRunner(ctx, lisApp.LaunchSvc.RunFbProcessing, lisApp.Logger, lisApp.Config.Reflis.WorkerCount)
	gadsProcessingRunner := newProcessingRunner(ctx, lisApp.LaunchSvc.RunGAdsProcessing, lisApp.Logger, lisApp.Config.Reflis.WorkerCount)

	// acquire a connection from the pool for listening
	conn, err := lisApp.Db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("lisApp.Db.Acquire failed: %w", err)
	}
	defer conn.Release()

	// listen to the channel name defined in system.notify_change_trigger()
	pgChanName := "change"
	_, err = conn.Exec(ctx, fmt.Sprintf("LISTEN %s;", pgx.Identifier{pgChanName}.Sanitize()))
	if err != nil {
		return fmt.Errorf("conn.Exec (LISTEN) failed: channel '%s': %w", pgChanName, err)
	}

	// trigger runners once at startup in case listener was down when changes were made
	fbPrepRunner.Trigger()
	gadsPrepRunner.Trigger()

	fbProcessingRunner.Trigger()
	gadsProcessingRunner.Trigger()

	// display startup message
	lisApp.Logger.Info(fmt.Sprintf("listening for events on pg channel: %s", pgChanName))

	// listen for notifications and handle them
	listenErr := listen(ctx, conn.Conn(), fbPrepRunner, gadsPrepRunner, fbProcessingRunner, gadsProcessingRunner, lisApp.Logger)

	// cancel runners on every exit path (no-op if already canceled by a signal)
	stop()

	// wait for runners to exit, bounded so a runFunc that ignores ctx can't hang shutdown
	runnersDoneCh := make(chan struct{})
	go func() {
		defer close(runnersDoneCh)
		fbPrepRunner.Wait()
		gadsPrepRunner.Wait()
		fbProcessingRunner.Wait()
		gadsProcessingRunner.Wait()
	}()

	runErr := listenErr
	select {
	case <-runnersDoneCh:
	case <-time.After(10 * time.Second):
		runErr = errors.Join(runErr, errors.New("shutdown: timeout waiting for runners to exit"))
	}

	if listenErr == nil {
		lisApp.Logger.Info("shutdown signal received")
	}

	return runErr
}

func listen(ctx context.Context, conn *pgx.Conn,
	fbPrepRunner, gadsPrepRunner, fbProcessingRunner, gadsProcessingRunner *lysexec.CoalescingRunner,
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
					fbPrepRunner.Trigger()
				case "update":
					fbPrepRunner.Trigger()
					fbProcessingRunner.Trigger()
				default:
					logger.Error("no handler for action", "schema", changeP.Schema, "table", changeP.Table, "action", changeP.Action)
				}
			case "launcher_gads":
				switch changeP.Action {
				case "insert":
					gadsPrepRunner.Trigger()
				case "update":
					gadsPrepRunner.Trigger()
					gadsProcessingRunner.Trigger()
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
