package mutation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fairwindsops/polaris/pkg/config"
)

var oldYaml = `
pets:
  - name: fido
    owners:
      - name: Alice
      - name: Bob
        aliases:
        - Robert
  - name: scooby
`

var testCases = []struct{
	original string
	mutated string
	patch config.Mutation
}{{
	original: oldYaml,
	patch: config.Mutation{
		Op:    "add",
		Value: "Denver",
		Path:  "/pets/0/owners/*/location",
	},
	mutated: `pets:
  - name: fido
    owners:
      - name: Alice
        location: Denver
      - name: Bob
        aliases:
          - Robert
        location: Denver
  - name: scooby
`,
}, {
	original: oldYaml,
	patch: config.Mutation{
		Op:   "remove",
		Path: "/pets/0/owners/*/aliases",
	},
	mutated: `pets:
  - name: fido
    owners:
      - name: Alice
      - name: Bob
  - name: scooby
`,
}, {
	original: oldYaml,
	patch: config.Mutation{
		Op:    "add",
		Value: "rob",
		Path:  "/pets/0/owners/*/aliases/-",
	},
	mutated: `pets:
  - name: fido
    owners:
      - name: Alice
        aliases:
          - rob
      - name: Bob
        aliases:
          - Robert
          - rob
  - name: scooby
`,
}, {
	original: `
pets:
  - name: fido
`,
	patch: config.Mutation{
		Op:    "add",
		Value: "Alice",
		Path:  "/pets/0/owners/0/name",
	},
	mutated: `pets:
  - name: fido
    owners:
      - name: Alice
`,
}}

func TestApplyAllMutations(t *testing.T) {
	for _, tc := range testCases {
		mutated, err := ApplyAllMutations(tc.original, []config.Mutation{tc.patch})
		assert.NoError(t, err)
		assert.EqualValues(t, tc.mutated, mutated)
	}
}
