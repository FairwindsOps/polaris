package dashboard

import (
	"testing"

	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestGetGrade(t *testing.T) {
	input := validator.CountSummary{
		Successes: 10,
		Warnings:  3,
		Dangers:   1,
	}
	expectedOutput := "B-"
	actual := getGrade(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "A+", actual)
}

func TestGetWeatherIcon(t *testing.T) {
	input := validator.CountSummary{
		Successes: 10,
		Warnings:  3,
		Dangers:   1,
	}

	expectedOutput := "fa-cloud-sun"
	actual := getWeatherIcon(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "fa-cloud-showers-heavy", actual)
}
