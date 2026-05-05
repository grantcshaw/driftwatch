package environment

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Snapshot represents the captured state of an environment at a point in time.
type Snapshot struct {
	Environment string            `json:"environment"`
	CapturedAt  time.Time         `json:"captured_at"`
	Values      map[string]string `json:"values"`
	Checksum    string            `json:"checksum"`
}

// NewSnapshot creates a new Snapshot for the given environment and values,
// computing a deterministic checksum over the key-value pairs.
func NewSnapshot(env string, values map[string]string) (*Snapshot, error) {
	if env == "" {
		return nil, fmt.Errorf("environment name must not be empty")
	}

	checksum, err := computeChecksum(values)
	if err != nil {
		return nil, fmt.Errorf("computing checksum: %w", err)
	}

	return &Snapshot{
		Environment: env,
		CapturedAt:  time.Now().UTC(),
		Values:      values,
		Checksum:    checksum,
	}, nil
}

// Equal reports whether two snapshots have identical state (same checksum).
func (s *Snapshot) Equal(other *Snapshot) bool {
	if s == nil || other == nil {
		return s == other
	}
	return s.Checksum == other.Checksum
}

// DiffKeys returns the set of keys that differ between this snapshot and another.
// Keys present in one but not the other are included.
func (s *Snapshot) DiffKeys(other *Snapshot) []string {
	seen := make(map[string]struct{})
	var diffs []string

	for k, v := range s.Values {
		seen[k] = struct{}{}
		if ov, ok := other.Values[k]; !ok || ov != v {
			diffs = append(diffs, k)
		}
	}
	for k := range other.Values {
		if _, ok := seen[k]; !ok {
			diffs = append(diffs, k)
		}
	}
	return diffs
}

// computeChecksum produces a stable SHA-256 hash over the values map.
func computeChecksum(values map[string]string) (string, error) {
	b, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h), nil
}
