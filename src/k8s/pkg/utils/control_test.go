package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock check function that returns true after 2 iterations.
func mockCheckFunc() func() bool {
	counter := 0
	return func() bool {
		counter++
		return counter > 1
	}
}

func TestWaitUntilReady(t *testing.T) {
	// Test case 1: Successful completion
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel1()

	err1 := WaitUntilReady(ctx1, mockCheckFunc(), time.Second, "test error message")
	if err1 != nil {
		t.Errorf("Expected no error, got: %v", err1)
	}

	// Test case 2: Context cancellation
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2() // Cancel the context immediately

	err2 := WaitUntilReady(ctx2, mockCheckFunc(), time.Second, "test error message")
	if err2 == nil || err2 != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err2)
	}

	// Test case 3: Timeout
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel3()

	err3 := WaitUntilReady(ctx3, func() bool { return false }, time.Second, "test error message")
	if err3 == nil || !errors.Is(err3, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err3)
	}
}
