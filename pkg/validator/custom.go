package validator

import (
	"sync"

	"github.com/qri-io/jsonschema"
)

type validatorFunction func(test schemaTestCase) (bool, []jsonschema.ValError, error)

var validatorMapper = map[string]validatorFunction{}
var lock = &sync.Mutex{}

func registerCustomChecks(name string, check validatorFunction) {
	lock.Lock()
	defer lock.Unlock()

	validatorMapper[name] = check
}
