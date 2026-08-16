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
package cmd

import (
	"github.com/seacrew/helm-compose/internal/compose"
	"github.com/seacrew/helm-compose/internal/config"
	"github.com/spf13/cobra"
)

var planReleases []string

var planCmd = &cobra.Command{
	Use:     "plan [RELEASE ...]",
	Aliases: []string{"diff"},
	Short:   "Show the release actions Helm Compose would perform without changing the cluster.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		config, err := config.ParseComposeFile(composeFile)
		if err != nil {
			return err
		}

		releases := append([]string{}, args...)
		releases = append(releases, planReleases...)
		return compose.PlanTo(cmd.OutOrStdout(), config, releases)
	},
}

func init() {
	planCmd.Flags().StringSliceVarP(&planReleases, "release", "r", nil, "Release name to plan (can be specified multiple times)")
	rootCmd.AddCommand(planCmd)
}
