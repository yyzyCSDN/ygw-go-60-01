package model

import (
	"bytes"
	"testing"
	"time"
)

func TestNewEventGeneratesStableID(t *testing.T) {
	first := NewEvent("", "order.created", []byte(`{"sku":"a"}`))
	second := NewEvent("", "order.created", []byte(`{"sku":"a"}`))
	if first.ID == "" || first.ID != second.ID {
		t.Fatalf("content-derived event id should be stable, got %q vs %q", first.ID, second.ID)
	}
}

func TestEventBodyCopiesPayload(t *testing.T) {
	event := NewEvent("evt-1", "audit.trail", []byte(`{"ok":true}`))
	body := event.Body()
	if !bytes.Equal(body, event.Payload) {
		t.Fatalf("Body should contain the same bytes as payload, got %q vs %q", body, event.Payload)
	}
}

func TestEventBodyEmptyPayload(t *testing.T) {
	event := NewEvent("evt-empty", "audit.trail", nil)
	body := event.Body()
	if len(body) != 0 {
		t.Fatalf("empty payload body should be empty, got %q", body)
	}
}

func TestCallbackValidateRejectsBadURL(t *testing.T) {
	cb := NewCallback("cb-1", "order.created", "ftp://host/hook", "s")
	if err := cb.Validate(); err == nil {
		t.Fatal("ftp scheme should be rejected")
	}
}

func TestCallbackMatchesEnabledOnly(t *testing.T) {
	cb := NewCallback("cb-1", "order.created", "http://host/hook", "s")
	if !cb.Matches("order.created") {
		t.Fatal("enabled callback should match its event type")
	}
	cb.Enabled = false
	if cb.Matches("order.created") {
		t.Fatal("disabled callback must not match")
	}
}

func TestDeliveryTaskTransitions(t *testing.T) {
	now := time.Now().UTC()
	task := NewTask("evt-1", "cb-1")
	task.MarkDelivering(now)
	if task.Status != TaskDelivering {
		t.Fatalf("expected delivering, got %s", task.Status)
	}
	task.MarkDelivered(now)
	if task.Status != TaskDelivered {
		t.Fatalf("expected delivered, got %s", task.Status)
	}
	task.MarkRetrying(now.Add(time.Minute), "boom", now)
	if task.Attempt != 1 || !task.DueAt(now.Add(time.Minute)) {
		t.Fatal("retrying task should become due at its next attempt time")
	}
	task.MarkDead("exhausted", now)
	if task.Status != TaskDead {
		t.Fatalf("expected dead, got %s", task.Status)
	}
}

func TestDeadLetterStatusFlow(t *testing.T) {
	letter := NewDeadLetter("evt-1", "cb-1", "timeout", []byte(`{}`))
	if letter.Status != DeadLetterPending {
		t.Fatalf("new dead letter should be pending, got %s", letter.Status)
	}
	if letter.Reason != "timeout" {
		t.Fatalf("reason not preserved: %q", letter.Reason)
	}
}

func TestOffsetSnapshotRoundTrip(t *testing.T) {
	snap := NewOffsetSnapshot("cb-1", 12)
	if snap.CallbackID != "cb-1" || snap.Sequence != 12 {
		t.Fatalf("snapshot fields not preserved: %+v", snap)
	}
}
