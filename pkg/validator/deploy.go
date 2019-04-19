// Copyright 2019 ReactiveOps
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
)

// ValidateDeploy validates a single deployment, returns a ResourceResult.
func ValidateDeploy(conf conf.Configuration, deploy *appsv1.Deployment) ResourceResult {
	pod := deploy.Spec.Template.Spec
	resResult := ValidatePod(conf, &pod)
	resResult.Name = deploy.Name
	resResult.Type = "Deployment"
	return resResult
}

// ValidateDeploys validates that each deployment conforms to the Fairwinds config,
// returns a list of ResourceResults organized by namespace.
func ValidateDeploys(config conf.Configuration, k8sAPI *kube.API) (NamespacedResults, error) {
	nsResults := NamespacedResults{}
	deploys, err := k8sAPI.GetDeploys()
	if err != nil {
		return nsResults, err
	}

	for _, deploy := range deploys.Items {
		resResult := ValidateDeploy(config, &deploy)
		nsResults = addResult(resResult, nsResults, deploy.Namespace)
	}

	return nsResults, nil
}

func addResult(resResult ResourceResult, nsResults NamespacedResults, nsName string) NamespacedResults {
	nsResult := &NamespacedResult{}

	// If there is already data stored for this namespace name,
	// then append to the ResourceResults to the existing data.
	switch nsResults[nsName] {
	case nil:
		nsResult = &NamespacedResult{
			Summary: &ResultSummary{},
			Results: []ResourceResult{},
		}
		nsResults[nsName] = nsResult
	default:
		nsResult = nsResults[nsName]
	}

	nsResult.Results = append(nsResult.Results, resResult)
	nsResult.Summary.appendResults(*resResult.Summary)

	return nsResults
}
