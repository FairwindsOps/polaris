package mutation

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/thoas/go-funk"
)

// ApplyAllSchemaMutations applies available mutation to a single resource
func ApplyAllSchemaMutations(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource) (kube.GenericResource, error) {
	resByte, err := resource.Resource.MarshalJSON()
	if err != nil {
		return resource, err
	}
	for checkID, _ := range conf.Checks {
		if funk.Contains(conf.Mutations, checkID) {
			customCheck, err := resolveCheck(conf, checkID, resource)
			if err != nil {
				return resource, err
			}
			for _, mutation := range customCheck.Mutation {
				patchJSON, err := json.Marshal(mutation)
				if err != nil {
					return resource, err
				}
				patch, err := jsonpatch.DecodePatch(patchJSON)
				if err != nil {
					return resource, err
				}
				resByte, err = patch.Apply(resByte)
				if err != nil {
					return resource, err
				}

			}
		}
	}
	mutated, err := kube.NewGenericResourceFromBytes(resByte)
	if err != nil {
		return resource, err
	}

	return mutated, nil
}

func resolveCheck(conf *config.Configuration, checkID string, resource kube.GenericResource) (*config.SchemaCheck, error) {
	check, ok := conf.CustomChecks[checkID]
	if !ok {
		check, ok = config.BuiltInChecks[checkID]
	}
	if !ok {
		return nil, fmt.Errorf("Check %s not found", checkID)
	}
	checkPtr, err := check.TemplateForResource(resource.Resource.Object)
	if err != nil {
		return nil, err
	}
	return checkPtr, nil
}
