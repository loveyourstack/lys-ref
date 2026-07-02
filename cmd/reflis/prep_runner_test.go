package main

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestPreparationRunner_TriggerWhileRunning_CoalescesToOneExtraRun(t *testing.T) {
	ctx := t.Context()

	allowFirstRunCh := make(chan struct{})
	firstRunStartedCh := make(chan struct{})
	secondRunStartedCh := make(chan struct{})

	var callCount atomic.Int32
	runner := newPreparationRunner(ctx, func(ctx context.Context) error {
		n := callCount.Add(1)
		switch n {
		case 1:
			close(firstRunStartedCh)
			<-allowFirstRunCh
		case 2:
			close(secondRunStartedCh)
		}
		return nil
	}, testLogger())

	runner.trigger()

	select {
	case <-firstRunStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first run to start")
	}

	// While run 1 is in-flight, many triggers should still collapse into one extra run.
	for range 10 {
		runner.trigger()
	}

	close(allowFirstRunCh)

	select {
	case <-secondRunStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second run to start")
	}

	time.Sleep(350 * time.Millisecond)

	if got := callCount.Load(); got != 2 {
		t.Fatalf("expected exactly 2 runs (coalesced), got %d", got)
	}
}

func TestPreparationRunner_Wait_BlocksUntilRunCompletes(t *testing.T) {
	ctx := t.Context()

	releaseRunCh := make(chan struct{})
	runStartedCh := make(chan struct{})

	runner := newPreparationRunner(ctx, func(ctx context.Context) error {
		close(runStartedCh)
		<-releaseRunCh
		return nil
	}, testLogger())

	runner.trigger()

	select {
	case <-runStartedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for run to start")
	}

	waitDoneCh := make(chan struct{})
	go func() {
		runner.wait()
		close(waitDoneCh)
	}()

	select {
	case <-waitDoneCh:
		t.Fatal("wait returned before run completed")
	case <-time.After(200 * time.Millisecond):
	}

	close(releaseRunCh)

	select {
	case <-waitDoneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("wait did not return after run completed")
	}
}

func TestPreparationRunner_TriggerAfterCancel_DoesNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var callCount atomic.Int32
	runner := newPreparationRunner(ctx, func(ctx context.Context) error {
		callCount.Add(1)
		return nil
	}, testLogger())

	runner.trigger()

	time.Sleep(200 * time.Millisecond)
	runner.wait()

	if got := callCount.Load(); got != 0 {
		t.Fatalf("expected 0 runs after canceled context, got %d", got)
	}
}
