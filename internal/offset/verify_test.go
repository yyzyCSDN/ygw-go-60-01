package offset_test

import (
	"testing"

	"hookrelay/internal/offset"
)

func TestOffsetNoRegressOnRestart(t *testing.T) {
	before := offset.NewStore()
	if err := before.Acknowledge("cb-order", 42); err != nil {
		t.Fatalf("ack failed: %v", err)
	}
	snapshots := before.Snapshot()
	restarted := offset.NewStore()
	if err := restarted.Restore(snapshots); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
	if got := restarted.Current("cb-order"); got != 42 {
		t.Fatalf("restart must restore the latest offset, got %d", got)
	}
}
