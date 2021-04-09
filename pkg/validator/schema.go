package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/qri-io/jsonschema"
	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
)

var (
	schemaBox     = (*packr.Box)(nil)
	builtInChecks = map[string]config.SchemaCheck{}
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
		"pdbDisruptionsAllowedGreaterThanZero",
		"metadataAndNameMismatched",
	}
)

type schemaTestCase struct {
	Target          config.TargetKind
	Resource        kube.GenericResource
	IsInitContianer bool
	Container       *corev1.Container
}

func init() {
	schemaBox = packr.New("Schemas", "../../checks")
	for _, checkID := range checkOrder {
		contents, err := schemaBox.Find(checkID + ".yaml")
		if err != nil {
			panic(err)
		}
		check, err := config.ParseCheck(checkID, contents)
		if err != nil {
			panic(err)
		}
		builtInChecks[checkID] = check
	}
}

func resolveCheck(conf *config.Configuration, checkID string, test schemaTestCase) (*config.SchemaCheck, error) {
	if !conf.DisallowExemptions && hasExemptionAnnotation(test.Resource.ObjectMeta, checkID) {
		return nil, nil
	}
	check, ok := conf.CustomChecks[checkID]
	if !ok {
		check, ok = builtInChecks[checkID]
	}
	if !ok {
		return nil, fmt.Errorf("Check %s not found", checkID)
	}

	containerName := ""
	if test.Container != nil {
		containerName = test.Container.Name
	}
	if !conf.IsActionable(check.ID, test.Resource.ObjectMeta, containerName) {
		return nil, nil
	}
	if !check.IsActionable(test.Target, test.Resource.Kind, test.IsInitContianer) {
		return nil, nil
	}
	checkPtr, err := check.TemplateForResource(test.Resource.Resource.Object)
	if err != nil {
		return nil, err
	}
	return checkPtr, nil
}

func makeResult(conf *config.Configuration, check *config.SchemaCheck, passes bool, issues []jsonschema.ValError) ResultMessage {
	details := []string{}
	for _, issue := range issues {
		details = append(details, issue.Message)
	}
	result := ResultMessage{
		ID:       check.ID,
		Severity: conf.Checks[check.ID],
		Category: check.Category,
		Success:  passes,
		// FIXME: need to fix the tests before adding this back
		//Details: details,
	}
	if passes {
		result.Message = check.SuccessMessage
	} else {
		result.Message = check.FailureMessage
	}
	return result
}

const exemptionAnnotationKey = "polaris.fairwinds.com/exempt"
const exemptionAnnotationPattern = "polaris.fairwinds.com/%s-exempt"

func hasExemptionAnnotation(objMeta metaV1.Object, checkID string) bool {
	annot := objMeta.GetAnnotations()
	val := annot[exemptionAnnotationKey]
	if strings.ToLower(val) == "true" {
		return true
	}
	checkKey := fmt.Sprintf(exemptionAnnotationPattern, checkID)
	val = annot[checkKey]
	if strings.ToLower(val) == "true" {
		return true
	}
	return false
}

