package test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
)

var testCases = []testCase{}

type testCase struct {
	check   string
	input   []byte
	failure bool
}

func init() {
	_, baseDir, _, _ := runtime.Caller(0)
	baseDir = filepath.Dir(baseDir) + "/checks"
	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {
		check := dir.Name()
		checkDir := baseDir + "/" + check
		cases, err := ioutil.ReadDir(checkDir)
		if err != nil {
			panic(err)
		}
		for _, tc := range cases {
			body, err := ioutil.ReadFile(checkDir + "/" + tc.Name())
			if err != nil {
				panic(err)
			}
			testCases = append(testCases, testCase{
				check:   check,
				input:   body,
				failure: strings.Contains(tc.Name(), "failure"),
			})
		}
	}
}

func TestChecks(t *testing.T) {
	for _, tc := range testCases {
		workload, err := kube.GetWorkloadFromBytes(tc.input)
		assert.NoError(t, err)
		c, err := config.Parse([]byte("checks:\n  " + tc.check + ": danger"))
		assert.NoError(t, err)
		var result validator.Result
		result, err = validator.ValidateController(&c, *workload)
		assert.NoError(t, err)
		summary := result.GetSummary()
		if tc.failure {
			message := "Check " + tc.check + " passed unexpectedly"
			assert.Equal(t, uint(0), summary.Successes, message)
			assert.Equal(t, uint(1), summary.Dangers, message)
		} else {
			message := "Check " + tc.check + " failed unexpectedly"
			assert.Equal(t, uint(1), summary.Successes, message)
			assert.Equal(t, uint(0), summary.Dangers, message)
		}
	}
}
