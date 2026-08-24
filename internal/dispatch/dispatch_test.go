package dispatch

import (
	"errors"
	"testing"
	"time"
)

func TestResultSuccess(t *testing.T) {
	ok := &Result{StatusCode: 200, Body: []byte(`{}`)}
	if !ok.Success() {
		t.Fatal("2xx result should be successful")
	}
	bad := &Result{StatusCode: 500}
	if bad.Success() {
		t.Fatal("5xx result must not be successful")
	}
	errored := &Result{Err: errors.New("boom")}
	if errored.Success() {
		t.Fatal("errored result must not be successful")
	}
}

func TestResultErrorText(t *testing.T) {
	if got := (&Result{Err: errors.New("boom")}).Error(); got != "boom" {
		t.Fatalf("expected boom, got %q", got)
	}
	if got := (&Result{StatusCode: 429}).Error(); got != "callback returned status 429" {
		t.Fatalf("unexpected status error text: %q", got)
	}
}

func TestResultBodyTextTruncates(t *testing.T) {
	long := make([]byte, 512)
	result := &Result{Body: long}
	if len(result.BodyText()) != 256 {
		t.Fatalf("body text should truncate to 256 bytes, got %d", len(result.BodyText()))
	}
	if (&Result{}).BodyText() != "" {
		t.Fatal("empty body should yield empty text")
	}
}

func TestVerifySignatureRoundTrip(t *testing.T) {
	body := []byte(`{"sku":"a"}`)
	secret := "shared-secret"
	mac := hmac256(secret, body, "1754000000")
	if err := VerifySignature(secret, body, "1754000000", mac); err != nil {
		t.Fatalf("valid signature should verify: %v", err)
	}
	if err := VerifySignature(secret, body, "1754000001", mac); err == nil {
		t.Fatal("tampered timestamp should fail verification")
	}
}

func TestHeaderNamesAndTimestampFormat(t *testing.T) {
	tsHeader, sigHeader := HeaderNames()
	if tsHeader != "X-Hook-Timestamp" || sigHeader != "X-Hook-Signature" {
		t.Fatalf("unexpected header names: %q %q", tsHeader, sigHeader)
	}
	now := time.Unix(1754000000, 0)
	if FormatTimestamp(now) != "1754000000" {
		t.Fatalf("unexpected timestamp format: %q", FormatTimestamp(now))
	}
}
