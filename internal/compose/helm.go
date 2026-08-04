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
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	cfg "github.com/seacrew/helm-compose/internal/config"
	"github.com/seacrew/helm-compose/internal/util"
)

var (
	helm        = os.Getenv("HELM_BIN")
	versionRE   = regexp.MustCompile(`Version:\s*"([^"]+)"`)
	minVersion  = semver.MustParse("v3.0.0")
	executeHelm = util.Execute
)

type HelmCommand string

const (
	HELM_UPGRADE   HelmCommand = "upgrade"
	HELM_UNINSTALL HelmCommand = "uninstall"
	HELM_TEMPLATE  HelmCommand = "template"
)

func CompatibleHelmVersion() error {
	cmd := exec.Command(helm, "version")
	util.DebugPrint("Executing %s", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run `%s version`: %v", os.Getenv("HELM_BIN"), err)
	}

	versionOutput := string(output)

	matches := versionRE.FindStringSubmatch(versionOutput)
	if matches == nil {
		return fmt.Errorf("failed to find version in output %#v", versionOutput)
	}

	helmVersion, err := semver.NewVersion(matches[1])
	if err != nil {
		return fmt.Errorf("failed to parse version %#v: %v", matches[1], err)
	}

	if minVersion.GreaterThan(helmVersion) {
		return fmt.Errorf("helm compose requires at least helm version %s", minVersion.String())
	}
	return nil
}

func addHelmRepository(name string, url string) error {
	output, err := util.Execute(helm, "repo", "add", "--force-update", name, url)

	if err != nil {
		return errors.New(output)
	}

	return nil
}

func installHelmRelease(name string, release *cfg.Release) error {
	args, err := createHelmArguments(HELM_UPGRADE, name, release)
	if err != nil {
		return err
	}

	return helmExec(name, args)
}

func templateHelmRelease(name string, release *cfg.Release) error {
	args, err := createHelmArguments(HELM_TEMPLATE, name, release)
	if err != nil {
		return err
	}

	return helmExec("", args)
}

func uninstallHelmRelease(name string, release *cfg.Release) error {
	var args []string

	args = append(args, "uninstall")

	if release.Namespace != "" {
		args = append(args, fmt.Sprintf("--namespace=%s", release.Namespace))
	}

	if release.KubeConfig != "" {
		args = append(args, fmt.Sprintf("--kubeconfig=%s", release.KubeConfig))
	}

	if release.KubeContext != "" {
		args = append(args, fmt.Sprintf("--kube-context=%s", release.KubeContext))
	}

	if release.DeletionStrategy != "" {
		args = append(args, fmt.Sprintf("--cascade=%s", release.DeletionStrategy))
	}

	if release.DeletionTimeout != "" {
		args = append(args, fmt.Sprintf("--timeout=%s", release.DeletionTimeout))
	}

	if release.DeletionNoHooks {
		args = append(args, "--no-hooks")
	}

	if release.KeepHistory {
		args = append(args, "--keep-history")
	}

	args = append(args, name)

	return helmExec(name, args)
}

func createHelmArguments(command HelmCommand, name string, release *cfg.Release) ([]string, error) {
	var args []string

	args = append(args, string(command))

	if command == HELM_UPGRADE {
		args = append(args, "--install")
	}

	if release.ChartVersion != "" {
		args = append(args, fmt.Sprintf("--version=%s", release.ChartVersion))
	}

	if release.Namespace != "" {
		args = append(args, fmt.Sprintf("--namespace=%s", release.Namespace))
	}

	if release.ForceUpdate {
		args = append(args, "--force")
	}

	if release.HistoryMax < 0 {
		args = append(args, fmt.Sprintf("--history-max=%d", 0))
	} else if release.HistoryMax > 0 {
		args = append(args, fmt.Sprintf("--history-max=%d", release.HistoryMax))
	}

	if release.CreateNamespace {
		args = append(args, "--create-namespace")
	}

	if release.CleanUpOnFail {
		args = append(args, "--cleanup-on-fail")
	}

	if release.DependencyUpdate {
		args = append(args, "--dependency-update")
	}

	if release.SkipTLSVerify {
		args = append(args, "--insecure-skip-tls-verify")
	}

	if release.SkipCRDs {
		args = append(args, "--skip-crds")
	}

	if release.PostRenderer != "" && release.PostRendererPlugin != "" {
		return nil, fmt.Errorf("release %q cannot set both postRenderer and postRendererPlugin", name)
	}

	postRenderer := release.PostRenderer
	if release.PostRendererPlugin != "" {
		postRenderer = release.PostRendererPlugin
	}

	if postRenderer != "" {
		args = append(args, fmt.Sprintf("--post-renderer=%s", postRenderer))
	}

	if len(release.PostRendererArgs) > 0 {
		for _, postRendererArg := range release.PostRendererArgs {
			args = append(args, fmt.Sprintf("--post-renderer-args=%s", postRendererArg))
		}
	}

	if release.CAFile != "" {
		args = append(args, fmt.Sprintf("--ca-file=%s", release.CAFile))
	}

	if release.CertFile != "" {
		args = append(args, fmt.Sprintf("--cert-file=%s", release.CertFile))
	}

	if release.KeyFile != "" {
		args = append(args, fmt.Sprintf("--key-file=%s", release.KeyFile))
	}

	if release.Timeout != "" {
		args = append(args, fmt.Sprintf("--timeout=%s", release.Timeout))
	}

	if release.Wait {
		args = append(args, "--wait")
	}

	if release.KubeConfig != "" {
		args = append(args, fmt.Sprintf("--kubeconfig=%s", release.KubeConfig))
	}

	if release.KubeContext != "" {
		args = append(args, fmt.Sprintf("--kube-context=%s", release.KubeContext))
	}

	for _, file := range release.ValueFiles {
		args = append(args, fmt.Sprintf("--values=%s", file))
	}

	var jsonValues []string
	for key := range release.Values {
		data := util.ConvertJson(release.Values[key])
		values, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		jsonValues = append(jsonValues, fmt.Sprintf("%s=%s", key, values))
	}

	if len(jsonValues) > 0 {
		args = append(args, fmt.Sprintf("--set-json=%s", strings.Join(jsonValues, ",")))
	}

	args = append(args, name)
	args = append(args, release.Chart)

	return args, nil
}

func helmExec(name string, args []string) error {
	cp := util.NewColorPrinter(name)
	output, executeErr := executeHelm(helm, args...)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		if len(name) == 0 {
			fmt.Printf("%s\n", scanner.Text())
		} else {
			cp.Printf("%s |\t\t%s", name, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return executeErr
}
