// Copyright 2022 FairwindsOps, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsActionable determines whether a check is actionable given the current configuration
func (conf Configuration) IsActionable(ruleID string, objMeta metav1.Object, containerName string) bool {
	if severity, ok := conf.Checks[ruleID]; !ok || !severity.IsActionable() {
		return false
	}
	if conf.DisallowExemptions || conf.DisallowConfigExemptions {
		return true
	}
	for _, exemption := range conf.Exemptions {
		if exemption.Namespace != "" && exemption.Namespace != objMeta.GetNamespace() {
			continue
		}

		checkIfRuleMatches := false
		for _, rule := range exemption.Rules {
			if rule != ruleID {
				continue
			}
			checkIfRuleMatches = true
			break
		}

		if len(exemption.Rules) == 0 || checkIfRuleMatches {
			if !isExemptionCheckMatched(exemption.ControllerNames, objMeta.GetName()) {
				continue
			}
			if isExemptionCheckMatched(exemption.ContainerNames, containerName) {
				return false
			}
		}
	}
	return true
}

func isExemptionCheckMatched(arr []string, predicate string) bool {
	if len(arr) == 0 {
		return true
	}

	for _, container := range arr {
		if strings.HasPrefix(predicate, container) {
			return true
		}
	}
	return false
}
