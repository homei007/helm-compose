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
	"reflect"
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	cfg "github.com/seacrew/helm-compose/internal/config"
)

func TestHelmShortVersionParsesHelm3AndHelm4(t *testing.T) {
	for _, versionOutput := range []string{"v3.11.0", "v4.2.3"} {
		version, err := semver.NewVersion(strings.TrimSpace(versionOutput))
		if err != nil {
			t.Fatalf("failed to parse Helm version %q: %v", versionOutput, err)
		}
		if version.LessThan(minVersion) {
			t.Fatalf("Helm version %q was incorrectly rejected", versionOutput)
		}
	}
}

func TestCreateHelmArgumentsSupportsHelm3PostRendererPath(t *testing.T) {
	release := &cfg.Release{
		Chart:            "example/chart",
		PostRenderer:     "./renderer",
		PostRendererArgs: []string{"--label", "value"},
	}

	args, err := createHelmArguments(HELM_UPGRADE, "example", release)
	if err != nil {
		t.Fatalf("createHelmArguments returned an error: %v", err)
	}

	want := []string{
		"upgrade",
		"--install",
		"--post-renderer=./renderer",
		"--post-renderer-args=--label",
		"--post-renderer-args=value",
		"example",
		"example/chart",
	}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("unexpected Helm arguments: got %#v, want %#v", args, want)
	}
}

func TestCreateHelmArgumentsSupportsHelm4PostRendererPlugin(t *testing.T) {
	release := &cfg.Release{
		Chart:              "example/chart",
		PostRendererPlugin: "inject-labels",
		PostRendererArgs:   []string{"--label", "value"},
	}

	args, err := createHelmArguments(HELM_UPGRADE, "example", release)
	if err != nil {
		t.Fatalf("createHelmArguments returned an error: %v", err)
	}

	if got, want := args[2], "--post-renderer=inject-labels"; got != want {
		t.Fatalf("unexpected post-renderer argument: got %q, want %q", got, want)
	}
	wantArgs := []string{"--post-renderer-args=--label", "--post-renderer-args=value"}
	if got := args[3:5]; !reflect.DeepEqual(got, wantArgs) {
		t.Fatalf("unexpected post-renderer arguments: got %#v, want %#v", got, wantArgs)
	}
}

func TestCreateHelmArgumentsRejectsAmbiguousPostRendererConfiguration(t *testing.T) {
	release := &cfg.Release{
		Chart:              "example/chart",
		PostRenderer:       "./renderer",
		PostRendererPlugin: "inject-labels",
	}

	_, err := createHelmArguments(HELM_UPGRADE, "example", release)
	if err == nil || !strings.Contains(err.Error(), "both postRenderer and postRendererPlugin") {
		t.Fatalf("expected an ambiguous post-renderer error, got %v", err)
	}
}
