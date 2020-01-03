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
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/validator/controllers"
)

// ValidatePod validates that each pod conforms to the Polaris config, returns a ResourceResult.
func ValidatePod(conf *config.Configuration, controller controllers.Interface) (PodResult, error) {
	podResults, err := applyPodSchemaChecks(conf, controller)
	if err != nil {
		return PodResult{}, err
	}

	pRes := PodResult{
		Results:          podResults,
		ContainerResults: []ContainerResult{},
	}

	pRes.ContainerResults, err = ValidateAllContainers(conf, controller)
	if err != nil {
		return pRes, err
	}

	return pRes, nil
}
