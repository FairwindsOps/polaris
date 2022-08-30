package mutation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gomodules.xyz/jsonpatch/v2"
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

var locationYaml = `pets:
  - name: fido
    owners:
      - name: Alice
        location: Denver
      - name: Bob
        aliases:
          - Robert
        location: Denver
  - name: scooby
`

var redactedYaml = `pets:
  - name: fido
    owners:
      - name: Alice
      - name: Bob
  - name: scooby
`
var aliasesYaml = `pets:
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
`

func TestApplyAllMutations(t *testing.T) {

	mutation := jsonpatch.Operation{
		Operation: "add",
		Value:     "Denver",
		Path:      "/pets/0/owners/*/location",
	}

	mutated, err := ApplyAllMutations(oldYaml, []jsonpatch.Operation{mutation})
	assert.NoError(t, err)
	assert.EqualValues(t, locationYaml, mutated)

	mutation = jsonpatch.Operation{
		Operation: "remove",
		Path:      "/pets/0/owners/*/aliases",
	}

	mutated, err = ApplyAllMutations(oldYaml, []jsonpatch.Operation{mutation})
	assert.NoError(t, err)
	assert.EqualValues(t, redactedYaml, mutated)

	mutation = jsonpatch.Operation{
		Operation: "add",
		Value:     "rob",
		Path:      "/pets/0/owners/*/aliases/-",
	}
	mutated, err = ApplyAllMutations(oldYaml, []jsonpatch.Operation{mutation})
	assert.NoError(t, err)
	assert.EqualValues(t, aliasesYaml, mutated)
}
