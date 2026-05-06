package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

// Record represents a single drift detection event persisted to disk.
type Record struct {
	Timestamp   time.Time    `json:"timestamp"`
	Environment string       `json:"environment"`
	Drifts      []drift.Drift `json:"drifts"`
	DriftCount  int          `json:"drift_count"`
}

// Store persists drift records to a directory on disk.
type Store struct {
	dir string
}

// NewStore creates a new Store backed by the given directory.
// The directory is created if it does not exist.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("history: create store dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a drift record to disk using a timestamped filename.
func (s *Store) Save(env string, drifts []drift.Drift) error {
	rec := Record{
		Timestamp:   time.Now().UTC(),
		Environment: env,
		Drifts:      drifts,
		DriftCount:  len(drifts),
	}

	filename := fmt.Sprintf("%s_%s.json", env, rec.Timestamp.Format("20060102T150405Z"))
	path := filepath.Join(s.dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("history: create record file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rec); err != nil {
		return fmt.Errorf("history: encode record: %w", err)
	}
	return nil
}

// LoadAll reads all records for a given environment from the store directory.
func (s *Store) LoadAll(env string) ([]Record, error) {
	pattern := filepath.Join(s.dir, env+"_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("history: glob records: %w", err)
	}

	records := make([]Record, 0, len(matches))
	for _, path := range matches {
		rec, err := loadRecord(path)
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

func loadRecord(path string) (Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return Record{}, fmt.Errorf("history: open record %s: %w", path, err)
	}
	defer f.Close()

	var rec Record
	if err := json.NewDecoder(f).Decode(&rec); err != nil {
		return Record{}, fmt.Errorf("history: decode record %s: %w", path, err)
	}
	return rec, nil
}
