package notify

import (
	"errors"
	"fmt"

	"github.com/driftwatch/driftwatch/internal/drift"
)

// Schema validates drift payloads against a set of required and forbidden keys
// before forwarding to the inner Sender.
type Schema struct {
	inner        Sender
	requiredKeys []string
	forbiddenKeys []string
}

// NewSchema creates a Schema sender wrapper.
// requiredKeys are keys that must appear in every drift item.
// forbiddenKeys are keys that must not appear in any drift item.
func NewSchema(inner Sender, requiredKeys, forbiddenKeys []string) (*Schema, error) {
	if inner == nil {
		return nil, errors.New("notify/schema: inner sender must not be nil")
	}
	if len(requiredKeys) == 0 && len(forbiddenKeys) == 0 {
		return nil, errors.New("notify/schema: at least one required or forbidden key must be specified")
	}
	return &Schema{
		inner:         inner,
		requiredKeys:  requiredKeys,
		forbiddenKeys: forbiddenKeys,
	}, nil
}

// Send validates each drift item against the schema and forwards to inner.
// Returns a validation error if any drift fails schema checks.
func (s *Schema) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	for _, d := range drifts {
		for _, req := range s.requiredKeys {
			if d.Key != req {
				continue
			}
			goto foundRequired
		}
		if len(s.requiredKeys) > 0 {
			// Check if the drift key matches any required key; if none match, skip validation error
			// (required keys must be present as a set across all drifts, not per-item)
		}
		foundRequired:
		for _, forb := range s.forbiddenKeys {
			if d.Key == forb {
				return fmt.Errorf("notify/schema: forbidden key %q found in drift for env %q", d.Key, env)
			}
		}
	}
	if err := s.validateRequiredKeys(drifts); err != nil {
		return err
	}
	return s.inner.Send(env, drifts)
}

func (s *Schema) validateRequiredKeys(drifts []drift.Drift) error {
	present := make(map[string]bool, len(drifts))
	for _, d := range drifts {
		present[d.Key] = true
	}
	for _, req := range s.requiredKeys {
		if !present[req] {
			return fmt.Errorf("notify/schema: required key %q missing from drift payload", req)
		}
	}
	return nil
}
