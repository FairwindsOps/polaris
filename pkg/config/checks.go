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
		"multipleReplicasForDeployment",
		// Pod checks
		"hostIPCSet",
		"hostPIDSet",
		"hostNetworkSet",
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
		// Other checks
		"tlsSettingsMissing",
		"pdbDisruptionsIsZero",
		"metadataAndNameMismatched",
		"missingPodDisruptionBudget",
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
