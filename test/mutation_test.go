package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/stretchr/testify/assert"
)

var configYaml = `
checks:
  pullPolicyNotAlways: warning
  hostIPCSet: danger
  hostPIDSet: danger
  hostNetworkSet: danger
  hostPortSet: warning
  deploymentMissingReplicas: warning
  priorityClassNotSet: ignore
  runAsRootAllowed: danger
  cpuRequestsMissing: warning
  cpuLimitsMissing: warning
  memoryRequestsMissing: warning
  memoryLimitsMissing: warning
  readinessProbeMissing: warning
  livenessProbeMissing: warning
`

func TestMutations(t *testing.T) {
	c, err := config.Parse([]byte(configYaml))
	assert.NoError(t, err)
	assert.Len(t, c.Mutations, 0)
	for mutationStr := range mutationTestCasesMap {
		if len(mutationTestCasesMap[mutationStr]) == 0 {
			panic("No test cases found for " + mutationStr)
		}
		for _, tc := range mutationTestCasesMap[mutationStr] {
			newConfig := c
			key := fmt.Sprintf("%s/%s", tc.check, strings.ReplaceAll(tc.filename, "failure", "mutated"))
			mutatedYamlContent, ok := mutatedYamlContentMap[key]
			assert.True(t, ok)
			assert.Len(t, tc.resources.Resources, 1)
			newConfig.Checks = map[string]config.Severity{}
			newConfig.Checks[mutationStr] = config.SeverityDanger
			newConfig.Mutations = []string{mutationStr}
			results, err := validator.ApplyAllSchemaChecksToResourceProvider(&newConfig, tc.resources)
			assert.NoError(t, err)
			assert.Len(t, results, 1)
			allMutations := mutation.GetMutationsFromResults(results)
			assert.Len(t, allMutations, 1)
			for _, resources := range tc.resources.Resources {
				assert.Len(t, resources, 1)
				key := fmt.Sprintf("%s/%s/%s", resources[0].Kind, resources[0].Resource.GetName(), resources[0].Resource.GetNamespace())
				mutations := allMutations[key]
				yamlContent, err := mutation.ApplyAllMutations(tc.manifest, mutations)
				assert.NoError(t, err)
				assert.EqualValues(t, mutatedYamlContent, yamlContent, "Mutation test case for "+tc.check+"/"+tc.filename+" failed")
			}
		}
	}
}
