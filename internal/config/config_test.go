/*
Copyright © 2023 The Helm Compose Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseSimpleConfig(t *testing.T) {
	config, err := ParseComposeFile("../../examples/simple-compose.yaml")

	if err != nil {
		log.Fatal(err)
	}

	if config.Storage.Name != "simple" {
		log.Fatalf("Was expecting revision name 'simple' but got '%s'", config.Storage.Name)
	}

	if config.Storage.Type != Local {
		log.Fatalf("Was expecting revision provider type '%s' but got '%s'", Local, config.Storage.Type)
	}

	if len(config.Releases) != 2 {
		log.Fatalf("Was expecting 2 release but got %d", len(config.Releases))
	}
}

func TestFindComposeConfigOnlyChecksRequestedDirectory(t *testing.T) {
	directory := t.TempDir()
	want := filepath.Join(directory, "helm-compose.yaml")
	if err := os.WriteFile(want, []byte("apiVersion: 1.1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(directory, "nested")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "compose.yaml"), []byte("apiVersion: 1.1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if got := findComposeConfigIn(directory); !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("unexpected discovered files: got %#v, want %#v", got, []string{want})
	}
}

func TestValidateReleaseDependencies(t *testing.T) {
	tests := []struct {
		name     string
		releases map[string]Release
		wantErr  string
	}{
		{
			name: "valid",
			releases: map[string]Release{
				"app":      {Needs: []string{"database"}},
				"database": {},
			},
		},
		{
			name:     "unknown",
			releases: map[string]Release{"app": {Needs: []string{"missing"}}},
			wantErr:  "unknown release",
		},
		{
			name:     "self",
			releases: map[string]Release{"app": {Needs: []string{"app"}}},
			wantErr:  "cannot depend on itself",
		},
		{
			name: "duplicate",
			releases: map[string]Release{
				"app":      {Needs: []string{"database", "database"}},
				"database": {},
			},
			wantErr: "more than once",
		},
		{
			name: "cycle",
			releases: map[string]Release{
				"app":      {Needs: []string{"database"}},
				"database": {Needs: []string{"app"}},
			},
			wantErr: "app -> database -> app",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateReleaseDependencies(test.releases)
			if test.wantErr == "" && err != nil {
				t.Fatalf("ValidateReleaseDependencies returned an error: %v", err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("expected error containing %q, got %v", test.wantErr, err)
			}
		})
	}
}

func TestNeedsRequiresAPIVersion11(t *testing.T) {
	_, err := parseComposeData([]byte("apiVersion: 1.0\nreleases:\n  app:\n    chart: example/app\n    needs: [database]\n  database:\n    chart: example/database\n"))
	if err == nil || !strings.Contains(err.Error(), "apiVersion 1.1+") {
		t.Fatalf("expected an apiVersion error, got %v", err)
	}
}
