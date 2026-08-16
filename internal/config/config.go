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
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v2"
)

var (
	V1_0 = semver.MustParse("1.0")
	V1_1 = semver.MustParse("1.1")
)

func findComposeConfig() []string {
	return findComposeConfigIn(".")
}

func findComposeConfigIn(directory string) []string {
	var files []string
	filenames := []string{
		"helm-compose.yaml",
		"helm-compose.yml",
		"helmcompose.yaml",
		"helmcompose.yml",
		"helmcompose",
		"compose.yaml",
		"compose.yml",
	}

	for _, filename := range filenames {
		path := filepath.Join(directory, filename)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			files = append(files, path)
		}
	}

	return files
}

func ParseComposeFile(filename string) (*Config, error) {
	var files []string
	if filename == "" {
		files = findComposeConfig()
	} else if filename == "-" {
		reader := bufio.NewReader(os.Stdin)

		data := []byte{}
		for {
			b, err := reader.ReadBytes('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			data = append(data, b...)
		}

		return parseComposeData(data)
	} else if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("provided compose file not found")
	} else {
		files = []string{filename}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no compose file found")
	}

	if len(files) > 1 {
		return nil, fmt.Errorf("expected only one compose file but found multiple: %v", files)
	}

	file, err := os.ReadFile(files[0])
	if err != nil {
		return nil, err
	}

	return parseComposeData(file)
}

func parseComposeData(data []byte) (*Config, error) {
	config := Config{}
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if err := validateCompose(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateCompose(config *Config) error {
	if config.Version == "" {
		return fmt.Errorf("missing apiVersion in config")
	}

	version, err := semver.NewVersion(config.Version)
	if err != nil {
		return fmt.Errorf("failed to parse apiVersion: %s", config.Version)
	}

	if version.LessThan(V1_0) {
		return fmt.Errorf("helm compose requires at least apiVersion 1.0 but got %s", config.Version)
	}

	if err := validateComposeFeatures(version, config); err != nil {
		return err
	}

	if err := ValidateReleaseDependencies(config.Releases); err != nil {
		return err
	}

	return nil
}

func validateComposeFeatures(version *semver.Version, config *Config) error {
	if err := validateCompose1_1(version, config); err != nil {
		return fmt.Errorf("apiVersion 1.1+ necessary: %s", err)
	}

	return nil
}

func validateCompose1_1(version *semver.Version, config *Config) error {
	if version.GreaterThan(V1_0) {
		return nil
	}

	for name, release := range config.Releases {
		if release.Wait {
			return fmt.Errorf("trying to use 'wait' in release '%s'", name)
		}
		if len(release.Needs) > 0 {
			return fmt.Errorf("trying to use 'needs' in release '%s'", name)
		}
	}

	return nil
}

func ValidateReleaseDependencies(releases map[string]Release) error {
	names := make([]string, 0, len(releases))
	for name := range releases {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		seen := map[string]struct{}{}
		for _, dependency := range releases[name].Needs {
			if dependency == name {
				return fmt.Errorf("release %q cannot depend on itself", name)
			}
			if _, ok := releases[dependency]; !ok {
				return fmt.Errorf("release %q depends on unknown release %q", name, dependency)
			}
			if _, ok := seen[dependency]; ok {
				return fmt.Errorf("release %q declares dependency %q more than once", name, dependency)
			}
			seen[dependency] = struct{}{}
		}
	}

	state := make(map[string]int, len(releases))
	stack := make([]string, 0, len(releases))
	var visit func(string) error
	visit = func(name string) error {
		switch state[name] {
		case 1:
			start := 0
			for i, item := range stack {
				if item == name {
					start = i
					break
				}
			}
			cycle := append(append([]string{}, stack[start:]...), name)
			return fmt.Errorf("release dependency cycle: %s", strings.Join(cycle, " -> "))
		case 2:
			return nil
		}

		state[name] = 1
		stack = append(stack, name)
		dependencies := append([]string{}, releases[name].Needs...)
		sort.Strings(dependencies)
		for _, dependency := range dependencies {
			if err := visit(dependency); err != nil {
				return err
			}
		}
		stack = stack[:len(stack)-1]
		state[name] = 2
		return nil
	}

	for _, name := range names {
		if err := visit(name); err != nil {
			return err
		}
	}

	return nil
}
