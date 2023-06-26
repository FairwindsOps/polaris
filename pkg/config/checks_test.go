package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredFieldsOnBuiltInChecks(t *testing.T) {
	for _, v := range BuiltInChecks {
		assert.NotEmpty(t, v.SuccessMessage)
		assert.NotEmpty(t, v.FailureMessage)
		assert.NotEmpty(t, v.Category)
		assert.NotEmpty(t, v.Target)
	}
}
