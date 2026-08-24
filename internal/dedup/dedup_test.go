package dedup

import (
	"testing"
	"time"
)

func TestMarkCheckAndExpire(t *testing.T) {
	store := NewStore()
	key := store.Mark("cb-1", 1, time.Minute)
	if !store.Check(key) {
		t.Fatal("freshly marked key should be checkable")
	}
	if store.Len() != 1 {
		t.Fatalf("expected one key, got %d", store.Len())
	}
}

func TestCheckUnknownKey(t *testing.T) {
	store := NewStore()
	if store.Check("dk-missing") {
		t.Fatal("unknown key must not be checkable")
	}
}

func TestRemoveDeletesKey(t *testing.T) {
	store := NewStore()
	key := store.Mark("cb-1", 1, time.Minute)
	store.Remove(key)
	if store.Check(key) {
		t.Fatal("removed key must not be checkable")
	}
	if store.Len() != 0 {
		t.Fatalf("expected empty store after remove, got %d", store.Len())
	}
}

func TestCleanExpiredRemovesOldKeys(t *testing.T) {
	store := NewStore()
	key := store.Mark("cb-1", 1, time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	removed := store.CleanExpired(time.Now().UTC().Add(time.Minute))
	if removed != 1 {
		t.Fatalf("expected one expired key removed, got %d", removed)
	}
	if store.Check(key) {
		t.Fatal("expired key should not be checkable")
	}
}

func TestDedupKeyStable(t *testing.T) {
	first := DedupKey("cb-1", 7)
	second := DedupKey("cb-1", 7)
	if first != second || first == "" {
		t.Fatalf("dedup key should be stable, got %q vs %q", first, second)
	}
	if DedupKey("cb-1", 7) == DedupKey("cb-1", 8) {
		t.Fatal("different sequences must produce different keys")
	}
}
