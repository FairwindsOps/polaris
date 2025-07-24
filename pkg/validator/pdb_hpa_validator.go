package validator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/qri-io/jsonschema"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func init() {
	registerCustomChecks("pdbMinAvailableGreaterThanHPAMinReplicas", pdbMinAvailableGreaterThanHPAMinReplicas)
}

func pdbMinAvailableGreaterThanHPAMinReplicas(test schemaTestCase) (bool, []jsonschema.KeyError, error) {
	if test.ResourceProvider == nil {
		return true, nil, nil
	}

	deployment := &appsv1.Deployment{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(test.Resource.Resource.Object, deployment)
	if err != nil {
		logrus.Warnf("error converting unstructured to Deployment: %v", err)
		return true, nil, nil
	}

	attachedPDB, err := hasPDBAttached(*deployment, test.ResourceProvider.Resources["policy/PodDisruptionBudget"])
	if err != nil {
		logrus.Warnf("error getting PodDisruptionBudget: %v", err)
		return true, nil, nil
	}

	attachedHPA, err := hasHPAAttached(*deployment, test.ResourceProvider.Resources["autoscaling/HorizontalPodAutoscaler"])
	if err != nil {
		logrus.Warnf("error getting HorizontalPodAutoscaler: %v", err)
		return true, nil, nil
	}

	if attachedPDB != nil && attachedHPA != nil {
		logrus.Debugf("both PDB and HPA are attached to deployment %s", deployment.Name)

		if attachedPDB.Spec.MinAvailable == nil {
			return true, nil, nil
		}

		pdbMinAvailable, isPercent, err := getIntOrPercentValueSafely(attachedPDB.Spec.MinAvailable)
		if err != nil {
			logrus.Warnf("error getting getIntOrPercentValueSafely: %v", err)
			return true, nil, nil
		}

		if isPercent {
			// if the value is a percentage, we need to calculate the actual value
			if attachedHPA.Spec.MinReplicas == nil {
				return true, nil, nil
			}

			pdbMinAvailable, err = intstr.GetScaledValueFromIntOrPercent(attachedPDB.Spec.MinAvailable, int(*attachedHPA.Spec.MinReplicas), true)
			if err != nil {
				logrus.Warnf("error getting minAvailable value from PodDisruptionBudget: %v", err)
				return true, nil, nil
			}
		}

		if attachedHPA.Spec.MinReplicas != nil && pdbMinAvailable > int(*attachedHPA.Spec.MinReplicas) {
			return false, []jsonschema.KeyError{
				{
					PropertyPath: "spec.minAvailable",
					InvalidValue: pdbMinAvailable,
					Message:      fmt.Sprintf("The minAvailable value in the PodDisruptionBudget(%s) is %d, which is greater than the minReplicas value in the HorizontalPodAutoscaler(%s) (%d)", attachedPDB.Name, pdbMinAvailable, attachedHPA.Name, *attachedHPA.Spec.MinReplicas),
				},
			}, nil
		}
	}

	return true, nil, nil
}

func hasPDBAttached(deployment appsv1.Deployment, pdbs []kube.GenericResource) (*policyv1.PodDisruptionBudget, error) {
	for _, generic := range pdbs {
		pdb := &policyv1.PodDisruptionBudget{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(generic.Resource.Object, pdb)
		if err != nil {
			return nil, fmt.Errorf("error converting unstructured to PodDisruptionBudget: %v", err)
		}

		if pdb.Spec.Selector == nil {
			continue
		}

		if matchesPDBForDeployment(deployment.Spec.Template.Labels, pdb.Spec.Selector.MatchLabels) {
			return pdb, nil
		}
	}
	return nil, nil
}

// matchesPDBForDeployment checks if the labels of the deployment match the labels of the PDB
func matchesPDBForDeployment(deploymentLabels, pdbLabels map[string]string) bool {
	for key, value := range pdbLabels {
		if deploymentLabels[key] == value {
			return true
		}
	}
	return false
}

func hasHPAAttached(deployment appsv1.Deployment, hpas []kube.GenericResource) (*autoscalingv1.HorizontalPodAutoscaler, error) {
	for _, generic := range hpas {
		hpa := &autoscalingv1.HorizontalPodAutoscaler{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(generic.Resource.Object, hpa)
		if err != nil {
			return nil, fmt.Errorf("error converting unstructured to HorizontalPodAutoscaler: %v", err)
		}

		if hpa.Spec.ScaleTargetRef.Kind == "Deployment" && hpa.Spec.ScaleTargetRef.Name == deployment.Name {
			return hpa, nil
		}
	}
	return nil, nil
}

// getIntOrPercentValueSafely is a safer version of getIntOrPercentValue based on private function intstr.getIntOrPercentValueSafely
func getIntOrPercentValueSafely(intOrStr *intstr.IntOrString) (int, bool, error) {
	switch intOrStr.Type {
	case intstr.Int:
		return intOrStr.IntValue(), false, nil
	case intstr.String:
		isPercent := false
		s := intOrStr.StrVal
		if strings.HasSuffix(s, "%") {
			isPercent = true
			s = strings.TrimSuffix(intOrStr.StrVal, "%")
		} else {
			return 0, false, fmt.Errorf("invalid type: string is not a percentage")
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0, false, fmt.Errorf("invalid value %q: %v", intOrStr.StrVal, err)
		}
		return int(v), isPercent, nil
	}
	return 0, false, fmt.Errorf("invalid type: neither int nor percentage")
}
