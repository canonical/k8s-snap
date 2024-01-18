package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Mock check function that returns true after 2 iterations.
func mockCheckFunc() (bool, error) {
	return true, nil
}

var testError = errors.New("test error")

// Mock check function that returns an error.
func mockErrorCheckFunc() (bool, error) {
	return false, testError
}

func TestWaitUntilReady(t *testing.T) {
	// Test case 1: Successful completion
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel1()

	err1 := WaitUntilReady(ctx1, mockCheckFunc)
	if err1 != nil {
		t.Errorf("Expected no error, got: %v", err1)
	}

	// Test case 2: Context cancellation
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2() // Cancel the context immediately

	err2 := WaitUntilReady(ctx2, mockCheckFunc)
	if err2 == nil || err2 != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err2)
	}

	// Test case 3: Timeout
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel3()

	err3 := WaitUntilReady(ctx3, func() (bool, error) { return false, nil })
	if err3 == nil || !errors.Is(err3, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err3)
	}

	// Test case 4: CheckFunc returns an error
	ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel4()

	err4 := WaitUntilReady(ctx4, mockErrorCheckFunc)
	if err4 == nil || !errors.Is(err4, testError) {
		t.Errorf("Expected test error, got: %v", err4)
	}
}
