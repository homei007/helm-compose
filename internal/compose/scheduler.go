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
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

type operation struct {
	name  string
	needs []string
	run   func(context.Context) error
}

type operationResult struct {
	name string
	err  error
}

func operationGroups(operations []operation) ([][]operation, error) {
	byName := make(map[string]operation, len(operations))
	indegree := make(map[string]int, len(operations))
	dependents := make(map[string][]string, len(operations))

	for _, operation := range operations {
		if _, exists := byName[operation.name]; exists {
			return nil, fmt.Errorf("duplicate operation for release %q", operation.name)
		}
		byName[operation.name] = operation
		indegree[operation.name] = 0
	}

	for _, operation := range operations {
		seen := map[string]struct{}{}
		for _, dependency := range operation.needs {
			if _, exists := byName[dependency]; !exists {
				return nil, fmt.Errorf("release %q depends on release %q outside the operation set", operation.name, dependency)
			}
			if _, exists := seen[dependency]; exists {
				continue
			}
			seen[dependency] = struct{}{}
			indegree[operation.name]++
			dependents[dependency] = append(dependents[dependency], operation.name)
		}
	}

	ready := make([]string, 0, len(operations))
	for name, count := range indegree {
		if count == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)

	groups := make([][]operation, 0)
	processed := 0
	for len(ready) > 0 {
		groupNames := ready
		ready = nil
		group := make([]operation, 0, len(groupNames))

		for _, name := range groupNames {
			group = append(group, byName[name])
			processed++
		}
		groups = append(groups, group)

		for _, name := range groupNames {
			for _, dependent := range dependents[name] {
				indegree[dependent]--
				if indegree[dependent] == 0 {
					ready = append(ready, dependent)
				}
			}
		}
		sort.Strings(ready)
	}

	if processed != len(operations) {
		blocked := make([]string, 0)
		for name, count := range indegree {
			if count > 0 {
				blocked = append(blocked, name)
			}
		}
		sort.Strings(blocked)
		return nil, fmt.Errorf("release dependency cycle involving: %v", blocked)
	}

	return groups, nil
}

func runOperations(ctx context.Context, operations []operation, concurrency int) error {
	if concurrency < 0 {
		return fmt.Errorf("concurrency must be zero or greater")
	}

	groups, err := operationGroups(operations)
	if err != nil {
		return err
	}

	runContext, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, group := range groups {
		if err := runOperationGroup(runContext, cancel, group, concurrency); err != nil {
			return err
		}
	}

	return nil
}

func runOperationGroup(ctx context.Context, cancel context.CancelFunc, group []operation, concurrency int) error {
	if len(group) == 0 {
		return nil
	}

	workerCount := concurrency
	if workerCount == 0 || workerCount > len(group) {
		workerCount = len(group)
	}

	jobs := make(chan operation)
	results := make(chan operationResult, len(group))
	var workers sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case operation, ok := <-jobs:
					if !ok {
						return
					}
					if err := operation.run(ctx); err != nil {
						results <- operationResult{name: operation.name, err: err}
						cancel()
						return
					}
				}
			}
		}()
	}

sendLoop:
	for _, operation := range group {
		select {
		case jobs <- operation:
		case <-ctx.Done():
			break sendLoop
		}
	}
	close(jobs)
	workers.Wait()
	close(results)

	operationErrors := make([]operationResult, 0)
	for result := range results {
		operationErrors = append(operationErrors, result)
	}
	if len(operationErrors) == 0 {
		return ctx.Err()
	}

	sort.Slice(operationErrors, func(i, j int) bool {
		return operationErrors[i].name < operationErrors[j].name
	})
	errs := make([]error, 0, len(operationErrors))
	for _, result := range operationErrors {
		if errors.Is(result.err, context.Canceled) && len(operationErrors) > 1 {
			continue
		}
		errs = append(errs, fmt.Errorf("release %q: %w", result.name, result.err))
	}
	if len(errs) == 0 {
		return context.Canceled
	}

	return errors.Join(errs...)
}
