package notify

import (
	"testing"
	"time"
)

func TestNewRateLimiter_InvalidWindow(t *testing.T) {
	_, err := NewRateLimiter(0, 3)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNewRateLimiter_InvalidMaxSends(t *testing.T) {
	_, err := NewRateLimiter(time.Minute, 0)
	if err == nil {
		t.Fatal("expected error for zero maxSends")
	}
}

func TestRateLimiter_AllowsUpToMax(t *testing.T) {
	rl, err := NewRateLimiter(time.Minute, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		if !rl.Allow("prod") {
			t.Fatalf("expected Allow to return true on attempt %d", i+1)
		}
	}

	if rl.Allow("prod") {
		t.Fatal("expected Allow to return false after max sends reached")
	}
}

func TestRateLimiter_IndependentPerEnv(t *testing.T) {
	rl, _ := NewRateLimiter(time.Minute, 2)

	rl.Allow("prod")
	rl.Allow("prod")

	if !rl.Allow("staging") {
		t.Fatal("staging should not be affected by prod rate limit")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl, _ := NewRateLimiter(time.Minute, 2)

	rl.Allow("prod")
	rl.Allow("prod")
	if rl.Allow("prod") {
		t.Fatal("expected prod to be rate limited")
	}

	rl.Reset("prod")
	if !rl.Allow("prod") {
		t.Fatal("expected prod to be allowed after reset")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	rl, _ := NewRateLimiter(time.Minute, 5)

	if r := rl.Remaining("prod"); r != 5 {
		t.Fatalf("expected 5 remaining, got %d", r)
	}

	rl.Allow("prod")
	rl.Allow("prod")

	if r := rl.Remaining("prod"); r != 3 {
		t.Fatalf("expected 3 remaining, got %d", r)
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl, _ := NewRateLimiter(50*time.Millisecond, 2)

	rl.Allow("prod")
	rl.Allow("prod")
	if rl.Allow("prod") {
		t.Fatal("expected prod to be rate limited before window expires")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("prod") {
		t.Fatal("expected prod to be allowed after window expired")
	}
}
