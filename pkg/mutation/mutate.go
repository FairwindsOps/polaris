package mutation

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
)

// ApplyAllSchemaMutations applies available mutation to a single resource
func ApplyAllSchemaMutations(conf *config.Configuration, resourceProvider *kube.ResourceProvider, resource kube.GenericResource, mutations []map[string]interface{}) (kube.GenericResource, error) {
	resByte := resource.OriginalObjectJSON
	var jsonByte []byte
	mutationByte, err := json.Marshal(mutations)
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

// GetMutationsFromResults returns all mutations from results
func GetMutationsFromResults(conf *config.Configuration, results []validator.Result) map[string][]map[string]interface{} {
	allMutationsFromResults := make(map[string][]map[string]interface{})
	for _, result := range results {
		key := fmt.Sprintf("%s/%s/%s", result.Kind, result.Name, result.Namespace)

		for _, resultMessage := range result.Results {
			if len(resultMessage.Mutations) > 0 {
				mutations, ok := allMutationsFromResults[key]
				if !ok {
					mutations = make([]map[string]interface{}, 0)
				}
				allMutationsFromResults[key] = append(mutations, resultMessage.Mutations...)
			}
		}

		for _, resultMessage := range result.PodResult.Results {
			if len(resultMessage.Mutations) > 0 {
				mutations, ok := allMutationsFromResults[key]
				if !ok {
					mutations = make([]map[string]interface{}, 0)
				}
				allMutationsFromResults[key] = append(mutations, resultMessage.Mutations...)
			}
		}

		for _, containerResult := range result.PodResult.ContainerResults {
			for _, resultMessage := range containerResult.Results {
				if len(resultMessage.Mutations) > 0 {
					mutations, ok := allMutationsFromResults[key]
					if !ok {
						mutations = make([]map[string]interface{}, 0)
					}
					allMutationsFromResults[key] = append(mutations, resultMessage.Mutations...)
				}
			}
		}

	}
	return allMutationsFromResults
}
