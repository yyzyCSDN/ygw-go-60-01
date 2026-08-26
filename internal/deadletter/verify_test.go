package deadletter_test

import (
	"context"
	"errors"
	"testing"

	"hookrelay/internal/deadletter"
	"hookrelay/internal/model"
)

type failingSink struct{}

func (failingSink) Save(_ context.Context, _ *model.DeadLetter) error {
	return errors.New("storage full")
}

func TestDeadLetterWriteErrorNotSwallowed(t *testing.T) {
	store := deadletter.NewStore()
	letter := model.NewDeadLetter("evt-1", "cb-1", "boom", []byte(`{}`))
	err := store.WriteWithSink(context.Background(), letter, failingSink{})
	if err == nil {
		t.Fatal("dead letter write failure must propagate to the caller")
	}
	if store.Len() != 0 {
		t.Fatalf("failed write must not leave a record, got %d", store.Len())
	}
}
