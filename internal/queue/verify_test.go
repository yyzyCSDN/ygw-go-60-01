package queue_test

import (
	"fmt"
	"testing"

	"hookrelay/internal/model"
	"hookrelay/internal/queue"
)

func TestBatchWindowNoOverlap(t *testing.T) {
	q := queue.NewQueue()
	for i := 1; i <= 10; i++ {
		event := model.NewEvent(fmt.Sprintf("evt-%d", i), "order.created", []byte(`{}`))
		if _, err := q.Enqueue(event); err != nil {
			t.Fatalf("enqueue %d failed: %v", i, err)
		}
	}
	first := q.Batch(0, 5)
	second := q.Batch(5, 5)
	if len(first) != 5 || len(second) != 5 {
		t.Fatalf("expected two windows of five events, got %d and %d", len(first), len(second))
	}
	seen := make(map[string]bool)
	for _, event := range first {
		if seen[event.ID] {
			t.Fatalf("duplicate event inside first window: %s", event.ID)
		}
		seen[event.ID] = true
	}
	for _, event := range second {
		if seen[event.ID] {
			t.Fatalf("event %s appears in both windows", event.ID)
		}
		seen[event.ID] = true
	}
}
