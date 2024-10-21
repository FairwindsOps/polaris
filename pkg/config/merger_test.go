package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var defaults = `
checks:
  deploymentMissingReplicas: warning
  priorityClassNotSet: warning
  tagNotSpecified: danger
existing:
  sub:
    key: value
`

var overrides = `
checks:
  pullPolicyNotAlways: ignore
  tagNotSpecified: overrides
existing:
  sub:
    key1: value1
  new: value
new:
  key: value
`

func TestMergeYaml(t *testing.T) {
	mergedContent, err := mergeYaml([]byte(defaults), []byte(overrides))
	assert.NoError(t, err)

	expectedYAML := `checks:
    deploymentMissingReplicas: warning
    priorityClassNotSet: warning
    pullPolicyNotAlways: ignore
    tagNotSpecified: overrides
existing:
    new: value
    sub:
        key: value
        key1: value1
new:
    key: value
`

	assert.Equal(t, expectedYAML, string(mergedContent))
}
