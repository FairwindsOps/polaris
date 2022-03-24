package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
)

var configYaml = `
checks:
  pullPolicyNotAlways: warning
mutations:
  - pullPolicyNotAlways
`

func TestMutations(t *testing.T) {
	c, err := config.Parse([]byte(configYaml))
	assert.NoError(t, err)
	assert.Len(t, c.Mutations, 1)

	for _, tc := range testCases {
		if tc.failure && funk.Contains(c.Mutations, tc.check) {
			key := fmt.Sprintf("%s/%s", tc.check, strings.ReplaceAll(tc.filename, "failure", "success"))
			successResources, ok := successResourceMap[key]
			assert.True(t, ok)
			assert.Len(t, tc.resources.Resources, 1)
			assert.Len(t, successResources.Resources, 1)
			results, err := validator.ApplyAllSchemaChecksToResourceProvider(&c, tc.resources)
			assert.NoError(t, err)
			assert.Len(t, results, 1)
			allMutations := mutation.GetMutationsFromResults(&c, results)
			assert.Len(t, allMutations, 1)
			for kind, resources := range tc.resources.Resources {
				key := fmt.Sprintf("%s/%s/%s", resources[0].Kind, resources[0].Resource.GetName(), resources[0].Resource.GetNamespace())
				mutations := allMutations[key]
				assert.Len(t, mutations, 1)
				mutated, err := mutation.ApplyAllSchemaMutations(&c, tc.resources, resources[0], mutations)
				assert.NoError(t, err)
				expected := successResources.Resources[kind][0]
				assert.Equal(t, expected.Resource.Object, mutated.Resource.Object)
			}
		}
	}
}
