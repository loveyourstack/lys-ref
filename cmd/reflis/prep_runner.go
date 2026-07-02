package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// preparationRunner is a struct that manages the execution of a preparation function.
// It ensures that the preparation function is not run concurrently and handles debouncing of triggers.
type preparationRunner struct {
	ctx         context.Context
	logger      *slog.Logger
	runPrepFunc func(ctx context.Context) error

	mu      sync.Mutex // guards pending and running
	pending bool
	running bool

	wg sync.WaitGroup // ensures that the preparation function can be waited on for graceful shutdown
}

func newPreparationRunner(ctx context.Context, runPrepFunc func(ctx context.Context) error, logger *slog.Logger) *preparationRunner {
	return &preparationRunner{
		ctx:         ctx,
		logger:      logger,
		runPrepFunc: runPrepFunc,
	}
}

// trigger sets the pending flag and starts the preparation function in a goroutine if it's not already running.
func (r *preparationRunner) trigger() {

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

			// run the prep function
			if err := r.runPrepFunc(r.ctx); err != nil && r.ctx.Err() == nil {
				r.logger.Error("r.runPrepFunc failed", "error", err)
			}
		}
	})
}

// wait waits for the preparation runner's goroutine to finish.
func (r *preparationRunner) wait() {
	r.wg.Wait()
}
