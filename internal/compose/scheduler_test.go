package compose

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	cfg "github.com/seacrew/helm-compose/internal/config"
)

func TestOperationGroupsRespectDependencies(t *testing.T) {
	operations := []operation{
		{name: "app", needs: []string{"database", "cache"}},
		{name: "cache"},
		{name: "database"},
		{name: "worker", needs: []string{"database"}},
	}

	groups, err := operationGroups(operations)
	if err != nil {
		t.Fatalf("operationGroups returned an error: %v", err)
	}

	got := operationGroupNames(groups)
	want := [][]string{{"cache", "database"}, {"app", "worker"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected operation groups: got %#v, want %#v", got, want)
	}
}

func TestReleaseOperationsReverseDependenciesForUninstall(t *testing.T) {
	releases := map[string]cfg.Release{
		"app":      {Needs: []string{"database"}},
		"database": {},
	}
	operations := releaseOperations(releases, []string{"app", "database"}, true, func(context.Context, string, *cfg.Release) error { return nil })
	groups, err := operationGroups(operations)
	if err != nil {
		t.Fatalf("operationGroups returned an error: %v", err)
	}

	got := operationGroupNames(groups)
	want := [][]string{{"app"}, {"database"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected uninstall groups: got %#v, want %#v", got, want)
	}
}

func TestRunOperationsHonorsConcurrencyLimit(t *testing.T) {
	started := make(chan string, 3)
	release := make(chan struct{})
	operations := make([]operation, 0, 3)
	for _, name := range []string{"a", "b", "c"} {
		name := name
		operations = append(operations, operation{name: name, run: func(context.Context) error {
			started <- name
			<-release
			return nil
		}})
	}

	done := make(chan error, 1)
	go func() {
		done <- runOperations(context.Background(), operations, 2)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for bounded operations to start")
		}
	}
	select {
	case name := <-started:
		t.Fatalf("release %q started above the concurrency limit", name)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("runOperations returned an error: %v", err)
	}
}

func TestRunOperationsCancelsPeersAndSkipsLaterGroups(t *testing.T) {
	started := make(chan string, 2)
	fail := make(chan struct{})
	canceled := make(chan struct{}, 1)
	var mutex sync.Mutex
	lateStarted := false

	operations := []operation{
		{name: "a", run: func(context.Context) error {
			started <- "a"
			<-fail
			return errors.New("boom")
		}},
		{name: "b", run: func(ctx context.Context) error {
			started <- "b"
			<-ctx.Done()
			canceled <- struct{}{}
			return ctx.Err()
		}},
		{name: "c", needs: []string{"a", "b"}, run: func(context.Context) error {
			mutex.Lock()
			lateStarted = true
			mutex.Unlock()
			return nil
		}},
	}

	done := make(chan error, 1)
	go func() {
		done <- runOperations(context.Background(), operations, 2)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for concurrent operations to start")
		}
	}
	close(fail)

	err := <-done
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected the original operation error, got %v", err)
	}
	select {
	case <-canceled:
	default:
		t.Fatal("expected the peer operation to receive cancellation")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if lateStarted {
		t.Fatal("an operation in a later dependency group was started")
	}
}

func operationGroupNames(groups [][]operation) [][]string {
	result := make([][]string, len(groups))
	for i, group := range groups {
		for _, operation := range group {
			result[i] = append(result[i], operation.name)
		}
	}
	return result
}
