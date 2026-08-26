package dedup_test

import (
	"testing"
	"time"

	"hookrelay/internal/dedup"
)

func TestDedupKeyRemovedAfterDelivery(t *testing.T) {
	store := dedup.NewStore()
	key := store.Mark("cb-order", 9, time.Minute)
	if !store.Check(key) {
		t.Fatal("freshly marked key should be present")
	}
	store.PruneThrough("cb-order", 9)
	if store.Check(key) {
		t.Fatal("dedup key must be removed once delivery passes its sequence")
	}
	if store.Len() != 0 {
		t.Fatalf("expected all keys pruned, got %d", store.Len())
	}
}
