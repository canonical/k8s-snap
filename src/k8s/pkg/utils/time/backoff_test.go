package time

import (
	"testing"
	"time"
)

func TestLinearBackoff(t *testing.T) {
	randInt64N = func(n int64) int64 { return n }

	tests := []struct {
		name      string
		attempt   int
		baseDelay time.Duration
		maxDelay  time.Duration
		expected  time.Duration
	}{
		{
			name:      "zero attempt",
			attempt:   0,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  1 * time.Second,
			expected:  150 * time.Millisecond,
		},
		{
			name:      "first attempt",
			attempt:   1,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  1 * time.Second,
			expected:  150 * time.Millisecond,
		},
		{
			name:      "second attempt",
			attempt:   2,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  1 * time.Second,
			expected:  300 * time.Millisecond,
		},
		{
			name:      "attempt exceeds max delay",
			attempt:   30,
			baseDelay: 100 * time.Millisecond,
			maxDelay:  1 * time.Second,
			expected:  1500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if LinearBackoff(tt.attempt, tt.baseDelay, tt.maxDelay) != tt.expected {
				t.Errorf("LinearBackoff(attempt=%d, baseDelay=%v, maxDelay=%v) = %v, want %v",
					tt.attempt, tt.baseDelay, tt.maxDelay, LinearBackoff(tt.attempt, tt.baseDelay, tt.maxDelay), tt.expected)
			}
		})
	}
}
