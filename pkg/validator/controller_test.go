package validator

import (
	"testing"

	conf "github.com/reactiveops/polaris/pkg/config"
	"github.com/reactiveops/polaris/pkg/kube"
	"github.com/reactiveops/polaris/test"
	"github.com/stretchr/testify/assert"
)

func TestValidateController(t *testing.T) {
	k8s := test.SetupTestAPI()
	test.SetupAddControllers(k8s, "test")
	resources, err := kube.CreateResourceProviderFromAPI(k8s, "test")
	assert.Equal(t, err, nil, "error should be nil")

	c := conf.Configuration{}

	nsResults := &NamespacedResults{}
	ValidateControllers(c, resources, nsResults)
	assert.Equal(t, 1, len((*nsResults)["test"].DeploymentResults), "")
	assert.Equal(t, 1, len((*nsResults)["test"].StatefulSetResults), "")
}

func TestLimitRanges(t *testing.T) {
	k8s := test.SetupTestAPI()
	test.SetupAddControllers(k8s, "test")
	resources, err := kube.CreateResourceProviderFromAPI(k8s, "test")
	assert.Equal(t, err, nil, "error should be nil")

	c := conf.Configuration{
		Resources: conf.Resources{
			CPURequestsMissing:    conf.SeverityError,
			CPULimitsMissing:      conf.SeverityError,
			MemoryRequestsMissing: conf.SeverityError,
			MemoryLimitsMissing:   conf.SeverityError,
		},
	}

	nsResults := &NamespacedResults{}
	ValidateControllers(c, resources, nsResults)
	assert.Equal(t, 1, len((*nsResults)["test"].DeploymentResults), "")
	podRes := (*nsResults)["test"].DeploymentResults[0].PodResult
	assert.Equal(t, uint(0), podRes.Summary.Totals.Successes, "")
	assert.Equal(t, uint(4), podRes.Summary.Totals.Errors, "")

	test.SetupAddLimitRanges(k8s, "test")
	resources, err = kube.CreateResourceProviderFromAPI(k8s, "test")
	nsResults = &NamespacedResults{}
	ValidateControllers(c, resources, nsResults)
	podRes = (*nsResults)["test"].DeploymentResults[0].PodResult
	assert.Equal(t, uint(4), podRes.Summary.Totals.Successes, "")
	assert.Equal(t, uint(0), podRes.Summary.Totals.Errors, "")
}
