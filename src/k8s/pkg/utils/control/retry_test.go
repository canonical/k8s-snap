package control

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryFor(t *testing.T) {
	t.Run("Retry succeeds", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		retryCount := 3
		count := 0

		err := RetryFor(ctx, retryCount, 50*time.Millisecond, func() error {
			count++
			if count < retryCount {
				return errors.New("failed")
			}
			return nil
		})
		if err != nil {
			t.Errorf("Expected nil error, got: %v", err)
		}
		if count != retryCount {
			t.Errorf("Expected retry count %d, got: %d", retryCount, count)
		}
	})

	t.Run("Retry fails with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		retryCount := 3

		err := RetryFor(ctx, retryCount, time.Second, func() error {
			time.Sleep(200 * time.Millisecond)
			return errors.New("failed")
		})

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	})

	t.Run("Retry exhausts without success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		retryCount := 3

		err := RetryFor(ctx, retryCount, 100*time.Millisecond, func() error {
			return errors.New("failed")
		})

		if err == nil {
			t.Error("Expected non-nil error, got nil")
		}
	})
}

func TestRetryForIf(t *testing.T) {
	nonRetryable := errors.New("not retryable")
	retryable := errors.New("retryable")
	isRetryable := func(err error) bool { return errors.Is(err, retryable) }

	t.Run("Retry succeeds on retryable errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		retryCount := 3
		count := 0

		err := RetryForIf(ctx, retryCount, 50*time.Millisecond, isRetryable, func() error {
			count++
			if count < retryCount {
				return retryable
			}
			return nil
		})
		if err != nil {
			t.Errorf("Expected nil error, got: %v", err)
		}
		if count != retryCount {
			t.Errorf("Expected retry count %d, got: %d", retryCount, count)
		}
	})

	t.Run("Returns immediately on a non-retryable error", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		count := 0

		err := RetryForIf(ctx, 5, time.Second, isRetryable, func() error {
			count++
			return nonRetryable
		})

		if !errors.Is(err, nonRetryable) {
			t.Errorf("Expected nonRetryable error, got: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected retryFunc to be called once, got: %d", count)
		}
	})

	t.Run("Retry fails with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := RetryForIf(ctx, 3, time.Second, isRetryable, func() error {
			time.Sleep(200 * time.Millisecond)
			return retryable
		})

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	})

	t.Run("Retry exhausts without success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := RetryForIf(ctx, 3, 100*time.Millisecond, isRetryable, func() error {
			return retryable
		})

		if !errors.Is(err, retryable) {
			t.Errorf("Expected retryable error, got: %v", err)
		}
	})
}
