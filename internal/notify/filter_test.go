package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// captureSender records calls made to Send.
type captureSender struct {
	calls [][]drift.Drift
	err   error
}

func (c *captureSender) Send(_ string, drifts []drift.Drift) error {
	c.calls = append(c.calls, drifts)
	return c.err
}

func makeFilterDrifts(keys []string, severity string) []drift.Drift {
	var out []drift.Drift
	for _, k := range keys {
		out = append(out, drift.Drift{Key: k, Severity: severity})
	}
	return out
}

func TestFilter_NoDrifts_SkipsSend(t *testing.T) {
	cap := &captureSender{}
	f := NewFilter(cap, FilterConfig{})
	if err := f.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.calls) != 0 {
		t.Errorf("expected no calls, got %d", len(cap.calls))
	}
}

func TestFilter_MinSeverityCritical_FiltersWarnings(t *testing.T) {
	cap := &captureSender{}
	f := NewFilter(cap, FilterConfig{MinSeverity: "critical"})
	drifts := append(
		makeFilterDrifts([]string{"KEY_A"}, "warning"),
		makeFilterDrifts([]string{"KEY_B"}, "critical")...,
	)
	if err := f.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.calls) != 1 || len(cap.calls[0]) != 1 {
		t.Fatalf("expected 1 call with 1 drift, got %v", cap.calls)
	}
	if cap.calls[0][0].Key != "KEY_B" {
		t.Errorf("expected KEY_B, got %s", cap.calls[0][0].Key)
	}
}

func TestFilter_Cooldown_SuppressesRepeat(t *testing.T) {
	cap := &captureSender{}
	now := time.Now()
	f := NewFilter(cap, FilterConfig{Cooldown: 10 * time.Minute})
	f.nowFunc = func() time.Time { return now }

	drifts := makeFilterDrifts([]string{"KEY_X"}, "warning")
	_ = f.Send("prod", drifts)
	_ = f.Send("prod", drifts) // should be suppressed

	if len(cap.calls) != 1 {
		t.Errorf("expected 1 call due to cooldown, got %d", len(cap.calls))
	}
}

func TestFilter_Cooldown_AllowsAfterExpiry(t *testing.T) {
	cap := &captureSender{}
	now := time.Now()
	f := NewFilter(cap, FilterConfig{Cooldown: 5 * time.Minute})
	f.nowFunc = func() time.Time { return now }

	drifts := makeFilterDrifts([]string{"KEY_Y"}, "warning")
	_ = f.Send("prod", drifts)

	f.nowFunc = func() time.Time { return now.Add(10 * time.Minute) }
	_ = f.Send("prod", drifts)

	if len(cap.calls) != 2 {
		t.Errorf("expected 2 calls after cooldown expiry, got %d", len(cap.calls))
	}
}

func TestFilter_PropagatesSenderError(t *testing.T) {
	sentinel := errors.New("backend down")
	cap := &captureSender{err: sentinel}
	f := NewFilter(cap, FilterConfig{})
	drifts := makeFilterDrifts([]string{"KEY_Z"}, "warning")
	err := f.Send("prod", drifts)
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
