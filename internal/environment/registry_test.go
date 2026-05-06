package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func makeFileCollector(t *testing.T, name, content string) *Collector {
	t.Helper()
	tmp := filepath.Join(t.TempDir(), name+".conf")
	if err := os.WriteFile(tmp, []byte(content), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	c, err := NewCollector(CollectorConfig{Name: name, Source: SourceFile, FilePath: tmp})
	if err != nil {
		t.Fatalf("NewCollector: %v", err)
	}
	return c
}

func TestRegistry_RegisterAndCollect(t *testing.T) {
	r := NewRegistry()
	c := makeFileCollector(t, "staging", "KEY=value\n")
	if err := r.Register(c); err != nil {
		t.Fatalf("Register: %v", err)
	}
	errs := r.CollectAll()
	if errs != nil {
		t.Fatalf("CollectAll errors: %v", errs)
	}
	snap, ok := r.Snapshot("staging")
	if !ok {
		t.Fatal("expected snapshot for staging")
	}
	if snap.Data["KEY"] != "value" {
		t.Errorf("expected KEY=value, got %q", snap.Data["KEY"])
	}
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	r := NewRegistry()
	c1 := makeFileCollector(t, "prod", "A=1\n")
	c2 := makeFileCollector(t, "prod", "A=2\n")
	if err := r.Register(c1); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	if err := r.Register(c2); err == nil {
		t.Fatal("expected error on duplicate register")
	}
}

func TestRegistry_SnapshotMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Snapshot("nonexistent")
	if ok {
		t.Fatal("expected snapshot to be missing")
	}
}

func TestRegistry_CollectAll_PartialFailure(t *testing.T) {
	r := NewRegistry()
	good := makeFileCollector(t, "good", "X=1\n")
	bad, _ := NewCollector(CollectorConfig{Name: "bad", Source: SourceFile, FilePath: "/no/such/file"})
	_ = r.Register(good)
	_ = r.Register(bad)
	errs := r.CollectAll()
	if errs == nil {
		t.Fatal("expected errors map for bad collector")
	}
	if _, hasErr := errs["bad"]; !hasErr {
		t.Error("expected error entry for 'bad' collector")
	}
	if _, ok := r.Snapshot("good"); !ok {
		t.Error("expected snapshot for 'good' collector despite other failure")
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	for _, name := range []string{"alpha", "beta", "gamma"} {
		c := makeFileCollector(t, name, "")
		if err := r.Register(c); err != nil {
			t.Fatalf("Register %s: %v", name, err)
		}
	}
	names := r.Names()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}
}
