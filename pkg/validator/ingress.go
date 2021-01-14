// Copyright 2019 FairwindsOps Inc
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
	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"k8s.io/api/extensions/v1beta1"
)

// ValidateIngresses validates all the ingresses in a ResourceProvider
func ValidateIngresses(config *conf.Configuration, kubeResources *kube.ResourceProvider) ([]Result, error) {
	var results []Result
	for _, ingress := range kubeResources.Ingresses {
		result, err := ValidateIngress(config, ingress)
		if err != nil {
			return []Result{}, err
		}
		results = append(results, result)
	}
	return results, nil
}

// ValidateIngress validates a single ingress
func ValidateIngress(config *conf.Configuration, ingress v1beta1.Ingress) (Result, error) {
	results, err := applyIngressSchemaChecks(config, ingress)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Kind:      "Ingress",
		Name:      ingress.ObjectMeta.GetName(),
		Namespace: ingress.ObjectMeta.GetNamespace(),
		Results:   results,
	}

	return result, nil
}
