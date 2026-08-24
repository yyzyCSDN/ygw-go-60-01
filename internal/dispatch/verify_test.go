package dispatch_test

import (
	"context"
	"io"
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

type trackingBody struct {
	io.ReadCloser
	closed chan<- bool
}

func (b *trackingBody) Close() error {
	b.closed <- true
	return b.ReadCloser.Close()
}

type recordingTransport struct {
	inner  http.RoundTripper
	closed chan<- bool
}

func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.inner.RoundTrip(req)
	if err == nil {
		resp.Body = &trackingBody{ReadCloser: resp.Body, closed: rt.closed}
	}
	return resp, err
}

func TestHTTPClientClosedAfterCallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	closed := make(chan bool, 4)
	registry := route.NewRegistry()
	cb := model.NewCallback("cb-test", "order.created", server.URL, "secret")
	if err := registry.Register(cb); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	eventQueue := queue.NewQueue()
	offsets := offset.NewStore()
	dedupStore := dedup.NewStore()
	deadStore := deadletter.NewStore()
	source := clock.NewSource(clock.SystemClock{})
	scheduler := retry.NewScheduler(retry.NewPolicy(time.Second, 30*time.Second, 3), source)
	client := dispatch.NewCallbackClientWithTransport(&recordingTransport{inner: http.DefaultTransport, closed: closed})
	dispatcher, err := dispatch.NewDispatcher(dispatch.Config{
		Registry:   registry,
		Queue:      eventQueue,
		Offsets:    offsets,
		Dedup:      dedupStore,
		Retry:      scheduler,
		DeadLetter: deadStore,
		Clock:      source,
		Client:     client,
		Timeout:    2 * time.Second,
		Secret:     "secret",
		DedupTTL:   time.Minute,
	})
	if err != nil {
		t.Fatalf("create dispatcher failed: %v", err)
	}
	event := model.NewEvent("evt-1", "order.created", []byte(`{"a":1}`))
	if _, err := eventQueue.Enqueue(event); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if err := dispatcher.DeliverTo(context.Background(), event, "cb-test"); err != nil {
		t.Fatalf("delivery failed: %v", err)
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("callback response body was never closed")
	}
}
