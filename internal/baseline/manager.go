package baseline

import (
	"fmt"
	"time"

	"github.com/example/driftwatch/internal/environment"
)

// Manager coordinates capturing and promoting baselines for environments.
type Manager struct {
	store    *Store
	registry *environment.Registry
}

// NewManager returns a Manager backed by store and registry.
func NewManager(store *Store, registry *environment.Registry) *Manager {
	return &Manager{store: store, registry: registry}
}

// Capture collects a fresh snapshot for envName and saves it as the baseline.
func (m *Manager) Capture(envName string) (*environment.Snapshot, error) {
	snap, err := m.registry.Snapshot(envName)
	if err != nil {
		return nil, fmt.Errorf("baseline manager: collect %s: %w", envName, err)
	}
	if err := m.store.Save(snap); err != nil {
		return nil, fmt.Errorf("baseline manager: save %s: %w", envName, err)
	}
	return snap, nil
}

// Promote saves an already-collected snapshot as the new baseline.
func (m *Manager) Promote(snap *environment.Snapshot) error {
	if snap == nil {
		return fmt.Errorf("baseline manager: cannot promote nil snapshot")
	}
	return m.store.Save(snap)
}

// Current returns the stored baseline for envName.
func (m *Manager) Current(envName string) (*environment.Snapshot, error) {
	return m.store.Load(envName)
}

// Age returns how long ago the baseline for envName was collected.
// Returns ErrNotFound if no baseline exists.
func (m *Manager) Age(envName string) (time.Duration, error) {
	snap, err := m.store.Load(envName)
	if err != nil {
		return 0, err
	}
	return time.Since(snap.Collected), nil
}
