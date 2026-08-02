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
	"fmt"
	"sync"

	cfg "github.com/seacrew/helm-compose/internal/config"
	prov "github.com/seacrew/helm-compose/internal/provider"
	"github.com/seacrew/helm-compose/internal/util"
)

var (
	loadConfig  = prov.Load
	storeConfig = prov.Store
)

func RunUp(config *cfg.Config, releases []string) error {
	releaseNames, err := selectReleaseNames(config.Releases, releases)
	if err != nil {
		return err
	}

	for name, url := range config.Repositories {
		if err := addHelmRepository(name, url); err != nil {
			return err
		}
	}

	previousConfig, err := loadConfig(config)
	if err != nil {
		return err
	}

	operations := make([]func() error, 0, len(releaseNames))
	for _, name := range releaseNames {
		name := name
		release := config.Releases[name]
		operations = append(operations, func() error {
			if err := installHelmRelease(name, &release); err != nil {
				return fmt.Errorf("release %q: %w", name, err)
			}
			return nil
		})
	}

	if len(releases) == 0 && previousConfig != nil {
		for name, release := range previousConfig.Releases {
			if _, ok := config.Releases[name]; ok {
				continue
			}

			name := name
			release := release
			operations = append(operations, func() error {
				if err := uninstallHelmRelease(name, &release); err != nil {
					return fmt.Errorf("release %q: %w", name, err)
				}
				return nil
			})
		}
	}

	if err := runConcurrently(operations); err != nil {
		return err
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
	previousConfig, err := loadConfig(config)
	if err != nil {
		return err
	}
	storage := config.Storage

	if previousConfig != nil {
		config = previousConfig
	}

	releaseNames, err := selectReleaseNames(config.Releases, releases)
	if err != nil {
		return err
	}

	operations := make([]func() error, 0, len(releaseNames))
	for _, name := range releaseNames {
		name := name
		release := config.Releases[name]
		operations = append(operations, func() error {
			if err := uninstallHelmRelease(name, &release); err != nil {
				return fmt.Errorf("release %q: %w", name, err)
			}
			return nil
		})
	}

	if err := runConcurrently(operations); err != nil {
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

func selectReleaseNames(releases map[string]cfg.Release, selected []string) ([]string, error) {
	if len(selected) == 0 {
		names := make([]string, 0, len(releases))
		for name := range releases {
			names = append(names, name)
		}
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

	return names, nil
}

func runConcurrently(operations []func() error) error {
	var wg sync.WaitGroup
	errorsChannel := make(chan error, len(operations))

	for _, operation := range operations {
		wg.Add(1)
		go func(operation func() error) {
			defer wg.Done()
			if err := operation(); err != nil {
				errorsChannel <- err
			}
		}(operation)
	}

	wg.Wait()
	close(errorsChannel)

	var operationErrors []error
	for err := range errorsChannel {
		operationErrors = append(operationErrors, err)
	}

	return errors.Join(operationErrors...)
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
	util.PrintColors = false

	for name, url := range config.Repositories {
		if err := addHelmRepository(name, url); err != nil {
			return err
		}
	}

	for name, release := range config.Releases {
		if len(releases) == 0 {
			if err := templateHelmRelease(name, &release); err != nil {
				return fmt.Errorf("release %q: %w", name, err)
			}
			continue
		}

		for _, rel := range releases {
			if rel == name {
				if err := templateHelmRelease(name, &release); err != nil {
					return fmt.Errorf("release %q: %w", name, err)
				}
			}
		}
	}

	return nil
}
