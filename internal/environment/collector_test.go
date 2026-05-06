package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCollector_EmptyName(t *testing.T) {
	_, err := NewCollector(CollectorConfig{Source: SourceEnv})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNewCollector_UnknownSource(t *testing.T) {
	_, err := NewCollector(CollectorConfig{Name: "x", Source: "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestCollector_CollectFromEnv_NoPrefix(t *testing.T) {
	t.Setenv("DW_TEST_KEY", "hello")
	c, err := NewCollector(CollectorConfig{Name: "test", Source: SourceEnv})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap, err := c.Collect()
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if snap.Data["DW_TEST_KEY"] != "hello" {
		t.Errorf("expected DW_TEST_KEY=hello, got %q", snap.Data["DW_TEST_KEY"])
	}
}

func TestCollector_CollectFromEnv_WithPrefix(t *testing.T) {
	t.Setenv("DW_APP_PORT", "8080")
	t.Setenv("OTHER_VAR", "ignored")
	c, err := NewCollector(CollectorConfig{Name: "app", Source: SourceEnv, Prefix: "DW_APP_"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap, err := c.Collect()
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if _, ok := snap.Data["OTHER_VAR"]; ok {
		t.Error("OTHER_VAR should have been filtered out")
	}
	if snap.Data["DW_APP_PORT"] != "8080" {
		t.Errorf("expected DW_APP_PORT=8080, got %q", snap.Data["DW_APP_PORT"])
	}
}

func TestCollector_CollectFromFile(t *testing.T) {
	content := "# comment\nHOST=localhost\nPORT=5432\n"
	tmp := filepath.Join(t.TempDir(), "env.conf")
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	c, err := NewCollector(CollectorConfig{Name: "db", Source: SourceFile, FilePath: tmp})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap, err := c.Collect()
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if snap.Data["HOST"] != "localhost" {
		t.Errorf("expected HOST=localhost, got %q", snap.Data["HOST"])
	}
	if snap.Data["PORT"] != "5432" {
		t.Errorf("expected PORT=5432, got %q", snap.Data["PORT"])
	}
}

func TestCollector_CollectFromFile_Missing(t *testing.T) {
	c, err := NewCollector(CollectorConfig{Name: "x", Source: SourceFile, FilePath: "/nonexistent/path.conf"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.Collect()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
