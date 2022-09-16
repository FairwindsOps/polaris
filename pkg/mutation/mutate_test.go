package mutation

import (
	"strings"
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

var testCases = []struct {
	original string
	mutated  string
	patch    config.Mutation
	message  string
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
}, {
	original: `
obj:
  foo:
    bar:
      - a
      - b
  baz: quux
`,
	patch: config.Mutation{
		Op: "replace",
		Value: map[string]interface{}{
			"bar": []string{"c", "d"},
		},
		Path: "/obj/foo",
	},
	mutated: `obj:
  foo:
    bar:
      - c
      - d
  baz: quux
`,
}, {
	original: `
foo: bar
`,
	patch: config.Mutation{
		Op:      "replace",
		Value:   "baz",
		Path:    "/foo",
		Comment: "# We set this to baz",
	},
	mutated: `
foo: baz # We set this to baz
`,
	message: "Expected a comment to appear",
}, {
	original: `
foo: bar
`,
	patch: config.Mutation{
		Op: "add",
		Value: map[string]interface{}{
			"baz": "quux",
		},
		Path:    "/extra",
		Comment: "# These are extra things",
	},
	mutated: `
foo: bar
extra:
  # These are extra things
  baz: quux
`,
	message: "Expected a comment to appear next to an object",
}}

func TestApplyAllMutations(t *testing.T) {
	for _, tc := range testCases {
		mutated, err := ApplyAllMutations(tc.original, []config.Mutation{tc.patch})
		assert.NoError(t, err)
		assert.EqualValues(t, strings.TrimSpace(tc.mutated), strings.TrimSpace(mutated), tc.message)
	}
}
