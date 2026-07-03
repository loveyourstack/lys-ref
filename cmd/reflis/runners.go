package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/loveyourstack/lys/lysexec"
)

func newPreparationRunner(
	ctx context.Context,
	runPreparationFunc func(context.Context) error,
	logger *slog.Logger,
) *lysexec.CoalescingRunner {
	return lysexec.NewCoalescingRunner(ctx, runPreparationFunc, "runPreparationFunc", 100*time.Millisecond, logger)
}

func newProcessingRunner(
	ctx context.Context,
	runProcessingFunc func(context.Context, int) error,
	logger *slog.Logger,
	workerCount int,
) *lysexec.CoalescingRunner {
	if workerCount <= 0 {
		log.Fatalf("workerCount must be > 0, but is %d", workerCount)
	}
	return lysexec.NewCoalescingRunner(ctx,
		func(ctx context.Context) error {
			return runProcessingFunc(ctx, workerCount)
		}, "runProcessingFunc", 100*time.Millisecond, logger)

}
