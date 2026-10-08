package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {

	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	err := ctx.Err()
	if err != nil {
		return nil, err
	}

	if len(in) == 0 {
		return []Result[R]{}, nil
	}

	result := make([]Result[R], len(in))

	countGoroutine := min(workers, len(in))

	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		indIn int
	)

	for range countGoroutine {

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				mu.Lock()
				if indIn >= len(in) || ctx.Err() != nil {
					mu.Unlock()
					return
				}
				localIndIn := indIn
				indIn++
				mu.Unlock()
				r, err := fn(ctx, in[localIndIn])

				if err != nil {
					result[localIndIn] = Result[R]{Err: err}
				} else {
					result[localIndIn] = Result[R]{r, err}
				}
			}
		}()
	}

	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return result, nil
}
