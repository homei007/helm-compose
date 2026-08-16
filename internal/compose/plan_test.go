package compose

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	cfg "github.com/seacrew/helm-compose/internal/config"
)

func TestCreatePlanShowsDeterministicDependencyGroups(t *testing.T) {
	previous := testConfig(map[string]cfg.Release{
		"app":      {Chart: "example/app", ChartVersion: "1.0.0", Needs: []string{"database"}},
		"database": {Chart: "example/database", ChartVersion: "1.0.0"},
		"old":      {Chart: "example/old"},
	})
	current := testConfig(map[string]cfg.Release{
		"app":      {Chart: "example/app", ChartVersion: "2.0.0", Needs: []string{"database"}},
		"database": {Chart: "example/database", ChartVersion: "1.0.0"},
	})
	stubLoadedConfig(t, previous)

	items, err := CreatePlan(current, nil)
	if err != nil {
		t.Fatalf("CreatePlan returned an error: %v", err)
	}

	want := []PlanItem{
		{Group: 1, Action: PlanSync, Release: "database"},
		{Group: 2, Action: PlanUpgrade, Release: "app"},
		{Group: 3, Action: PlanUninstall, Release: "old"},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("unexpected plan: got %#v, want %#v", items, want)
	}
}

func TestCreatePlanSelectedReleaseIncludesDependenciesWithoutRemovals(t *testing.T) {
	previous := testConfig(map[string]cfg.Release{
		"database": {Chart: "example/database"},
		"old":      {Chart: "example/old"},
	})
	current := testConfig(map[string]cfg.Release{
		"app":      {Chart: "example/app", Needs: []string{"database"}},
		"database": {Chart: "example/database"},
	})
	stubLoadedConfig(t, previous)

	items, err := CreatePlan(current, []string{"app"})
	if err != nil {
		t.Fatalf("CreatePlan returned an error: %v", err)
	}

	want := []PlanItem{
		{Group: 1, Action: PlanSync, Release: "database"},
		{Group: 2, Action: PlanInstall, Release: "app"},
	}
	if !reflect.DeepEqual(items, want) {
		t.Fatalf("unexpected selected plan: got %#v, want %#v", items, want)
	}
}

func TestPlanToWritesStableTable(t *testing.T) {
	stubLoadedConfig(t, nil)
	config := testConfig(map[string]cfg.Release{
		"app": {Chart: "example/app"},
	})

	var output bytes.Buffer
	if err := PlanTo(&output, config, nil); err != nil {
		t.Fatalf("PlanTo returned an error: %v", err)
	}

	for _, expected := range []string{"GROUP", "ACTION", "RELEASE", "install", "app"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("plan output %q does not contain %q", output.String(), expected)
		}
	}
}

func stubLoadedConfig(t *testing.T, config *cfg.Config) {
	t.Helper()
	oldLoad := loadConfig
	loadConfig = func(*cfg.Config) (*cfg.Config, error) { return config, nil }
	t.Cleanup(func() { loadConfig = oldLoad })
}
