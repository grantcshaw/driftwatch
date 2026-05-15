package drift

import (
	"fmt"

	"github.com/yourorg/driftwatch/internal/environment"
)

// Drift describes a single configuration key that differs between
// a baseline snapshot and the current environment snapshot.
type Drift struct {
	Key           string
	BaselineValue string
	CurrentValue  string
	Severity      string            // "warning" or "critical"
	Metadata      map[string]string // optional annotations
}

// Detector compares a baseline snapshot against a current snapshot
// and produces a list of Drift items.
type Detector struct {
	criticalKeys map[string]struct{}
}

// NewDetector constructs a Detector. criticalKeys is an optional set of
// key names that should be reported with severity "critical" instead of
// the default "warning".
func NewDetector(criticalKeys []string) *Detector {
	cm := make(map[string]struct{}, len(criticalKeys))
	for _, k := range criticalKeys {
		cm[k] = struct{}{}
	}
	return &Detector{criticalKeys: cm}
}

// Detect returns all keys that differ between baseline and current.
// Keys present only in baseline are reported as missing (CurrentValue = "").
// Keys present only in current are reported as extra (BaselineValue = "").
func (d *Detector) Detect(baseline, current *environment.Snapshot) ([]Drift, error) {
	if baseline == nil {
		return nil, fmt.Errorf("detector: baseline snapshot must not be nil")
	}
	if current == nil {
		return nil, fmt.Errorf("detector: current snapshot must not be nil")
	}

	var drifts []Drift

	for k, bv := range baseline.Values {
		cv, ok := current.Values[k]
		if !ok {
			drifts = append(drifts, d.newDrift(k, bv, ""))
			continue
		}
		if bv != cv {
			drifts = append(drifts, d.newDrift(k, bv, cv))
		}
	}

	for k, cv := range current.Values {
		if _, exists := baseline.Values[k]; !exists {
			drifts = append(drifts, d.newDrift(k, "", cv))
		}
	}

	return drifts, nil
}

func (d *Detector) newDrift(key, baseline, current string) Drift {
	sev := "warning"
	if _, ok := d.criticalKeys[key]; ok {
		sev = "critical"
	}
	return Drift{
		Key:           key,
		BaselineValue: baseline,
		CurrentValue:  current,
		Severity:      sev,
	}
}
