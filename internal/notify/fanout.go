package notify

import (
	"errors"
	"fmt"

	"github.com/your-org/driftwatch/internal/drift"
)

// Fanout sends drift results to multiple senders concurrently and collects
// all errors. Unlike MultiSender (which is sequential), Fanout dispatches
// all sends in parallel and waits for all to finish.
type Fanout struct {
	senders []Sender
}

// NewFanout creates a Fanout that dispatches to the given senders concurrently.
// At least one sender must be provided.
func NewFanout(senders ...Sender) (*Fanout, error) {
	if len(senders) == 0 {
		return nil, errors.New("fanout: at least one sender required")
	}
	for i, s := range senders {
		if s == nil {
			return nil, fmt.Errorf("fanout: sender at index %d is nil", i)
		}
	}
	return &Fanout{senders: senders}, nil
}

// Add appends a sender to the fanout. Returns an error if s is nil.
func (f *Fanout) Add(s Sender) error {
	if s == nil {
		return errors.New("fanout: cannot add nil sender")
	}
	f.senders = append(f.senders, s)
	return nil
}

// Len returns the number of registered senders.
func (f *Fanout) Len() int { return len(f.senders) }

// Send dispatches drifts to all senders concurrently. All errors are joined
// and returned as a single combined error; a partial failure does not prevent
// other senders from being called.
func (f *Fanout) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	type result struct {
		idx int
		err error
	}

	ch := make(chan result, len(f.senders))
	for i, s := range f.senders {
		go func(idx int, s Sender) {
			ch <- result{idx: idx, err: s.Send(env, drifts)}
		}(i, s)
	}

	var errs []error
	for range f.senders {
		r := <-ch
		if r.err != nil {
			errs = append(errs, fmt.Errorf("fanout[%d]: %w", r.idx, r.err))
		}
	}

	return errors.Join(errs...)
}
