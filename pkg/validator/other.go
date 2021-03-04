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
)

// ValidateOtherKinds validates all the unstructured objects in a ResourceProvider
func ValidateOtherKinds(config *conf.Configuration, kubeResources *kube.ResourceProvider) ([]Result, error) {
	var results []Result
	for _, resources := range kubeResources.OtherKinds {
		for _, res := range resources {
			result, err := ValidateOtherKind(config, res)
			if err != nil {
				return []Result{}, err
			}
			results = append(results, result)
		}
	}
	return results, nil
}

// ValidateOtherKind validates a single unstructured object
func ValidateOtherKind(config *conf.Configuration, res kube.GenericResource) (Result, error) {
	results, err := applyTopLevelSchemaChecks(config, res, false)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Kind:      res.Kind,
		Name:      res.ObjectMeta.GetName(),
		Namespace: res.ObjectMeta.GetNamespace(),
		Results:   results,
	}

	return result, nil
}
