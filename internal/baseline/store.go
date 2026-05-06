// Package baseline manages the storage and retrieval of environment snapshots
// used as reference points for drift detection.
package baseline

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/driftwatch/internal/environment"
)

// ErrNotFound is returned when no baseline exists for a given environment.
var ErrNotFound = errors.New("baseline not found")

// Store persists and retrieves baseline snapshots on disk.
type Store struct {
	dir string
}

// NewStore creates a Store rooted at dir, creating the directory if needed.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("baseline: create dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save writes snapshot as the current baseline for its environment.
func (s *Store) Save(snap *environment.Snapshot) error {
	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	path := s.filePath(snap.EnvName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("baseline: write %s: %w", path, err)
	}
	return nil
}

// Load retrieves the stored baseline for envName.
// Returns ErrNotFound if no baseline has been saved yet.
func (s *Store) Load(envName string) (*environment.Snapshot, error) {
	path := s.filePath(envName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, envName)
	}
	if err != nil {
		return nil, fmt.Errorf("baseline: read %s: %w", path, err)
	}
	var snap environment.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	return &snap, nil
}

// Delete removes the stored baseline for envName.
// Returns nil if no baseline exists.
func (s *Store) Delete(envName string) error {
	err := os.Remove(s.filePath(envName))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) filePath(envName string) string {
	return filepath.Join(s.dir, envName+".baseline.json")
}
