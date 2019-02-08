package dashboard

import (
	"testing"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/validator"
	"github.com/reactiveops/fairwinds/test"
	"github.com/stretchr/testify/assert"
)

func TestGetTemplateData(t *testing.T) {
	k8s := test.SetupTestAPI()
	k8s = test.SetupAddDeploys(k8s, "test")

	c := conf.Configuration{
		HealthChecks: conf.Probes{
			Readiness: conf.ResourceRequire{
				Require: true,
			},
		},
	}

	sum := validator.ResultSummary{
		Successes: uint(4),
		Warnings:  uint(0),
		Failures:  uint(1),
	}

	res := validator.ResourceResult{}

	n := validator.NamespacedResult{
		Summary: &sum,
		Results: []validator.ResourceResult{res},
	}

	expectedTmplData := TemplateData{
		ClusterSummary: &sum,
		NamespacedResults: validator.NamespacedResults{
			"test": &n,
		},
	}
	actualTmplData, _ := getTemplateData(c, k8s)

	assert.Equal(t, actualTmplData.ClusterSummary.Failures, expectedTmplData.ClusterSummary.Failures, "should be equal")
	assert.Equal(t, actualTmplData.ClusterSummary.Successes, expectedTmplData.ClusterSummary.Successes, "should be equal")
	assert.Equal(t, len(actualTmplData.NamespacedResults["test"].Results), 1, "should be equal")
	assert.Equal(t, len(actualTmplData.NamespacedResults["test"].Results[0].ContainerResults[0].Messages), 5, "should be equal")
}
