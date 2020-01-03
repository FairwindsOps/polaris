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
	resources, err := kube.CreateResourceProviderFromAPI(k8s, "test")
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
		Warnings:  uint(4),
		Errors:    uint(4),
	}

	actualAudit, err := RunAudit(c, resources)
	assert.Equal(t, err, nil, "error should be nil")

	assert.EqualValues(t, sum, actualAudit.GetSummary())
	assert.Equal(t, actualAudit.SourceType, "Cluster", "should be from a cluster")
	assert.Equal(t, actualAudit.SourceName, "test", "should be from a cluster")

	assert.Equal(t, 6, len(actualAudit.Results))

	assert.Equal(t, "Deployments", actualAudit.Results[0].Type)
	assert.Equal(t, 1, len(actualAudit.Results[0].PodResult.ContainerResults))
	assert.Equal(t, 2, len(actualAudit.Results[0].PodResult.ContainerResults[0].Messages))

	assert.Equal(t, "StatefulSets", actualAudit.Results[1].Type)
	assert.Equal(t, 1, len(actualAudit.Results[1].PodResult.ContainerResults))
	assert.Equal(t, 2, len(actualAudit.Results[1].PodResult.ContainerResults[0].Messages))

	assert.Equal(t, "DaemonSets", actualAudit.Results[2].Type)
	assert.Equal(t, 1, len(actualAudit.Results[2].PodResult.ContainerResults))
	assert.Equal(t, 2, len(actualAudit.Results[2].PodResult.ContainerResults[0].Messages))

	assert.Equal(t, "Jobs", actualAudit.Results[3].Type)
	assert.Equal(t, 1, len(actualAudit.Results[3].PodResult.ContainerResults))
	assert.Equal(t, 0, len(actualAudit.Results[3].PodResult.ContainerResults[0].Messages))

	assert.Equal(t, "CronJobs", actualAudit.Results[4].Type)
	assert.Equal(t, 1, len(actualAudit.Results[4].PodResult.ContainerResults))
	assert.Equal(t, 0, len(actualAudit.Results[4].PodResult.ContainerResults[0].Messages))

	assert.Equal(t, "ReplicationController", actualAudit.Results[5].Type)
	assert.Equal(t, 1, len(actualAudit.Results[5].PodResult.ContainerResults))
	assert.Equal(t, 2, len(actualAudit.Results[5].PodResult.ContainerResults[0].Messages))
}
