package offset

import "testing"

func TestNewStoreStartsEmpty(t *testing.T) {
	store := NewStore()
	if store.CallbackCount() != 0 {
		t.Fatalf("new store should have no callbacks, got %d", store.CallbackCount())
	}
	if len(store.All()) != 0 {
		t.Fatal("new store should have no offsets")
	}
	if len(store.Snapshot()) != 0 {
		t.Fatal("new store should have no snapshots")
	}
}

func TestCurrentDefaultsToZero(t *testing.T) {
	store := NewStore()
	if store.Current("cb-unknown") != 0 {
		t.Fatal("unknown callback should default to offset zero")
	}
}

func TestSnapshotIndependentCopies(t *testing.T) {
	store := NewStore()
	snap := store.Snapshot()
	if len(snap) != 0 {
		t.Fatal("snapshot should be empty for fresh store")
	}
	all := store.All()
	all["injected"] = 9
	if store.CallbackCount() != 0 {
		t.Fatal("mutating All copy must not affect store")
	}
}
