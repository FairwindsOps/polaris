package dashboard

import (
	"testing"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestWarningWidth(t *testing.T) {
	input1 := validator.CountSummary{
		Successes: 0,
		Warnings:  0,
		Dangers:   0,
	}
	input2 := 6

	expectedOutput := uint(0x6)
	actual := getWarningWidth(input1, input2)

	assert.Equal(t, expectedOutput, actual)

	input1 = validator.CountSummary{
		Successes: 10,
		Warnings:  3,
		Dangers:   1,
	}
	input2 = 3

	expectedOutput = uint(0x2)
	actual = getWarningWidth(input1, input2)

	assert.Equal(t, expectedOutput, actual)
}
func TestSuccessWidth(t *testing.T) {
	input1 := validator.CountSummary{
		Successes: 0,
		Warnings:  0,
		Dangers:   0,
	}
	input2 := 6

	expectedOutput := uint(0x6)
	actual := getSuccessWidth(input1, input2)

	assert.Equal(t, expectedOutput, actual)

	input1 = validator.CountSummary{
		Successes: 8,
		Warnings:  6,
		Dangers:   4,
	}
	input2 = 7

	expectedOutput = uint(0x3)
	actual = getSuccessWidth(input1, input2)

	assert.Equal(t, expectedOutput, actual)
}

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

func TestGetResultClass(t *testing.T) {
	input := validator.ResultMessage{
		ID:       "",
		Message:  "",
		Details:  []string(nil),
		Success:  false,
		Severity: "",
		Category: "",
	}

	expectedOutput := " failure"
	actual := getResultClass(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, " success", actual)

	input = validator.ResultMessage{
		ID:       "",
		Message:  "",
		Details:  []string(nil),
		Success:  true,
		Severity: "",
		Category: "",
	}

	expectedOutput = " success"
	actual = getResultClass(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, " failure", actual)
}

func TestGetWeatherText(t *testing.T) {
	input := validator.CountSummary{
		Successes: 10,
		Warnings:  3,
		Dangers:   1,
	}

	expectedOutput := "Mostly smooth sailing"
	actual := getWeatherText(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "Storms ahead, be careful", actual)
}

func TestGetIcon(t *testing.T) {
	input := validator.ResultMessage{
		ID:       "",
		Message:  "",
		Details:  []string(nil),
		Success:  false,
		Severity: "",
		Category: "",
	}

	expectedOutput := "fas fa-times"
	actual := getIcon(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "fas fa-check", actual)

	input = validator.ResultMessage{
		ID:       "",
		Message:  "",
		Details:  []string(nil),
		Success:  true,
		Severity: "",
		Category: "",
	}

	expectedOutput = "fas fa-check"
	actual = getIcon(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "fas fa-times", actual)

	input = validator.ResultMessage{
		ID:       "",
		Message:  "",
		Details:  []string(nil),
		Success:  false,
		Severity: config.SeverityWarning,
		Category: "",
	}

	expectedOutput = "fas fa-exclamation"
	actual = getIcon(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "fas fa-times", actual)
}

func TestGetCategoryLink(t *testing.T) {
	input := "Efficiency"

	expectedOutput := "https://polaris.docs.fairwinds.com/checks/efficiency"
	actual := getCategoryLink(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "ttps://polaris.docs.fairwinds.com/checks/reliability", actual)
}

func TestGetCategoryInfo(t *testing.T) {
	input := "Security"

	expectedOutput :=
		`
			Kubernetes provides a great deal of configurability when it comes to the
			security of your workloads. A key principle here involves limiting the level
			of access any individual workload has. Polaris has validations for a number of
			best practices, mostly focused on ensuring that unnecessary access has not
			been granted to an application workload.
		`
	actual := getCategoryInfo(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "fas fa-times", actual)

	input = "Reliability"

	expectedOutput =
		`
			Kubernetes is built to reliabily run highly available applications.
			Polaris includes a number of checks to ensure that you are maximizing
			the reliability potential of Kubernetes.
		`
	actual = getCategoryInfo(input)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, "fas fa-times", actual)
}

func TestStringInSlice(t *testing.T) {
	input1 := "a"
	input2 := []string{"a", "b", "cde"}

	expectedOutput := true
	actual := stringInSlice(input1, input2)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, false, actual)

	input1 = "f"
	input2 = []string{"a", "b", "cde"}

	expectedOutput = false
	actual = stringInSlice(input1, input2)

	assert.Equal(t, expectedOutput, actual)
	assert.NotEqual(t, true, actual)
}

