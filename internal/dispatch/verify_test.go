package dispatch_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hookrelay/internal/clock"
	"hookrelay/internal/deadletter"
	"hookrelay/internal/dedup"
	"hookrelay/internal/dispatch"
	"hookrelay/internal/model"
	"hookrelay/internal/offset"
	"hookrelay/internal/queue"
	"hookrelay/internal/retry"
	"hookrelay/internal/route"
)

func newTestDispatcher(t *testing.T, targetURL string) (*dispatch.Dispatcher, *route.Registry, *queue.Queue, *offset.Store) {
	t.Helper()
	registry := route.NewRegistry()
	cb := model.NewCallback("cb-test", "order.created", targetURL, "secret")
	if err := registry.Register(cb); err != nil {
		t.Fatalf("register callback failed: %v", err)
	}
	eventQueue := queue.NewQueue()
	offsets := offset.NewStore()
	dedupStore := dedup.NewStore()
	deadStore := deadletter.NewStore()
	source := clock.NewSource(clock.SystemClock{})
	scheduler := retry.NewScheduler(retry.NewPolicy(time.Second, 30*time.Second, 3), source)
	client := dispatch.NewCallbackClient(5 * time.Second)
	dispatcher, err := dispatch.NewDispatcher(dispatch.Config{
		Registry:   registry,
		Queue:      eventQueue,
		Offsets:    offsets,
		Dedup:      dedupStore,
		Retry:      scheduler,
		DeadLetter: deadStore,
		Clock:      source,
		Client:     client,
		Timeout:    50 * time.Millisecond,
		Secret:     "secret",
		DedupTTL:   time.Minute,
	})
	if err != nil {
		t.Fatalf("create dispatcher failed: %v", err)
	}
	return dispatcher, registry, eventQueue, offsets
}

func TestEmptyPayloadNoNilPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	dispatcher, _, eventQueue, offsets := newTestDispatcher(t, server.URL)
	event := model.NewEvent("evt-empty", "order.created", nil)
	if _, err := eventQueue.Enqueue(event); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if err := dispatcher.DeliverTo(context.Background(), event, "cb-test"); err != nil {
		t.Fatalf("empty payload delivery must not panic: %v", err)
	}
	if got := offsets.Current("cb-test"); got != event.Seq {
		t.Fatalf("empty payload delivery should still advance the offset, got %d", got)
	}
}