// ApplyAllSchemaChecksToAllResources applies available checks to a list of resources
func ApplyAllSchemaChecksToAllResources(conf *config.Configuration, resources []kube.GenericResource) ([]Result, error) {
	results := []Result{}
	for _, resource := range resources {
		result, err := ApplyAllSchemaChecks(conf, resource)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// ApplyAllSchemaChecks applies available checks to a single resource
func ApplyAllSchemaChecks(conf *config.Configuration, resource kube.GenericResource) (Result, error) {
	if resource.PodSpec == nil {
		return applyNonControllerSchemaChecks(conf, resource)
	}
	return applyControllerSchemaChecks(conf, resource)
}

func applyNonControllerSchemaChecks(conf *config.Configuration, resource kube.GenericResource) (Result, error) {
	finalResult := Result{
		Kind:      resource.Kind,
		Name:      resource.ObjectMeta.GetName(),
		Namespace: resource.ObjectMeta.GetNamespace(),
	}
	resultSet, err := applyTopLevelSchemaChecks(conf, resource, false)
	finalResult.Results = resultSet
	return finalResult, err
}

func applyControllerSchemaChecks(conf *config.Configuration, resource kube.GenericResource) (Result, error) {
	finalResult := Result{
		Kind:      resource.Kind,
		Name:      resource.ObjectMeta.GetName(),
		Namespace: resource.ObjectMeta.GetNamespace(),
	}
	resultSet, err := applyTopLevelSchemaChecks(conf, resource, true)
	if err != nil {
		return finalResult, err
	}
	finalResult.Results = resultSet

	podRS, err := applyPodSchemaChecks(conf, resource)
	if err != nil {
		return finalResult, err
	}
	podRes := PodResult{
		Results:          podRS,
		ContainerResults: []ContainerResult{},
	}
	finalResult.PodResult = &podRes

	for _, container := range resource.PodSpec.InitContainers {
		results, err := applyContainerSchemaChecks(conf, resource, &container, true)
		if err != nil {
			return finalResult, err
		}
		cRes := ContainerResult{
			Name:    container.Name,
			Results: results,
		}
		podRes.ContainerResults = append(podRes.ContainerResults, cRes)
	}
	for _, container := range resource.PodSpec.Containers {
		results, err := applyContainerSchemaChecks(conf, resource, &container, false)
		if err != nil {
			return finalResult, err
		}
		cRes := ContainerResult{
			Name:    container.Name,
			Results: results,
		}
		podRes.ContainerResults = append(podRes.ContainerResults, cRes)
	}

	return finalResult, nil
}

func applyTopLevelSchemaChecks(conf *config.Configuration, res kube.GenericResource, isController bool) (ResultSet, error) {
	test := schemaTestCase{
		Resource: res,
	}
	if isController {
		test.Target = config.TargetController
	}
	return applySchemaChecks(conf, test)
}

func applyPodSchemaChecks(conf *config.Configuration, controller kube.GenericResource) (ResultSet, error) {
	test := schemaTestCase{
		Target:   config.TargetPod,
		Resource: controller,
	}
	return applySchemaChecks(conf, test)
}

func applyContainerSchemaChecks(conf *config.Configuration, controller kube.GenericResource, container *corev1.Container, isInit bool) (ResultSet, error) {
	test := schemaTestCase{
		Target:          config.TargetContainer,
		Resource:        controller,
		Container:       container,
		IsInitContianer: isInit,
	}
	return applySchemaChecks(conf, test)
}

func applySchemaChecks(conf *config.Configuration, test schemaTestCase) (ResultSet, error) {
	results := ResultSet{}
	checkIDs := getSortedKeys(conf.Checks)
	for _, checkID := range checkIDs {
		result, err := applySchemaCheck(conf, checkID, test)
		if err != nil {
			return results, err
		}
		if result != nil {
			results[checkID] = *result
		}
	}
	return results, nil
}

func applySchemaCheck(conf *config.Configuration, checkID string, test schemaTestCase) (*ResultMessage, error) {
	check, err := resolveCheck(conf, checkID, test)
	if err != nil {
		return nil, err
	} else if check == nil {
		return nil, nil
	}
	var passes bool
	var issues []jsonschema.ValError
	if check.SchemaTarget != "" {
		if check.SchemaTarget == config.TargetPod && check.Target == config.TargetContainer {
			podCopy := *test.Resource.PodSpec
			podCopy.InitContainers = []corev1.Container{}
			podCopy.Containers = []corev1.Container{*test.Container}
			passes, issues, err = check.CheckPod(&podCopy)
		} else {
			return nil, fmt.Errorf("Unknown combination of target (%s) and schema target (%s)", check.Target, check.SchemaTarget)
		}
	} else if check.Target == config.TargetPod {
		passes, issues, err = check.CheckPod(test.Resource.PodSpec)
	} else if check.Target == config.TargetContainer {
		passes, issues, err = check.CheckContainer(test.Container)
	} else {
		passes, issues, err = check.CheckObject(test.Resource.Resource.Object)
	}
	if err != nil {
		return nil, err
	}
	result := makeResult(conf, check, passes, issues)
	return &result, nil
}

func getSortedKeys(m map[string]config.Severity) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
