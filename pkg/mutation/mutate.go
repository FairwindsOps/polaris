package mutation

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/thoas/go-funk"
)

// ApplyAllSchemaMutations applies available mutation to a single resource
func ApplyAllSchemaMutations(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource) (kube.GenericResource, error) {
	resByte := resource.OriginalObjectJSON
	var jsonByte []byte
	for checkID := range conf.Checks {
		if funk.Contains(conf.Mutations, checkID) {
			customCheck, err := resolveCheck(conf, checkID, resource)
			if err != nil {
				return resource, err
			}

			mutationByte, err := json.Marshal(customCheck.Mutations)
			if err != nil {
				return resource, err
			}

			patch, err := jsonpatch.DecodePatch(mutationByte)
			if err != nil {
				return resource, err
			}
			jsonByte, err = patch.Apply(resByte)
			if err != nil {
				return resource, err
			}

		}
	}
	mutated, err := kube.NewGenericResourceFromBytes(jsonByte)
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
