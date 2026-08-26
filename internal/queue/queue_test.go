package queue

import (
	"testing"

	"hookrelay/internal/model"
)

func TestEnqueueAssignsSequentialNumbers(t *testing.T) {
	q := NewQueue()
	first := model.NewEvent("evt-1", "order.created", []byte(`{}`))
	second := model.NewEvent("evt-2", "order.created", []byte(`{}`))
	seq1, err := q.Enqueue(first)
	if err != nil || seq1 != 1 {
		t.Fatalf("first enqueue should return seq 1, got %d (err=%v)", seq1, err)
	}
	seq2, err := q.Enqueue(second)
	if err != nil || seq2 != 2 {
		t.Fatalf("second enqueue should return seq 2, got %d (err=%v)", seq2, err)
	}
	if q.Len() != 2 {
		t.Fatalf("expected queue length 2, got %d", q.Len())
	}
}

func TestEnqueueRejectsDuplicateID(t *testing.T) {
	q := NewQueue()
	event := model.NewEvent("evt-1", "order.created", []byte(`{}`))
	_, _ = q.Enqueue(event)
	if _, err := q.Enqueue(model.NewEvent("evt-1", "order.created", []byte(`{}`))); err == nil {
		t.Fatal("duplicate event id should be rejected")
	}
}

func TestQueueLookups(t *testing.T) {
	q := NewQueue()
	event := model.NewEvent("evt-1", "audit.trail", []byte(`{}`))
	_, _ = q.Enqueue(event)
	byID, ok := q.GetByID("evt-1")
	if !ok || byID.ID != "evt-1" {
		t.Fatal("event should be found by id")
	}
}
