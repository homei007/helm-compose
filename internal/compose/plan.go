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
	"io"
	"reflect"
	"sort"
	"text/tabwriter"

	cfg "github.com/seacrew/helm-compose/internal/config"
)

type PlanAction string

const (
	PlanInstall   PlanAction = "install"
	PlanUpgrade   PlanAction = "upgrade"
	PlanUninstall PlanAction = "uninstall"
	PlanSync      PlanAction = "sync"
)

type PlanItem struct {
	Group   int
	Action  PlanAction
	Release string
}

func CreatePlan(config *cfg.Config, releases []string) ([]PlanItem, error) {
	if err := cfg.ValidateReleaseDependencies(config.Releases); err != nil {
		return nil, err
	}

	previousConfig, err := loadConfig(config)
	if err != nil {
		return nil, err
	}

	releaseNames, err := selectReleaseNames(config.Releases, releases, true)
	if err != nil {
		return nil, err
	}

	currentOperations := releaseOperations(config.Releases, releaseNames, false, func(_ context.Context, _ string, _ *cfg.Release) error {
		return nil
	})
	currentGroups, err := operationGroups(currentOperations)
	if err != nil {
		return nil, err
	}

	items := make([]PlanItem, 0, len(releaseNames))
	for groupIndex, group := range currentGroups {
		for _, operation := range group {
			action := PlanInstall
			if previousConfig != nil {
				if previousRelease, exists := previousConfig.Releases[operation.name]; exists {
					action = PlanUpgrade
					if reflect.DeepEqual(previousRelease, config.Releases[operation.name]) {
						action = PlanSync
					}
				}
			}
			items = append(items, PlanItem{Group: groupIndex + 1, Action: action, Release: operation.name})
		}
	}

	if len(releases) == 0 && previousConfig != nil {
		removedNames := make([]string, 0)
		for name := range previousConfig.Releases {
			if _, exists := config.Releases[name]; !exists {
				removedNames = append(removedNames, name)
			}
		}
		sort.Strings(removedNames)
		removedOperations := releaseOperations(previousConfig.Releases, removedNames, true, func(_ context.Context, _ string, _ *cfg.Release) error {
			return nil
		})
		removedGroups, err := operationGroups(removedOperations)
		if err != nil {
			return nil, err
		}
		for groupIndex, group := range removedGroups {
			for _, operation := range group {
				items = append(items, PlanItem{
					Group:   len(currentGroups) + groupIndex + 1,
					Action:  PlanUninstall,
					Release: operation.name,
				})
			}
		}
	}

	return items, nil
}

func PlanTo(writer io.Writer, config *cfg.Config, releases []string) error {
	items, err := CreatePlan(config, releases)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		_, err := fmt.Fprintln(writer, "No release actions planned.")
		return err
	}

	table := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, "GROUP\tACTION\tRELEASE"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(table, "%d\t%s\t%s\n", item.Group, item.Action, item.Release); err != nil {
			return err
		}
	}
	if err := table.Flush(); err != nil {
		return err
	}
	for _, item := range items {
		if item.Action == PlanSync {
			_, err := fmt.Fprintln(writer, "\nsync means the Compose configuration is unchanged; Helm will still reconcile the release.")
			return err
		}
	}
	return nil
}
