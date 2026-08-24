package offset_test

import (
	"testing"

	"hookrelay/internal/offset"
)

func TestDeliveryAckAdvancesOffset(t *testing.T) {
	store := offset.NewStore()
	if err := store.Acknowledge("cb-order", 7); err != nil {
		t.Fatalf("ack failed: %v", err)
	}
	if got := store.Current("cb-order"); got != 7 {
		t.Fatalf("expected offset 7 after ack, got %d", got)
	}
	snap := store.Snapshot()
	if len(snap) != 1 || snap[0].CallbackID != "cb-order" || snap[0].Sequence != 7 {
		t.Fatalf("snapshot must record the acknowledged offset, got %+v", snap)
	}
}
