package procsvc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/runstatus"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procpoint"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procrun"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procstep"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysexec"
	"github.com/loveyourstack/lys/lystype"
)

func (svc Service) CreateRunFromStep(ctx context.Context, db *pgxpool.Pool, stepId int64, kvPairs []string, withDeps bool) (runId int64, err error) {

	// create replacementMap from kvPairs
	replacementMap, err := svc.getReplacementMap(kvPairs)
	if err != nil {
		return 0, fmt.Errorf("svc.getReplacementMap failed: %w", err)
	}

	// select this step
	stepStore := procstep.Store{Db: db}
	step, err := stepStore.SelectById(ctx, stepId)
	if err != nil {
		return 0, fmt.Errorf("stepStore.SelectById failed: %w", err)
	}

	// execute cmd placeholder replacements
	cmd := step.Cmd
	for placeholder, replacement := range replacementMap {
		cmd = strings.ReplaceAll(cmd, ":"+placeholder, placeholder+"="+replacement)
	}

	// exit if cmd still contains unreplaced placeholders
	if strings.Contains(cmd, ":") {
		return 0, lyserr.User{Message: fmt.Sprintf("cmd contains unreplaced placeholders: %s", cmd)}
	}

	var depIds []int64
	var depSteps []procstep.Model

	if withDeps {

		// get dependant step ids, if any
		depIds, err = stepStore.SelectDepIds(ctx, stepId)
		if err != nil {
			return 0, fmt.Errorf("stepStore.SelectDepIds failed: %w", err)
		}

		// select dependants
		depSteps, err = stepStore.SelectByIds(ctx, depIds)
		if err != nil {
			return 0, fmt.Errorf("stepStore.SelectByIds failed: %w", err)
		}
	}

	// begin tx
	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// create run
	runInput := procrun.Input{
		FlowFk:   step.FlowFk,
		StepId:   step.Id,
		StepName: step.Name,
	}
	runId, err = procrun.InsertTx(ctx, tx, runInput)
	if err != nil {
		return 0, fmt.Errorf("procrun.InsertTx failed: %w", err)
	}

	// create point inputs for this step and its deps

	// remove dependsOn if deps are not used
	dependsOn := step.DependsOn
	if !withDeps {
		dependsOn = []int64{}
	}

	pointInputs := []procpoint.Input{
		{
			Cmd:          cmd,
			DependsOn:    dependsOn,
			DisplayOrder: step.DisplayOrder,
			ErrMsg:       "",
			FinishedAt:   lystype.Datetime{},
			RunFk:        runId,
			StartedAt:    lystype.Datetime{},
			Status:       runstatus.Waiting,
			StepId:       step.Id,
			StepName:     step.Name,
		},
	}

	if withDeps {

		for _, depStep := range depSteps {

			// execute cmd placeholder replacements
			cmd := depStep.Cmd
			for placeholder, replacement := range replacementMap {
				cmd = strings.ReplaceAll(cmd, ":"+placeholder, placeholder+"="+replacement)
			}

			pointInputs = append(pointInputs,
				procpoint.Input{
					Cmd:          cmd,
					DependsOn:    depStep.DependsOn, // refers to step ids: need to be changed after this to point ids
					DisplayOrder: depStep.DisplayOrder,
					ErrMsg:       "",
					FinishedAt:   lystype.Datetime{},
					RunFk:        runId,
					StartedAt:    lystype.Datetime{},
					Status:       runstatus.Waiting,
					StepId:       depStep.Id,
					StepName:     depStep.Name,
				},
			)
		}
	}

	// insert points
	_, err = procpoint.BulkInsertTx(ctx, tx, pointInputs)
	if err != nil {
		return 0, fmt.Errorf("procpoint.BulkInsertTx failed: %w", err)
	}

	// replace depends_on ids of new points
	if withDeps {
		numUpdated, err := procpoint.UpdateDependsOnIdsTx(ctx, tx, runId)
		if err != nil {
			return 0, fmt.Errorf("procpoint.UpdateDependsOnIdsTx failed: %w", err)
		}
		_ = numUpdated
		//fmt.Println("numUpdated", numUpdated)
	}

	// success: commit tx
	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("tx.Commit failed: %w", err)
	}

	return runId, nil
}

func (svc Service) doPoint(ctx context.Context, pointStore procpoint.Store, pointId int64, cmdParts []string) (err error) {

	fullCmd := strings.Join(cmdParts, " ")

	// set startedAt amd status to Running
	err = pointStore.SetRunning(ctx, pointId)
	if err != nil {
		return fmt.Errorf("pointStore.SetRunning failed for cmd: %s: %w", fullCmd, err)
	}

	// run cmd
	runOpts := lysexec.RunOptions{}
	runRes, err := lysexec.Run(ctx, cmdParts[0], runOpts, cmdParts[1:]...)
	if err != nil {

		// if cancelled via context (e.g. via errgroup), set to Interrupted, otherwise set to Error with errMsg
		// don't use run's context
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
			err2 := pointStore.SetInterrupted(context.Background(), pointId)
			if err2 != nil {
				return fmt.Errorf("pointStore.SetInterrupted failed for cmd: %s: %w", fullCmd, err2)
			}
		} else {
			// don't set full error since it contains path. Cut down to just the error message.
			_, errMsg, found := strings.Cut(err.Error(), " cliApp.ProcSvc.")
			if !found {
				errMsg = "error message not in expected format"
			}

			err2 := pointStore.SetError(context.Background(), errMsg, pointId)
			if err2 != nil {
				return fmt.Errorf("pointStore.SetError failed for cmd: %s: %w", fullCmd, err2)
			}
		}

		// on context cancellation, runRes.Canceled will be true
		svc.Logger.Debug("failure", slog.String("cmd", fullCmd), slog.Any("result", runRes))

		return fmt.Errorf("lysexec.Run failed for cmd: %s: %w", fullCmd, err)
	}

	//svc.Logger.Debug("success", slog.String("cmd", fullCmd), slog.Any("result", runRes))

	// set finishedAt and status to Completed
	err = pointStore.SetCompleted(ctx, pointId)
	if err != nil {
		return fmt.Errorf("pointStore.SetCompleted failed for cmd: %s: %w", fullCmd, err)
	}

	return nil
}

func (svc Service) getReplacementMap(kvPairs []string) (replacementMap map[string]string, err error) {

	replacementMap = make(map[string]string)
	for _, kvPair := range kvPairs {
		kV := strings.Split(kvPair, "=")
		if len(kV) != 2 {
			return nil, fmt.Errorf("key/value pair not in format k=v: %s", kvPair)
		}
		replacementMap[kV[0]] = kV[1]
	}

	return replacementMap, nil
}

func (svc Service) RunFakeCmd(ctx context.Context, name string, sleepSecs int, logger *slog.Logger, args ...string) error {

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	logger.Debug("started", slog.String("name", name), slog.String("args", strings.Join(args, " ")))

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range sleepSecs {
		select {
		case <-ctx.Done(): // allow context cancellation
			return ctx.Err()
		case <-ticker.C:
			if rng.Intn(20) == 0 { // 5% chance of error each second

				switch rng.Intn(3) {
				case 0:
					return fmt.Errorf("fake application error")
				case 1:
					return lyserr.Db{Err: fmt.Errorf("fake database error"), Stmt: "fake stmt"}
				case 2:
					return lyserr.Ext{Err: fmt.Errorf("fake external API error"), Message: "fake external message"}
				}
			}
		}
	}

	logger.Debug("finished", slog.String("name", name), slog.String("args", strings.Join(args, " ")))

	return nil
}
