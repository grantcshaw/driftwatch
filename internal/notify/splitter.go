package notify

import (
	"fmt"

	"github.com/example/driftwatch/internal/drift"
)

// SplitFunc decides which bucket name a drift belongs to.
type SplitFunc func(d drift.Drift) string

// Splitter partitions incoming drifts by a key function and routes each
// partition to the matching registered sender. Unmatched partitions are
// sent to a default sender when configured.
type Splitter struct {
	splitFn   SplitFunc
	routes    map[string]Sender
	defaultSn Sender
}

// NewSplitter creates a Splitter with the given split function.
func NewSplitter(fn SplitFunc) (*Splitter, error) {
	if fn == nil {
		return nil, fmt.Errorf("splitter: split function must not be nil")
	}
	return &Splitter{
		splitFn: fn,
		routes:  make(map[string]Sender),
	}, nil
}

// Register associates a bucket name with a sender.
func (s *Splitter) Register(bucket string, sender Sender) error {
	if bucket == "" {
		return fmt.Errorf("splitter: bucket name must not be empty")
	}
	if sender == nil {
		return fmt.Errorf("splitter: sender must not be nil")
	}
	s.routes[bucket] = sender
	return nil
}

// SetDefault registers a fallback sender for unmatched buckets.
func (s *Splitter) SetDefault(sender Sender) error {
	if sender == nil {
		return fmt.Errorf("splitter: default sender must not be nil")
	}
	s.defaultSn = sender
	return nil
}

// Send partitions drifts by bucket and dispatches each group to its sender.
func (s *Splitter) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	partitions := make(map[string][]drift.Drift)
	for _, d := range drifts {
		bucket := s.splitFn(d)
		partitions[bucket] = append(partitions[bucket], d)
	}

	var errs []error
	for bucket, group := range partitions {
		sender, ok := s.routes[bucket]
		if !ok {
			sender = s.defaultSn
		}
		if sender == nil {
			continue
		}
		if err := sender.Send(env, group); err != nil {
			errs = append(errs, fmt.Errorf("splitter bucket %q: %w", bucket, err))
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
