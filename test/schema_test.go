package test

import (
	"fmt"
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
	check    string
	filename string
	input    []byte
	failure  bool
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
				filename: tc.Name(),
				check:    check,
				input:    body,
				failure:  strings.Contains(tc.Name(), "failure"),
			})
		}
	}
}

func TestChecks(t *testing.T) {
	for _, tc := range testCases {
		res, err := kube.NewGenericResourceFromBytes(tc.input)
		if err != nil {
			fmt.Println("error parsing", string(tc.input))
			panic(err)
		}
		c, err := config.Parse([]byte("checks:\n  " + tc.check + ": danger"))
		if err != nil {
			panic(err)
		}
		result, err := validator.ApplyAllSchemaChecks(&c, res)
		if err != nil {
			panic(err)
		}
		summary := result.GetSummary()
		total := summary.Successes + summary.Dangers
		msg := fmt.Sprintf("Check %s ran %d times instead of 1", tc.check, total)
		if assert.Equal(t, uint(1), total, msg) {
			if tc.failure {
				message := "Check " + tc.check + " passed unexpectedly for " + tc.filename
				assert.Equal(t, uint(0), summary.Successes, message)
				assert.Equal(t, uint(1), summary.Dangers, message)
			} else {
				message := "Check " + tc.check + " failed unexpectedly for " + tc.filename
				assert.Equal(t, uint(1), summary.Successes, message)
				assert.Equal(t, uint(0), summary.Dangers, message)
			}
		}
	}
}
