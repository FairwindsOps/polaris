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
	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
)

var (
	// BuiltInChecks contains the checks that come pre-installed w/ Polaris
	BuiltInChecks = map[string]SchemaCheck{}
	schemaBox     = (*packr.Box)(nil)
	// We explicitly set the order to avoid thrash in the
	// tests as we migrate toward JSON schema
	checkOrder = []string{
		// Controller Checks
		"deploymentMissingReplicas",
		// Pod checks
		"hostIPCSet",
		"hostPIDSet",
		"hostNetworkSet",
		"automountServiceAccountToken",
		// Container checks
		"memoryLimitsMissing",
		"memoryRequestsMissing",
		"cpuLimitsMissing",
		"cpuRequestsMissing",
		"readinessProbeMissing",
		"livenessProbeMissing",
		"pullPolicyNotAlways",
		"tagNotSpecified",
		"hostPortSet",
		"runAsRootAllowed",
		"runAsPrivileged",
		"notReadOnlyRootFilesystem",
		"privilegeEscalationAllowed",
		"dangerousCapabilities",
		"insecureCapabilities",
		"priorityClassNotSet",
		"linuxHardening",
		"sensitiveContainerEnvVar",
		// Other checks
		"tlsSettingsMissing",
		"pdbDisruptionsIsZero",
		"metadataAndNameMismatched",
		"missingPodDisruptionBudget",
		"missingNetworkPolicy",
		"sensitiveConfigmapContent",
	}
)

func init() {
	schemaBox = packr.New("Schemas", "../../checks")
	for _, checkID := range checkOrder {
		contents, err := schemaBox.Find(checkID + ".yaml")
		if err != nil {
			panic(err)
		}
		check, err := ParseCheck(checkID, contents)
		if err != nil {
			logrus.Errorf("Error while parsing check %s", checkID)
			panic(err)
		}
		BuiltInChecks[checkID] = check
	}
}
