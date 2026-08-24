package clock

import (
	"testing"
	"time"
)

func TestManualClockSetAndAdvance(t *testing.T) {
	start := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	mc := NewManualClock(start)
	if !mc.Now().Equal(start) {
		t.Fatalf("manual clock should start at %v, got %v", start, mc.Now())
	}
	mc.Advance(5 * time.Second)
	expected := start.Add(5 * time.Second)
	if !mc.Now().Equal(expected) {
		t.Fatalf("manual clock should advance to %v, got %v", expected, mc.Now())
	}
	mc.Set(expected.Add(time.Hour))
	if !mc.Now().Equal(expected.Add(time.Hour)) {
		t.Fatal("manual clock should honor Set")
	}
}

func TestSystemClockReturnsRecentTime(t *testing.T) {
	before := time.Now().UTC()
	got := (SystemClock{}).Now()
	if got.Before(before.Add(-time.Minute)) || got.After(time.Now().UTC().Add(time.Minute)) {
		t.Fatalf("system clock returned implausible time: %v", got)
	}
}

func TestSourceDelegatesToClock(t *testing.T) {
	start := time.Date(2026, 8, 23, 9, 0, 0, 0, time.UTC)
	mc := NewManualClock(start)
	source := NewSource(mc)
	if !source.Now().Equal(start) {
		t.Fatalf("source should delegate to manual clock, got %v", source.Now())
	}
}
