package main

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"
)

type processingRunner struct {
	ctx               context.Context
	logger            *slog.Logger
	runProcessingFunc func(ctx context.Context, workerCount int) error
	workerCount       int

	mu      sync.Mutex // guards pending and running
	pending bool
	running bool

	wg sync.WaitGroup // ensures that the preparation function can be waited on for graceful shutdown
}

func newProcessingRunner(ctx context.Context, runProcessingFunc func(ctx context.Context, workerCount int) error,
	logger *slog.Logger, workerCount int) *processingRunner {

	if workerCount <= 0 {
		log.Fatalf("workerCount must be > 0, but is %d", workerCount)
	}

	return &processingRunner{
		ctx:               ctx,
		logger:            logger,
		runProcessingFunc: runProcessingFunc,
		workerCount:       workerCount,
	}
}

// trigger sets the pending flag and starts the processing function in a goroutine if it's not already running.
func (r *processingRunner) trigger() {

	// ensure no action is taken if the context is already canceled
	if r.ctx.Err() != nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// set pending true to indicate that another run is needed
	r.pending = true

	// exit if already running
	if r.running {
		return
	}

	r.running = true

	// start goroutine in wg so that graceful shutdown is possible using r.wait()
	r.wg.Go(func() {

		// run while pending is true, with a 100ms debounce delay between runs
		for {
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			// exit on context cancellation
			case <-r.ctx.Done():
				timer.Stop()

				r.mu.Lock()
				r.running = false
				r.mu.Unlock()

				return
			case <-timer.C:
			}

			// exit if no pending runs
			r.mu.Lock()
			if !r.pending {
				r.running = false
				r.mu.Unlock()
				return
			}

			r.pending = false
			r.mu.Unlock()

			// run the processing function
			if err := r.runProcessingFunc(r.ctx, r.workerCount); err != nil && r.ctx.Err() == nil {
				r.logger.Error("r.runProcessingFunc failed", "error", err)
			}
		}
	})
}

// wait waits for the processing runner's goroutine to finish.
func (r *processingRunner) wait() {
	r.wg.Wait()
}
