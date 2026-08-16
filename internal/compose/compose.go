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
	"fmt"
	"sort"

	cfg "github.com/seacrew/helm-compose/internal/config"
	prov "github.com/seacrew/helm-compose/internal/provider"
	"github.com/seacrew/helm-compose/internal/util"
)

var (
	loadConfig  = prov.Load
	storeConfig = prov.Store
)

func RunUp(config *cfg.Config, releases []string) error {
	return RunUpContext(context.Background(), config, releases, 0)
}

func RunUpContext(ctx context.Context, config *cfg.Config, releases []string, concurrency int) error {
	if err := cfg.ValidateReleaseDependencies(config.Releases); err != nil {
		return err
	}

	releaseNames, err := selectReleaseNames(config.Releases, releases, true)
	if err != nil {
		return err
	}

	if err := addHelmRepositories(ctx, config.Repositories); err != nil {
		return err
	}

	previousConfig, err := loadConfig(config)
	if err != nil {
		return err
	}

	operations := releaseOperations(config.Releases, releaseNames, false, func(ctx context.Context, name string, release *cfg.Release) error {
		return installHelmRelease(ctx, name, release)
	})
	if err := runOperations(ctx, operations, concurrency); err != nil {
		return err
	}

	if len(releases) == 0 && previousConfig != nil {
		removedNames := make([]string, 0)
		for name := range previousConfig.Releases {
			if _, ok := config.Releases[name]; ok {
				continue
			}
			removedNames = append(removedNames, name)
		}
		sort.Strings(removedNames)
		uninstallOperations := releaseOperations(previousConfig.Releases, removedNames, true, func(ctx context.Context, name string, release *cfg.Release) error {
			return uninstallHelmRelease(ctx, name, release)
		})
		if err := runOperations(ctx, uninstallOperations, concurrency); err != nil {
			return err
		}
	}

	state := config
	if len(releases) > 0 {
		state = mergeSelectedReleases(config, previousConfig, releaseNames)
	}

	if !state.Equal(previousConfig) {
		if err := storeConfig(state); err != nil {
			return err
		}
	}

	return nil
}

func RunDown(config *cfg.Config, releases []string) error {
	return RunDownContext(context.Background(), config, releases, 0)
}

func RunDownContext(ctx context.Context, config *cfg.Config, releases []string, concurrency int) error {
	previousConfig, err := loadConfig(config)
	if err != nil {
		return err
	}
	storage := config.Storage

	if previousConfig != nil {
		config = previousConfig
	}

	if err := cfg.ValidateReleaseDependencies(config.Releases); err != nil {
		return err
	}

	releaseNames, err := selectReleaseNames(config.Releases, releases, false)
	if err != nil {
		return err
	}

	if len(releases) > 0 && previousConfig != nil {
		state := cloneConfig(previousConfig)
		state.Storage = storage
		for _, name := range releaseNames {
			delete(state.Releases, name)
		}
		if err := cfg.ValidateReleaseDependencies(state.Releases); err != nil {
			return fmt.Errorf("cannot uninstall selected releases: %w", err)
		}
	}

	operations := releaseOperations(config.Releases, releaseNames, true, func(ctx context.Context, name string, release *cfg.Release) error {
		return uninstallHelmRelease(ctx, name, release)
	})
	if err := runOperations(ctx, operations, concurrency); err != nil {
		return err
	}

	if len(releases) > 0 && previousConfig != nil {
		state := cloneConfig(previousConfig)
		state.Storage = storage
		for _, name := range releaseNames {
			delete(state.Releases, name)
		}
		if !state.Equal(previousConfig) {
			if err := storeConfig(state); err != nil {
				return err
			}
		}
	}

	return nil
}

