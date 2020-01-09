package validator

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	packr "github.com/gobuffalo/packr/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/fairwindsops/polaris/pkg/config"
)

var (
	schemaBox     = (*packr.Box)(nil)
	builtInChecks = map[string]config.SchemaCheck{}
	// We explicitly set the order to avoid thrash in the
	// tests as we migrate toward JSON schema
	checkOrder = []string{
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
		"notReadOnlyRootFileSystem",
		"privilegeEscalationAllowed",
		"dangerousCapabilities",
		"insecureCapabilities",
	}
)

func init() {
	schemaBox = packr.New("Schemas", "../../checks")
	for _, checkID := range checkOrder {
		contents, err := schemaBox.Find(checkID + ".yaml")
		if err != nil {
			panic(err)
		}
		check, err := parseCheck(contents)
		if err != nil {
			panic(err)
		}
		check.ID = checkID
		builtInChecks[checkID] = check
	}
}

func parseCheck(rawBytes []byte) (config.SchemaCheck, error) {
	reader := bytes.NewReader(rawBytes)
	check := config.SchemaCheck{}
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&check); err != nil {
			if err == io.EOF {
				return check, nil
			}
			return check, fmt.Errorf("Decoding schema check failed: %v", err)
		}
	}
}

func applyPodSchemaChecks(conf *config.Configuration, pod *corev1.PodSpec, controllerName string, controllerType config.SupportedController, pv *PodValidation) error {
	checkIDs := getSortedKeys(conf.Checks)
	for _, checkID := range checkIDs {
		check, ok := conf.CustomChecks[checkID]
		if !ok {
			check, ok = builtInChecks[checkID]
		}
		if !ok {
			return fmt.Errorf("Check %s not found", checkID)
		}
		if !conf.IsActionable(check.ID, controllerName) {
			continue
		}
		if !check.IsActionable(config.TargetPod, controllerType, false) {
			continue
		}
		passes, err := check.CheckPod(pod)
		if err != nil {
			return err
		}
		if passes {
			pv.addSuccess(check.SuccessMessage, check.Category, check.ID)
		} else {
			severity := conf.Checks[checkID]
			pv.addFailure(check.FailureMessage, severity, check.Category, check.ID)
		}
	}
	return nil
}

func applyContainerSchemaChecks(conf *config.Configuration, controllerName string, controllerType config.SupportedController, cv *ContainerValidation) error {
	checkIDs := getSortedKeys(conf.Checks)
	for _, checkID := range checkIDs {
		check, ok := conf.CustomChecks[checkID]
		if !ok {
			check, ok = builtInChecks[checkID]
		}
		if !ok {
			return fmt.Errorf("Check %s not found", checkID)
		}
		if !conf.IsActionable(check.ID, controllerName) {
			continue
		}
		if !check.IsActionable(config.TargetContainer, controllerType, cv.IsInitContainer) {
			continue
		}
		var passes bool
		var err error
		if check.SchemaTarget == config.TargetPod {
			cv.parentPodSpec.Containers = []corev1.Container{*cv.Container}
			passes, err = check.CheckPod(&cv.parentPodSpec)
			cv.parentPodSpec.Containers = []corev1.Container{}
		} else {
			passes, err = check.CheckContainer(cv.Container)
		}
		if err != nil {
			return err
		}
		if passes {
			cv.addSuccess(check.SuccessMessage, check.Category, check.ID)
		} else {
			severity := conf.Checks[checkID]
			cv.addFailure(check.FailureMessage, severity, check.Category, check.ID)
		}
	}
	return nil
}

func getSortedKeys(m map[string]config.Severity) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
