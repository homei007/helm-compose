package provider

import (
	"os"
	"path/filepath"
	"testing"

	cfg "github.com/seacrew/helm-compose/internal/config"
)

func TestLocalProviderStoreCreatesNestedDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state")
	provider := newLocalProvider(&cfg.Storage{
		Name:              "test",
		Path:              path,
		NumberOfRevisions: 3,
	})
	encoded := "encoded"

	if err := provider.store(&encoded); err != nil {
		t.Fatalf("store returned an error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(path, "test-1")); err != nil {
		t.Fatalf("stored revision was not created: %v", err)
	}
}

func TestLocalProviderLoadDoesNotCreateMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	provider := newLocalProvider(&cfg.Storage{Name: "test", Path: path})

	data, err := provider.load()
	if err != nil {
		t.Fatalf("load returned an error: %v", err)
	}
	if data != nil {
		t.Fatalf("expected no stored config, got %q", string(*data))
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("load created or unexpectedly found the state directory: %v", err)
	}
}
