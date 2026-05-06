package environment

import (
	"fmt"
	"os"
	"strings"
)

// Source defines where environment data is collected from.
type Source string

const (
	SourceEnv  Source = "env"
	SourceFile Source = "file"
)

// CollectorConfig holds settings for a single environment collector.
type CollectorConfig struct {
	Name   string
	Source Source
	// Prefix filters env vars by prefix when Source is SourceEnv.
	Prefix string
	// FilePath is used when Source is SourceFile.
	FilePath string
}

// Collector gathers key/value pairs from a configured source.
type Collector struct {
	cfg CollectorConfig
}

// NewCollector creates a Collector from the given config.
func NewCollector(cfg CollectorConfig) (*Collector, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("collector name must not be empty")
	}
	if cfg.Source != SourceEnv && cfg.Source != SourceFile {
		return nil, fmt.Errorf("unknown source %q for collector %q", cfg.Source, cfg.Name)
	}
	return &Collector{cfg: cfg}, nil
}

// Collect gathers key/value pairs and returns a Snapshot.
func (c *Collector) Collect() (*Snapshot, error) {
	switch c.cfg.Source {
	case SourceEnv:
		return c.collectFromEnv()
	case SourceFile:
		return c.collectFromFile()
	}
	return nil, fmt.Errorf("unsupported source: %s", c.cfg.Source)
}

func (c *Collector) collectFromEnv() (*Snapshot, error) {
	data := make(map[string]string)
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if c.cfg.Prefix == "" || strings.HasPrefix(key, c.cfg.Prefix) {
			data[key] = val
		}
	}
	return NewSnapshot(c.cfg.Name, data)
}

func (c *Collector) collectFromFile() (*Snapshot, error) {
	raw, err := os.ReadFile(c.cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", c.cfg.FilePath, err)
	}
	data := make(map[string]string)
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return NewSnapshot(c.cfg.Name, data)
}
