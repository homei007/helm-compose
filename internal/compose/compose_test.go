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
package compose

import (
	"errors"
	"reflect"
	"testing"

	cfg "github.com/seacrew/helm-compose/internal/config"
)

func TestRunUpDoesNotStoreFailedDeployment(t *testing.T) {
	stored := false
	stubComposeDependencies(t,
		func(string, ...string) (string, error) {
			return "upgrade failed", errors.New("exit status 1")
		},
		func(*cfg.Config) (*cfg.Config, error) { return nil, nil },
		func(*cfg.Config) error {
			stored = true
			return nil
		},
	)

	config := testConfig(map[string]cfg.Release{
		"broken": {Chart: "example/broken"},
	})

	if err := RunUp(config, nil); err == nil {
		t.Fatal("expected RunUp to return the Helm error")
	}
	if stored {
		t.Fatal("failed deployment must not be stored as an applied revision")
	}
}

func TestRunUpDoesNotForgetFailedUninstall(t *testing.T) {
	stored := false
	previous := testConfig(map[string]cfg.Release{
		"removed": {Chart: "example/removed"},
	})
	stubComposeDependencies(t,
		func(string, ...string) (string, error) {
			return "uninstall failed", errors.New("exit status 1")
		},
		func(*cfg.Config) (*cfg.Config, error) { return previous, nil },
		func(*cfg.Config) error {
			stored = true
			return nil
		},
	)

	if err := RunUp(testConfig(nil), nil); err == nil {
		t.Fatal("expected RunUp to return the uninstall error")
	}
	if stored {
		t.Fatal("failed uninstall must not remove the release from applied state")
	}
}

func TestRunUpSelectedReleaseMergesAppliedState(t *testing.T) {
	previous := testConfig(map[string]cfg.Release{
		"wordpress":  {Chart: "example/wordpress", ChartVersion: "1.0.0"},
		"wordpress2": {Chart: "example/wordpress", ChartVersion: "1.0.0"},
	})
	current := testConfig(map[string]cfg.Release{
		"wordpress":  {Chart: "example/wordpress", ChartVersion: "2.0.0"},
		"wordpress2": {Chart: "example/wordpress", ChartVersion: "2.0.0"},
	})

	var commands [][]string
	var stored *cfg.Config
	stubComposeDependencies(t,
		func(_ string, args ...string) (string, error) {
			commands = append(commands, append([]string{}, args...))
			return "", nil
		},
		func(*cfg.Config) (*cfg.Config, error) { return previous, nil },
		func(config *cfg.Config) error {
			stored = cloneConfig(config)
			return nil
		},
	)

	if err := RunUp(current, []string{"wordpress2", "wordpress2"}); err != nil {
		t.Fatalf("RunUp returned an error: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("expected one Helm command, got %d", len(commands))
	}
	if got := commands[0][len(commands[0])-2]; got != "wordpress2" {
		t.Fatalf("expected wordpress2 to be upgraded, got %q", got)
	}
	if stored == nil {
		t.Fatal("expected merged applied state to be stored")
	}
	if got := stored.Releases["wordpress"].ChartVersion; got != "1.0.0" {
		t.Fatalf("unselected release was changed in applied state: %q", got)
	}
	if got := stored.Releases["wordpress2"].ChartVersion; got != "2.0.0" {
		t.Fatalf("selected release was not updated in applied state: %q", got)
	}
}

func TestRunDownSelectedReleaseUpdatesAppliedState(t *testing.T) {
	previous := testConfig(map[string]cfg.Release{
		"wordpress":  {Chart: "example/wordpress"},
		"wordpress2": {Chart: "example/wordpress"},
	})

	var commands [][]string
	var stored *cfg.Config
	stubComposeDependencies(t,
		func(_ string, args ...string) (string, error) {
			commands = append(commands, append([]string{}, args...))
			return "", nil
		},
		func(*cfg.Config) (*cfg.Config, error) { return previous, nil },
		func(config *cfg.Config) error {
			stored = cloneConfig(config)
			return nil
		},
	)

	if err := RunDown(testConfig(previous.Releases), []string{"wordpress2"}); err != nil {
		t.Fatalf("RunDown returned an error: %v", err)
	}

	if len(commands) != 1 {
		t.Fatalf("expected one Helm command, got %d", len(commands))
	}
	if want := []string{"uninstall", "wordpress2"}; !reflect.DeepEqual(commands[0], want) {
		t.Fatalf("unexpected Helm command: %#v", commands[0])
	}
	if stored == nil {
		t.Fatal("expected updated applied state to be stored")
	}
	if _, ok := stored.Releases["wordpress2"]; ok {
		t.Fatal("uninstalled release remains in applied state")
	}
	if _, ok := stored.Releases["wordpress"]; !ok {
		t.Fatal("unselected release was removed from applied state")
	}
}

func TestRunUpRejectsUnknownSelectedRelease(t *testing.T) {
	executed := false
	stubComposeDependencies(t,
		func(string, ...string) (string, error) {
			executed = true
			return "", nil
		},
		func(*cfg.Config) (*cfg.Config, error) { return nil, nil },
		func(*cfg.Config) error { return nil },
	)

	err := RunUp(testConfig(map[string]cfg.Release{"wordpress": {Chart: "example/wordpress"}}), []string{"missing"})
	if err == nil {
		t.Fatal("expected an unknown release error")
	}
	if executed {
		t.Fatal("Helm must not run when release validation fails")
	}
}

func stubComposeDependencies(
	t *testing.T,
	execute func(string, ...string) (string, error),
	load func(*cfg.Config) (*cfg.Config, error),
	store func(*cfg.Config) error,
) {
	t.Helper()
	oldExecute := executeHelm
	oldLoad := loadConfig
	oldStore := storeConfig
	executeHelm = execute
	loadConfig = load
	storeConfig = store
	t.Cleanup(func() {
		executeHelm = oldExecute
		loadConfig = oldLoad
		storeConfig = oldStore
	})
}

func testConfig(releases map[string]cfg.Release) *cfg.Config {
	return &cfg.Config{
		Version:  "1.1",
		Storage:  cfg.Storage{Type: cfg.Local, Name: "test"},
		Releases: releases,
	}
}
