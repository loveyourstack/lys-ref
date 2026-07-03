package launchsvc

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

var errQueueEmpty = errors.New("queue empty")

func (svc Service) RunFbProcessing(ctx context.Context, workerCount int) (err error) {

	if workerCount <= 0 {
		return fmt.Errorf("workerCount must be > 0, but is %d", workerCount)
	}

	// create a context that can be canceled to stop all workers
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// use an atomic boolean to signal workers to stop when the queue is empty
	var stopOnEmpty atomic.Bool

	// create an errgroup to manage the workers
	g, gctx := errgroup.WithContext(workerCtx)

	for i := range workerCount {
		g.Go(func() error {
			for {
				// exit if queue is empty
				if stopOnEmpty.Load() {
					svc.Logger.Debug("worker stopping due to empty queue atomic", "id", i)
					return nil
				}

				// exit on context cancellation
				if gctx.Err() != nil {
					svc.Logger.Debug("worker stopping due to context cancellation", "id", i)
					return nil
				}

				// process one item
				svc.Logger.Debug("worker trying to process next item", "id", i)
				err := svc.processFbWorker(gctx)
				if err == nil {
					continue
				}

				// if queue is empty, signal other workers to stop and exit normally
				if errors.Is(err, errQueueEmpty) {
					svc.Logger.Debug("worker stopping due to empty queue", "id", i)
					stopOnEmpty.Store(true)
					return nil
				}

				// hard error: stop all workers
				cancel()
				return err
			}
		})
	}

	// wait for all workers to finish
	if err := g.Wait(); err != nil {
		return fmt.Errorf("RunFbProcessing failed: %w", err)
	}

	return nil
}

func (svc Service) processFbWorker(ctx context.Context) (err error) {

	// claim next queued item
	item, err := svc.FbLaunchStore.ClaimNextForProcessing(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errQueueEmpty
		}
		return fmt.Errorf("svc.FbLaunchStore.ClaimNextForProcessing failed: %w", err)
	}

	// process item
	if err = svc.ProcessFbLaunchItem(ctx, item); err != nil {
		// hard error only; API/inconsistency errors should already be handled in ProcessFbLaunchItem
		return fmt.Errorf("svc.ProcessFbLaunchItem failed: %w", err)
	}

	return nil
}

// -------------------------------------------------------------------------------------------------------------------

func (svc Service) RunGAdsProcessing(ctx context.Context, workerCount int) (err error) {

	if workerCount <= 0 {
		return fmt.Errorf("workerCount must be > 0, but is %d", workerCount)
	}

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var stopOnEmpty atomic.Bool
	g, gctx := errgroup.WithContext(workerCtx)

	for range workerCount {
		g.Go(func() error {
			for {
				if stopOnEmpty.Load() {
					return nil
				}

				if gctx.Err() != nil {
					return nil
				}

				err := svc.processGAdsWorker(gctx)
				if err == nil {
					continue
				}

				if errors.Is(err, errQueueEmpty) {
					stopOnEmpty.Store(true)
					return nil
				}

				cancel()
				return err
			}
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("RunGAdsProcessing failed: %w", err)
	}

	return nil
}

func (svc Service) processGAdsWorker(ctx context.Context) (err error) {

	// claim next queued item
	item, err := svc.GAdsLaunchStore.ClaimNextForProcessing(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errQueueEmpty
		}
		return fmt.Errorf("svc.GAdsLaunchStore.ClaimNextForProcessing failed: %w", err)
	}

	// process item
	if err = svc.ProcessGAdsLaunchItem(ctx, item); err != nil {
		return fmt.Errorf("svc.ProcessGAdsLaunchItem failed: %w", err)
	}

	return nil
}
