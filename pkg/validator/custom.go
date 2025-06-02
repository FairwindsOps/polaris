package validator

import (
	"sync"

	"github.com/fairwindsops/polaris/pkg/config"
)

type validatorFunction func(test schemaTestCase) (bool, []config.ValError, error)

var validatorMapper = map[string]validatorFunction{}
var lock = &sync.Mutex{}

func registerCustomChecks(name string, check validatorFunction) {
	lock.Lock()
	defer lock.Unlock()

	validatorMapper[name] = check
}
