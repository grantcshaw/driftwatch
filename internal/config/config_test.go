package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	content := `
environments:
  - name: staging
    provider: aws
    region: us-east-1
  - name: production
    provider: aws
    region: us-west-2
check_interval: 10m
alerts:
  log_only: true
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(cfg.Environments))
	}
	if cfg.CheckInterval != 10*time.Minute {
		t.Errorf("expected 10m interval, got %v", cfg.CheckInterval)
	}
	if !cfg.Alerts.LogOnly {
		t.Error("expected log_only to be true")
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	content := `
environments:
  - name: dev
    provider: gcp
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 5*time.Minute {
		t.Errorf("expected default 5m interval, got %v", cfg.CheckInterval)
	}
}

func TestLoad_NoEnvironments(t *testing.T) {
	path := writeTempConfig(t, "check_interval: 1m\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing environments")
	}
}

func TestLoad_DuplicateEnvName(t *testing.T) {
	content := `
environments:
  - name: prod
    provider: aws
  - name: prod
    provider: gcp
`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate environment name")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
