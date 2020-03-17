package validator

import (
	"testing"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
)

func TestGetTemplateData(t *testing.T) {
	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	k8s = test.SetupAddExtraControllerVersions(k8s, "test-extra")
	resources, err := kube.CreateResourceProviderFromAPI(k8s, "test", nil)
	assert.Equal(t, err, nil, "error should be nil")

	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityError,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
		ControllersToScan: []conf.SupportedController{
			conf.Deployments,
			conf.StatefulSets,
			conf.DaemonSets,
			conf.Jobs,
			conf.CronJobs,
			conf.ReplicationControllers,
		},
	}

	sum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}

	actualAudit, err := RunAudit(c, resources)
	assert.Equal(t, err, nil, "error should be nil")

	assert.EqualValues(t, sum, actualAudit.GetSummary())
	assert.Equal(t, actualAudit.SourceType, "Cluster", "should be from a cluster")
	assert.Equal(t, actualAudit.SourceName, "test", "should be from a cluster")

	expected := []struct {
		kind    string
		results int
	}{
		{kind: "NakedPod", results: 2},
	}

	assert.Equal(t, len(expected), len(actualAudit.Results))
	for idx, result := range actualAudit.Results {
		assert.Equal(t, expected[idx].kind, result.Kind)
		assert.Equal(t, 1, len(result.PodResult.ContainerResults))
		assert.Equal(t, expected[idx].results, len(result.PodResult.ContainerResults[0].Results))
	}
}
