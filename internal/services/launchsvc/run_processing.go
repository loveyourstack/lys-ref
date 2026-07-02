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

	for range workerCount {
		g.Go(func() error {
			for {
				// exit if queue is empty
				if stopOnEmpty.Load() {
					return nil
				}

				// exit on context cancellation
				if gctx.Err() != nil {
					return nil
				}

				// process one item
				err := svc.processFbWorker(gctx)
				if err == nil {
					continue
				}

				// if queue is empty, signal other workers to stop and exit normally
				if errors.Is(err, errQueueEmpty) {
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

	// begin tx
	tx, err := svc.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("svc.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// select next queued item
	item, err := svc.FbLaunchStore.SelectNextQueuedTx(ctx, tx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errQueueEmpty
		}
		return fmt.Errorf("svc.FbLaunchStore.SelectNextQueuedTx failed: %w", err)
	}

	// process item
	if err = svc.ProcessFbLaunchItem(ctx, tx, item); err != nil {
		// hard error only; API/inconsistency errors should already be handled in ProcessFbLaunchItem
		return fmt.Errorf("svc.ProcessFbLaunchItem failed: %w", err)
	}

	// commit tx
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit failed: %w", err)
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

	// begin tx
	tx, err := svc.Db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("svc.Db.Begin failed: %w", err)
	}
	defer tx.Rollback(ctx)

	// select next queued item
	item, err := svc.GAdsLaunchStore.SelectNextQueuedTx(ctx, tx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return errQueueEmpty
		}
		return fmt.Errorf("svc.GAdsLaunchStore.SelectNextQueuedTx failed: %w", err)
	}

	// process item
	if err = svc.ProcessGAdsLaunchItem(ctx, tx, item); err != nil {
		return fmt.Errorf("svc.ProcessGAdsLaunchItem failed: %w", err)
	}

	// commit tx
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx.Commit failed: %w", err)
	}

	return nil
}
