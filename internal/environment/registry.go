package environment

import (
	"fmt"
	"sync"
)

// Registry manages a set of named Collectors and caches their latest Snapshots.
type Registry struct {
	mu         sync.RWMutex
	collectors map[string]*Collector
	snapshots  map[string]*Snapshot
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		collectors: make(map[string]*Collector),
		snapshots:  make(map[string]*Snapshot),
	}
}

// Register adds a Collector to the registry.
// Returns an error if a collector with the same name already exists.
func (r *Registry) Register(c *Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := c.cfg.Name
	if _, exists := r.collectors[name]; exists {
		return fmt.Errorf("collector %q already registered", name)
	}
	r.collectors[name] = c
	return nil
}

// CollectAll runs all registered collectors and stores the resulting snapshots.
// It returns a map of environment name to error for any collectors that failed.
func (r *Registry) CollectAll() map[string]error {
	r.mu.Lock()
	defer r.mu.Unlock()
	errs := make(map[string]error)
	for name, c := range r.collectors {
		snap, err := c.Collect()
		if err != nil {
			errs[name] = err
			continue
		}
		r.snapshots[name] = snap
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// Snapshot returns the latest cached snapshot for the given environment name.
func (r *Registry) Snapshot(name string) (*Snapshot, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.snapshots[name]
	return s, ok
}

// Names returns the registered environment names in insertion-stable order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.collectors))
	for n := range r.collectors {
		names = append(names, n)
	}
	return names
}
