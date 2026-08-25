package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRunLoop_RunsImmediatelyAndOnInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	calls := 0
	done := make(chan struct{})

	run := func(context.Context) error {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 2 {
			close(done)
		}
		return nil
	}

	go runLoop(ctx, 10*time.Millisecond, run, func(error) {})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second run")
	}

	mu.Lock()
	defer mu.Unlock()
	if calls < 2 {
		t.Fatalf("expected at least 2 calls, got %d", calls)
	}
}

func TestRunLoop_ReportsErrorFromRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	run := func(context.Context) error { return errors.New("boom") }
	onError := func(err error) { errCh <- err }

	go runLoop(ctx, time.Hour, run, onError)

	select {
	case err := <-errCh:
		if err.Error() != "boom" {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for error callback")
	}
}

func TestRunLoop_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var mu sync.Mutex
	calls := 0
	run := func(context.Context) error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	}

	loopDone := make(chan struct{})
	go func() {
		runLoop(ctx, time.Hour, run, func(error) {})
		close(loopDone)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-loopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for runLoop to return after cancel")
	}

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected exactly 1 immediate call before cancel, got %d", calls)
	}
}
