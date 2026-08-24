package dispatch_test

import (
	"testing"
	"time"

	"hookrelay/internal/clock"
	"hookrelay/internal/dispatch"
)

func TestSignatureUsesFreshTimestamp(t *testing.T) {
	start := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	manual := clock.NewManualClock(start)
	signer := dispatch.NewSigner(clock.NewSource(manual), "secret")
	first, firstStamp := signer.Sign([]byte(`{"a":1}`))
	manual.Advance(2 * time.Hour)
	second, secondStamp := signer.Sign([]byte(`{"a":1}`))
	if firstStamp.Equal(secondStamp) {
		t.Fatalf("signatures must use a fresh timestamp, both %v", firstStamp)
	}
	if !secondStamp.Equal(start.Add(2 * time.Hour)) {
		t.Fatalf("second signature should use the advanced clock, got %v", secondStamp)
	}
	if first == second {
		t.Fatal("signature must change when the timestamp changes")
	}
}
