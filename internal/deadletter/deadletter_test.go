package deadletter

import (
	"context"
	"testing"

	"hookrelay/internal/model"
)

func TestWriteGetList(t *testing.T) {
	store := NewStore()
	letter := model.NewDeadLetter("evt-1", "cb-1", "exhausted", []byte(`{}`))
	if err := store.Write(letter); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	got, ok := store.Get(letter.ID)
	if !ok || got.EventID != "evt-1" {
		t.Fatal("dead letter should be retrievable")
	}
	if store.Len() != 1 || len(store.List()) != 1 {
		t.Fatalf("expected one dead letter, got len=%d list=%d", store.Len(), len(store.List()))
	}
}

func TestWriteRejectsDuplicate(t *testing.T) {
	store := NewStore()
	letter := model.NewDeadLetter("evt-1", "cb-1", "exhausted", nil)
	if err := store.Write(letter); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if err := store.Write(letter); err == nil {
		t.Fatal("duplicate dead letter id should be rejected")
	}
}

func TestWriteRejectsEmptyID(t *testing.T) {
	store := NewStore()
	letter := &model.DeadLetter{}
	if err := store.Write(letter); err == nil {
		t.Fatal("empty dead letter should be rejected")
	}
}

func TestReplayResolveFlow(t *testing.T) {
	store := NewStore()
	letter := model.NewDeadLetter("evt-1", "cb-1", "boom", []byte(`{}`))
	_ = store.Write(letter)
	if err := store.MarkReplaying(letter.ID); err != nil {
		t.Fatalf("mark replaying failed: %v", err)
	}
	if err := store.Resolve(letter.ID); err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	resolved, _ := store.Get(letter.ID)
	if resolved.Status != model.DeadLetterResolved || resolved.ResolvedAt == nil {
		t.Fatal("resolved letter should carry resolved status and time")
	}
}

func TestMemorySinkWritesToStore(t *testing.T) {
	store := NewStore()
	sink := NewMemorySink()
	letter := model.NewDeadLetter("evt-2", "cb-2", "boom", nil)
	if err := store.WriteWithSink(context.Background(), letter, sink); err != nil {
		t.Fatalf("memory sink write failed: %v", err)
	}
	if err := store.WriteWithSink(context.Background(), letter, sink); err == nil {
		t.Fatal("duplicate write through sink should be rejected")
	}
}
