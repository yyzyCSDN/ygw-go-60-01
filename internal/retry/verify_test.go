package retry_test

import (
	"testing"
	"time"

	"hookrelay/internal/retry"
)

func TestRetryAppliesBackoff(t *testing.T) {
	policy := retry.NewPolicy(time.Second, 30*time.Second, 5)
	first := policy.NextDelay(1)
	second := policy.NextDelay(2)
	if first <= 0 {
		t.Fatalf("first retry must apply a positive backoff, got %v", first)
	}
	if second <= first {
		t.Fatalf("backoff must increase between attempts, got %v then %v", first, second)
	}
	if policy.NextDelay(0) != 0 {
		t.Fatal("attempt zero must not produce backoff")
	}
}
