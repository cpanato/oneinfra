/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package pipelines

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/oneinfra/oneinfra/pkg/constants"
	"github.com/oneinfra/oneinfra/scripts/oi-releaser/pipelines/azure"
)

// AzureTest builds the Azure test pipeline
func AzureTest() error {
	pipeline := azure.Pipeline{
		Variables: map[string]string{
			"CI": "1",
		},
		Jobs: []azure.Job{
			{
				Job:         "build",
				DisplayName: "Build",
				Pool:        azure.DefaultPool,
				Steps: []azure.Step{
					{
						Bash:        "make pull-builder",
						DisplayName: "Pull builder image",
					},
					{
						Bash:        "make",
						DisplayName: "Build",
					},
				},
			},
			{
				Job:         "unit_and_integration_tests",
				DisplayName: "Unit and Integration tests",
				Pool:        azure.DefaultPool,
				Steps: []azure.Step{
					{
						Bash:        "make pull-builder",
						DisplayName: "Pull builder image",
					},
					{
						Bash:        "make test",
						DisplayName: "Test",
					},
				},
			},
		},
	}
	pipeline.Jobs = append(
		pipeline.Jobs,
		e2eTestsWithKubernetesVersion("default")...,
	)
	for _, kubernetesVersion := range constants.ReleaseData.KubernetesVersions {
		pipeline.Jobs = append(
			pipeline.Jobs,
			e2eTestsWithKubernetesVersion(kubernetesVersion.Version)...,
		)
	}
	marshaledPipeline, err := yaml.Marshal(&pipeline)
	if err != nil {
		return err
	}
	fmt.Println("# Code generated by oi-releaser. DO NOT EDIT.")
	azurifiedYAML := strings.ReplaceAll(
		string(marshaledPipeline),
		"- _",
		"- ",
	)
	fmt.Print(azurifiedYAML)
	return nil
}

func e2eTestsWithKubernetesVersion(kubernetesVersion string) []azure.Job {
	underscoredVersion := strings.ReplaceAll(
		strings.ReplaceAll(kubernetesVersion, ".", "_"),
		"-", "_",
	)
	return []azure.Job{
		{
			Job:         fmt.Sprintf("e2e_%s_with_local_cri_endpoints", underscoredVersion),
			DisplayName: fmt.Sprintf("e2e tests (%s) with local CRI endpoints", kubernetesVersion),
			Pool:        azure.DefaultPool,
			Steps: []azure.Step{
				{
					Bash:        "make deps",
					DisplayName: "Install host dependencies",
					Env: map[string]string{
						"KUBERNETES_VERSION": kubernetesVersion,
					},
				},
				{
					Bash:        "make e2e",
					DisplayName: "Run end to end tests",
					Env: map[string]string{
						"KUBERNETES_VERSION": kubernetesVersion,
					},
				},
			},
		},
		{
			Job:         fmt.Sprintf("e2e_%s_with_remote_cri_endpoints", underscoredVersion),
			DisplayName: fmt.Sprintf("e2e tests (%s) with remote CRI endpoints", kubernetesVersion),
			Pool:        azure.DefaultPool,
			Steps: []azure.Step{
				{
					Bash:        "make deps",
					DisplayName: "Install host dependencies",
					Env: map[string]string{
						"KUBERNETES_VERSION": kubernetesVersion,
					},
				},
				{
					Bash:        "make e2e-remote",
					DisplayName: "Run end to end tests",
					Env: map[string]string{
						"KUBERNETES_VERSION": kubernetesVersion,
					},
				},
			},
		},
	}
}
