package retry

import (
	"testing"
	"time"

	"hookrelay/internal/clock"
)

func TestPolicyTotalAttempts(t *testing.T) {
	policy := NewPolicy(time.Second, 30*time.Second, 3)
	if policy.TotalAttempts() != 3 {
		t.Fatalf("expected three total attempts, got %d", policy.TotalAttempts())
	}
}

func TestPolicyClampsAttempts(t *testing.T) {
	policy := NewPolicy(time.Second, 30*time.Second, 0)
	if policy.TotalAttempts() != 1 {
		t.Fatalf("zero attempts should clamp to one, got %d", policy.TotalAttempts())
	}
}

func TestSchedulerRetryBudget(t *testing.T) {
	start := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	source := clock.NewManualClock(start)
	policy := NewPolicy(time.Second, 30*time.Second, 3)
	scheduler := NewScheduler(policy, source)
	if !scheduler.ShouldRetry(1) || !scheduler.ShouldRetry(2) {
		t.Fatal("attempts below the budget should allow retry")
	}
	if !scheduler.Exhausted(3) {
		t.Fatal("attempts at the budget should be exhausted")
	}
}

func TestDeadlineReturnsBoundedContext(t *testing.T) {
	start := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	source := clock.NewManualClock(start)
	policy := NewPolicy(time.Second, 30*time.Second, 3)
	scheduler := NewScheduler(policy, source)
	ctx, cancel := scheduler.Deadline(nilContext(), 50*time.Millisecond)
	defer cancel()
	if ctx.Err() != nil {
		t.Fatal("fresh context should not be canceled")
	}
}
