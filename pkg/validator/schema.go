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
	"github.com/fairwindsops/polaris/pkg/validator/controllers"
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

func resolveCheck(conf *config.Configuration, checkID string, controllerName string, controllerKind config.SupportedController, target config.TargetKind, isInitContainer bool) (*config.SchemaCheck, error) {
	check, ok := conf.CustomChecks[checkID]
	if !ok {
		check, ok = builtInChecks[checkID]
	}
	if !ok {
		return nil, fmt.Errorf("Check %s not found", checkID)
	}
	if !conf.IsActionable(check.ID, controllerName) {
		return nil, nil
	}
	if !check.IsActionable(target, controllerKind, isInitContainer) {
		return nil, nil
	}
	return &check, nil
}

func makeResult(conf *config.Configuration, check *config.SchemaCheck, passes bool) ResultMessage {
	result := ResultMessage{
		ID:       check.ID,
		Severity: conf.Checks[check.ID],
		Category: check.Category,
		Success:  passes,
	}
	if passes {
		result.Message = check.SuccessMessage
	} else {
		result.Message = check.FailureMessage
	}
	return result
}

func applyPodSchemaChecks(conf *config.Configuration, controller controllers.Interface) (ResultSet, error) {
	results := ResultSet{}
	checkIDs := getSortedKeys(conf.Checks)
	for _, checkID := range checkIDs {
		check, err := resolveCheck(conf, checkID, controller.GetName(), controller.GetKind(), config.TargetPod, false)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		} else if check == nil {
			continue
		}
		passes, err := check.CheckPod(controller.GetPodSpec())
		if err != nil {
			return nil, err
		}
		results[check.ID] = makeResult(conf, check, passes)
	}
	return results, nil
}

func applyContainerSchemaChecks(conf *config.Configuration, controller controllers.Interface, container *corev1.Container, isInit bool) (ResultSet, error) {
	results := ResultSet{}
	checkIDs := getSortedKeys(conf.Checks)
	for _, checkID := range checkIDs {
		check, err := resolveCheck(conf, checkID, controller.GetName(), controller.GetKind(), config.TargetContainer, isInit)
		if err != nil {
			return nil, err
		} else if check == nil {
			continue
		}
		var passes bool
		if check.SchemaTarget == config.TargetPod {
			podCopy := *controller.GetPodSpec()
			podCopy.InitContainers = []corev1.Container{}
			podCopy.Containers = []corev1.Container{*container}
			passes, err = check.CheckPod(&podCopy)
		} else {
			passes, err = check.CheckContainer(container)
		}
		if err != nil {
			return nil, err
		}
		results[check.ID] = makeResult(conf, check, passes)
	}
	return results, nil
}

func getSortedKeys(m map[string]config.Severity) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
