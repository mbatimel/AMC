package main

import (
	"context"
	"time"
)

// runLoop выполняет run немедленно, затем повторяет по interval,
// пока ctx не отменят. Каждая ошибка run (включая первый немедленный
// вызов) передаётся в onError; сам цикл при этом не останавливается.
func runLoop(ctx context.Context, interval time.Duration, run func(context.Context) error, onError func(error)) {
	if err := run(ctx); err != nil {
		onError(err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := run(ctx); err != nil {
				onError(err)
			}
		}
	}
}
