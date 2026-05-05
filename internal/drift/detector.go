package drift

import (
	"fmt"

	"github.com/yourusername/driftwatch/internal/environment"
)

// Result holds the outcome of a drift comparison between two snapshots.
type Result struct {
	BaseEnv    string
	TargetEnv  string
	Drifted    bool
	DiffKeys   []string
	MissingKeys []string
	ExtraKeys  []string
}

// String returns a human-readable summary of the drift result.
func (r Result) String() string {
	if !r.Drifted {
		return fmt.Sprintf("[OK] %s and %s are in sync", r.BaseEnv, r.TargetEnv)
	}
	return fmt.Sprintf("[DRIFT] %s vs %s — changed: %d, missing: %d, extra: %d",
		r.BaseEnv, r.TargetEnv, len(r.DiffKeys), len(r.MissingKeys), len(r.ExtraKeys))
}

// Detector compares environment snapshots to identify configuration drift.
type Detector struct{}

// NewDetector creates a new Detector instance.
func NewDetector() *Detector {
	return &Detector{}
}

// Compare checks two snapshots for drift and returns a detailed Result.
func (d *Detector) Compare(base, target *environment.Snapshot) Result {
	result := Result{
		BaseEnv:   base.EnvName,
		TargetEnv: target.EnvName,
	}

	baseKeys := base.Keys()
	targetMap := target.ToMap()
	baseMap := base.ToMap()

	for _, key := range baseKeys {
		targetVal, exists := targetMap[key]
		if !exists {
			result.MissingKeys = append(result.MissingKeys, key)
		} else if baseMap[key] != targetVal {
			result.DiffKeys = append(result.DiffKeys, key)
		}
	}

	for key := range targetMap {
		if _, exists := baseMap[key]; !exists {
			result.ExtraKeys = append(result.ExtraKeys, key)
		}
	}

	result.Drifted = len(result.DiffKeys) > 0 ||
		len(result.MissingKeys) > 0 ||
		len(result.ExtraKeys) > 0

	return result
}
