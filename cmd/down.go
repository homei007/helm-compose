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

var downReleases []string

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down [RELEASE ...]",
	Short: "Uninstall releases defined in your compose file.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		if err := compose.CompatibleHelmVersion(); err != nil {
			return err
		}

		config, err := config.ParseComposeFile(composeFile)
		if err != nil {
			return err
		}

		releases := append([]string{}, args...)
		releases = append(releases, downReleases...)
		return compose.RunDown(config, releases)
	},
}

func init() {
	downCmd.Flags().StringSliceVarP(&downReleases, "release", "r", nil, "Release name to uninstall (can be specified multiple times)")
	rootCmd.AddCommand(downCmd)
}
