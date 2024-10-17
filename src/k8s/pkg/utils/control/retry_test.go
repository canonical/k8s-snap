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