func selectReleaseNames(releases map[string]cfg.Release, selected []string, includeNeeds bool) ([]string, error) {
	if len(selected) == 0 {
		names := make([]string, 0, len(releases))
		for name := range releases {
			names = append(names, name)
		}
		sort.Strings(names)
		return names, nil
	}

	names := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, name := range selected {
		if _, ok := releases[name]; !ok {
			return nil, fmt.Errorf("release %q not found in compose configuration", name)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	if includeNeeds {
		var addDependencies func(string) error
		addDependencies = func(name string) error {
			dependencies := append([]string{}, releases[name].Needs...)
			sort.Strings(dependencies)
			for _, dependency := range dependencies {
				if _, ok := releases[dependency]; !ok {
					return fmt.Errorf("release %q depends on unknown release %q", name, dependency)
				}
				if _, ok := seen[dependency]; ok {
					continue
				}
				seen[dependency] = struct{}{}
				names = append(names, dependency)
				if err := addDependencies(dependency); err != nil {
					return err
				}
			}
			return nil
		}

		for _, name := range append([]string{}, names...) {
			if err := addDependencies(name); err != nil {
				return nil, err
			}
		}
	}

	sort.Strings(names)
	return names, nil
}

func releaseOperations(
	releases map[string]cfg.Release,
	names []string,
	reverse bool,
	run func(context.Context, string, *cfg.Release) error,
) []operation {
	selected := make(map[string]struct{}, len(names))
	operations := make(map[string]operation, len(names))
	for _, name := range names {
		selected[name] = struct{}{}
		name := name
		release := releases[name]
		operations[name] = operation{
			name: name,
			run: func(ctx context.Context) error {
				return run(ctx, name, &release)
			},
		}
	}

	for _, name := range names {
		for _, dependency := range releases[name].Needs {
			if _, ok := selected[dependency]; !ok {
				continue
			}
			if reverse {
				operation := operations[dependency]
				operation.needs = append(operation.needs, name)
				operations[dependency] = operation
			} else {
				operation := operations[name]
				operation.needs = append(operation.needs, dependency)
				operations[name] = operation
			}
		}
	}

	result := make([]operation, 0, len(names))
	for _, name := range names {
		operation := operations[name]
		sort.Strings(operation.needs)
		result = append(result, operation)
	}
	return result
}

func addHelmRepositories(ctx context.Context, repositories map[string]string) error {
	names := make([]string, 0, len(repositories))
	for name := range repositories {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := addHelmRepository(ctx, name, repositories[name]); err != nil {
			return err
		}
	}
	return nil
}

func mergeSelectedReleases(current, previous *cfg.Config, selected []string) *cfg.Config {
	state := cloneConfig(previous)
	state.Version = current.Version
	state.Storage = current.Storage
	state.Repositories = cloneRepositories(current.Repositories)

	for _, name := range selected {
		state.Releases[name] = current.Releases[name]
	}

	return state
}

func cloneConfig(config *cfg.Config) *cfg.Config {
	if config == nil {
		return &cfg.Config{Releases: map[string]cfg.Release{}}
	}

	clone := *config
	clone.Releases = make(map[string]cfg.Release, len(config.Releases))
	for name, release := range config.Releases {
		clone.Releases[name] = release
	}
	clone.Repositories = cloneRepositories(config.Repositories)
	return &clone
}

func cloneRepositories(repositories map[string]string) map[string]string {
	clone := make(map[string]string, len(repositories))
	for name, url := range repositories {
		clone[name] = url
	}
	return clone
}

func ListRevisions(config *cfg.Config) error {
	revisions, err := prov.List(config)
	if err != nil {
		return err
	}

	fmt.Printf("| Date             | Revision |\n")
	fmt.Printf("| ---------------- | -------- |\n")
	for _, rev := range revisions {
		fmt.Printf("| %d-%02d-%02d %02d:%02d | %8d |\n",
			rev.DateTime.Year(), rev.DateTime.Month(), rev.DateTime.Day(),
			rev.DateTime.Hour(), rev.DateTime.Minute(), rev.Revision)
	}

	return nil
}

func GetRevision(rev int, config *cfg.Config) error {
	revision, err := prov.Get(rev, config)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", *revision)

	return nil
}

func Template(config *cfg.Config, releases []string) error {
	return TemplateContext(context.Background(), config, releases)
}

func TemplateContext(ctx context.Context, config *cfg.Config, releases []string) error {
	util.PrintColors = false

	if err := addHelmRepositories(ctx, config.Repositories); err != nil {
		return err
	}

	releaseNames, err := selectReleaseNames(config.Releases, releases, false)
	if err != nil {
		return err
	}
	for _, name := range releaseNames {
		release := config.Releases[name]
		if err := templateHelmRelease(ctx, name, &release); err != nil {
			return fmt.Errorf("release %q: %w", name, err)
		}
	}

	return nil
}
