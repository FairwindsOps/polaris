package validator

import (
	"context"
	"testing"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
)

func TestGetTemplateData(t *testing.T) {
	k8s, dynamicClient := test.SetupTestAPI(test.GetMockControllers("test")...)
	resources, err := kube.CreateResourceProviderFromAPI(context.Background(), k8s, "test", &dynamicClient)
	assert.Equal(t, err, nil, "error should be nil")
	assert.Equal(t, 5, len(resources.Controllers))

	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityDanger,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
	}

	sum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(3),
		Dangers:   uint(3),
	}
	score := uint(0)

	var actualAudit AuditData
	actualAudit, err = RunAudit(c, resources, false)
	assert.Equal(t, err, nil, "error should be nil")
	assert.Equal(t, score, actualAudit.Score, "")
	assert.EqualValues(t, sum, actualAudit.GetSummary())
	assert.Equal(t, actualAudit.SourceType, "Cluster", "should be from a cluster")
	assert.Equal(t, actualAudit.SourceName, "test", "should be from a cluster")

	expectedResults := []struct {
		kind    string
		results int
	}{
		{kind: "StatefulSet", results: 2},
		{kind: "DaemonSet", results: 2},
		{kind: "Deployment", results: 2},
		{kind: "Job", results: 0},
		{kind: "CronJob", results: 0},
	}

	assert.Equal(t, len(expectedResults), len(actualAudit.Results))
	for _, result := range actualAudit.Results {
		found := false
		for _, expected := range expectedResults {
			if expected.kind != result.Kind {
				continue
			}
			found = true
			assert.Equal(t, 1, len(result.PodResult.ContainerResults))
			assert.Equal(t, expected.results, len(result.PodResult.ContainerResults[0].Results))
		}
		assert.Equal(t, found, true)
	}
}
