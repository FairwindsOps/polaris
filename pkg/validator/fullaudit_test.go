package validator

import (
	"testing"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/test"
	"github.com/stretchr/testify/assert"
)

func TestGetTemplateData(t *testing.T) {
	k8s := test.SetupTestAPI()
	k8s = test.SetupAddDeploys(k8s, "test")

	c := conf.Configuration{
		HealthChecks: conf.HealthChecks{
			ReadinessProbeMissing: conf.SeverityError,
			LivenessProbeMissing:  conf.SeverityWarning,
		},
	}

	sum := ResultSummary{
		Totals: CountSummary{
			Successes: uint(4),
			Warnings:  uint(1),
			Errors:    uint(1),
		},
		ByCategory: CategorySummary{},
	}
	sum.ByCategory["Health Checks"] = &CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Errors:    uint(1),
	}
	sum.ByCategory["Resources"] = &CountSummary{
		Successes: uint(4),
		Warnings:  uint(0),
		Errors:    uint(0),
	}

	actualAudit, err := RunAudit(c, k8s)
	assert.Equal(t, err, nil, "error should be nil")

	assert.EqualValues(t, sum, actualAudit.ClusterSummary.Results)
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].PodResults), "should be equal")
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].PodResults), "should be equal")
	assert.Equal(t, 1, len(actualAudit.NamespacedResults["test"].PodResults[0].ContainerResults), "should be equal")
	assert.Equal(t, 6, len(actualAudit.NamespacedResults["test"].PodResults[0].ContainerResults[0].Messages), "should be equal")
}
