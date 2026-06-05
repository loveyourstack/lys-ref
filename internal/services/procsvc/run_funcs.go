package procsvc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys-ref/internal/enums/runstatus"
	"github.com/loveyourstack/lys-ref/internal/stores/process/procpoint"
	"github.com/loveyourstack/lys/lysslice"
	"golang.org/x/sync/errgroup"
)

// RunOnly runs the step without its dependencies.
func (svc Service) RunOnly(ctx context.Context, db *pgxpool.Pool, runId int64) (err error) {

	// select point
	pointStore := procpoint.Store{Db: db}
	statusMap, err := pointStore.SelectStatusMapByRunId(ctx, runId)
	if err != nil {
		return fmt.Errorf("pointStore.SelectStatusMapByRunId failed: %w", err)
	}
	if len(statusMap[runstatus.Waiting]) != 1 {
		return fmt.Errorf("len(statusMap[runstatus.Waiting]) is not 1")
	}
	point := statusMap[runstatus.Waiting][0]

	// prefix and split cmd
	cmdParts, err := svc.prefixAndSplitCmd(point.Cmd)
	if err != nil {
		return fmt.Errorf("svc.prefixAndSplitCmd failed for cmd: %s: %w", point.Cmd, err)
	}

	err = svc.doPoint(ctx, pointStore, point.Id, cmdParts)
	if err != nil {
		return fmt.Errorf("svc.doPoint failed: %w", err)
	}

	return nil
}

// MustRunWithDeps runs the step with its dependencies using ErrGroup. It stops on the first error.
func (svc Service) MustRunWithDeps(ctx context.Context, db *pgxpool.Pool, runId int64) (err error) {

	pointStore := procpoint.Store{Db: db}

	iter := 0
	for {

		iter++
		svc.InfoLog.Debug("point loop", slog.Int("iter", iter))

		statusMap, err := pointStore.SelectStatusMapByRunId(ctx, runId)
		if err != nil {
			return fmt.Errorf("pointStore.SelectStatusMapByRunId failed: %w", err)
		}

		// break if no more Waiting points
		if len(statusMap[runstatus.Waiting]) == 0 {
			break
		}

		// get slice of Completed ids
		completedIds := make([]int64, len(statusMap[runstatus.Completed]))
		for i, item := range statusMap[runstatus.Completed] {
			completedIds[i] = item.Id
		}

		// will exit and cancel all running goroutines started by g.Go on any error
		g, ctx := errgroup.WithContext(ctx)

		for _, point := range statusMap[runstatus.Waiting] {

			// prefix and split cmd
			cmdParts, err := svc.prefixAndSplitCmd(point.Cmd)
			if err != nil {
				return fmt.Errorf("svc.prefixAndSplitCmd failed for cmd: %s: %w", point.Cmd, err)
			}

			// if point has dependencies
			if len(point.DependsOn) > 0 {

				// skip point if any of its dependencies are not Completed
				if !lysslice.ContainsAll(completedIds, point.DependsOn) {
					continue
				}
			}

			g.Go(func() error {
				return svc.doPoint(ctx, pointStore, point.Id, cmdParts)
			})
		}

		if err := g.Wait(); err != nil {

			// set any waiting points as cancelled (don't use errgroup's ctx)
			cancelErr := pointStore.CancelByRunId(context.Background(), runId)
			if cancelErr != nil {
				return fmt.Errorf("pointStore.CancelByRunId failed: %w", cancelErr)
			}

			return fmt.Errorf("doPoint failed: %w", err)
		}

	} // end for

	return nil
}

// RunWithDeps runs the step with its dependencies using WaitGroup. It continues on errors.
func (svc Service) RunWithDeps(ctx context.Context, db *pgxpool.Pool, runId int64) (err error) {

	pointStore := procpoint.Store{Db: db}

	iter := 0
	for {

		iter++
		svc.InfoLog.Debug("point loop", slog.Int("iter", iter))

		statusMap, err := pointStore.SelectStatusMapByRunId(ctx, runId)
		if err != nil {
			return fmt.Errorf("pointStore.SelectStatusMapByRunId failed: %w", err)
		}

		// break if no more Waiting points
		if len(statusMap[runstatus.Waiting]) == 0 {
			break
		}

		// get slice of Completed and Error ids
		finishedIds := []int64{}
		for _, item := range statusMap[runstatus.Completed] {
			finishedIds = append(finishedIds, item.Id)
		}
		for _, item := range statusMap[runstatus.Error] {
			finishedIds = append(finishedIds, item.Id)
		}

		var wg sync.WaitGroup

		for _, point := range statusMap[runstatus.Waiting] {

			// prefix and split cmd
			cmdParts, err := svc.prefixAndSplitCmd(point.Cmd)
			if err != nil {
				return fmt.Errorf("svc.prefixAndSplitCmd failed for cmd: %s: %w", point.Cmd, err)
			}

			// if point has dependencies
			if len(point.DependsOn) > 0 {

				// skip point if any of its dependencies are not finished
				if !lysslice.ContainsAll(finishedIds, point.DependsOn) {
					continue
				}
			}

			wg.Add(1)
			go func(ctx context.Context, pointStore procpoint.Store, pointId int64, cmdParts []string, infoLog *slog.Logger) {

				defer wg.Done()

				err := svc.doPoint(ctx, pointStore, pointId, cmdParts)
				if err != nil {
					infoLog.Error("svc.doPoint error on cmd", slog.String("cmd", strings.Join(cmdParts, " ")), slog.String("error", err.Error()))
				}

			}(ctx, pointStore, point.Id, cmdParts, svc.InfoLog)
		}

		wg.Wait()

	} // end for

	return nil
}
