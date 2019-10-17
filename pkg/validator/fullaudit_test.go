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
		HealthChecks: conf.HealthChecks{
			ReadinessProbeMissing: conf.SeverityError,
			LivenessProbeMissing:  conf.SeverityWarning,
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

	// TODO: split out the logic for calculating summaries into another set of tests
	sum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(0),
			Warnings:  uint(4),
			Errors:    uint(4),
		},
		ByCategory: CategorySummary{},
	}
	sum.ByCategory["Health Checks"] = &CountSummary{
		Successes: uint(0),
		Warnings:  uint(4),
		Errors:    uint(4),
	}

	actualAudit, err := RunAudit(c, resources)
	assert.Equal(t, err, nil, "error should be nil")

	assert.EqualValues(t, sum, actualAudit.ClusterSummary.Results)
	assert.Equal(t, actualAudit.SourceType, "Cluster", "should be from a cluster")
	assert.Equal(t, actualAudit.SourceName, "test", "should be from a cluster")

	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].DeploymentResults), "should be equal")
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].DeploymentResults), "should be equal")
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].DeploymentResults[0].PodResult.ContainerResults), "should be equal")
	assert.Equal(t, 2, len(actualAudit.NamespacedResults["test"].DeploymentResults[0].PodResult.ContainerResults[0].Messages), "should be equal")

	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].StatefulSetResults), "should be equal")
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].StatefulSetResults), "should be equal")
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].StatefulSetResults[0].PodResult.ContainerResults), "should be equal")
	assert.Equal(t, 2, len(actualAudit.NamespacedResults["test"].StatefulSetResults[0].PodResult.ContainerResults[0].Messages), "should be equal")
}
