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
	k8s, dynamicClient := test.SetupTestAPI()
	k8s = test.SetupAddControllers(context.Background(), k8s, "test")
	k8s = test.SetupAddExtraControllerVersions(context.Background(), k8s, "test-extra")
	// TODO figure out how to mock out dynamic client.
	// and add in pods for all controllers to fill out tests.
	resources, err := kube.CreateResourceProviderFromAPI(context.Background(), k8s, "test", &dynamicClient)
	assert.Equal(t, err, nil, "error should be nil")

	c := conf.Configuration{
		Checks: map[string]conf.Severity{
			"readinessProbeMissing": conf.SeverityDanger,
			"livenessProbeMissing":  conf.SeverityWarning,
		},
	}

	sum := CountSummary{
		Successes: uint(0),
		Warnings:  uint(1),
		Dangers:   uint(1),
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
		{kind: "Pod", results: 2},
	}

	assert.Equal(t, len(expected), len(actualAudit.Results))
	for idx, result := range actualAudit.Results {
		assert.Equal(t, expected[idx].kind, result.Kind)
		assert.Equal(t, 1, len(result.PodResult.ContainerResults))
		assert.Equal(t, expected[idx].results, len(result.PodResult.ContainerResults[0].Results))
	}
}
