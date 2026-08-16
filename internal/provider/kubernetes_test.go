package provider

import (
	"os"
	"path/filepath"
	"testing"

	cfg "github.com/seacrew/helm-compose/internal/config"
)

func TestKubernetesRESTConfigUsesConfiguredContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	data := []byte(`apiVersion: v1
kind: Config
clusters:
- name: one
  cluster:
    server: https://one.example
- name: two
  cluster:
    server: https://two.example
contexts:
- name: one
  context:
    cluster: one
    user: user
- name: two
  context:
    cluster: two
    user: user
current-context: one
users:
- name: user
  user:
    token: token
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	config, err := kubernetesRESTConfig(&cfg.Storage{KubeConfig: path, KubeContext: "two"})
	if err != nil {
		t.Fatalf("kubernetesRESTConfig returned an error: %v", err)
	}
	if config.Host != "https://two.example" {
		t.Fatalf("expected configured context host, got %q", config.Host)
	}
}
