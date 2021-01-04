// Copyright 2019 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dashboard

import (
	"fmt"
	"strings"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/validator"
)

func getWarningWidth(counts validator.CountSummary, fullWidth int) uint {
	denom := counts.Successes + counts.Warnings + counts.Dangers
	if denom == 0 {
		return uint(1 * float64(fullWidth))
	}
	res := float64(counts.Successes+counts.Warnings) / float64(denom) * float64(fullWidth)
	return uint(res)
}

func getSuccessWidth(counts validator.CountSummary, fullWidth int) uint {
	denom := counts.Successes + counts.Warnings + counts.Dangers
	if denom == 0 {
		return uint(1 * float64(fullWidth))
	}
	res := float64(counts.Successes) / float64(denom) * float64(fullWidth)
	return uint(res)
}

func getGrade(counts validator.CountSummary) string {
	score := counts.GetScore()
	if score >= 97 {
		return "A+"
	} else if score >= 93 {
		return "A"
	} else if score >= 90 {
		return "A-"
	} else if score >= 87 {
		return "B+"
	} else if score >= 83 {
		return "B"
	} else if score >= 80 {
		return "B-"
	} else if score >= 77 {
		return "C+"
	} else if score >= 73 {
		return "C"
	} else if score >= 70 {
		return "C-"
	} else if score >= 67 {
		return "D+"
	} else if score >= 63 {
		return "D"
	} else if score >= 60 {
		return "D-"
	} else {
		return "F"
	}
}

func getWeatherIcon(counts validator.CountSummary) string {
	score := counts.GetScore()
	if score >= 90 {
		return "fa-sun"
	} else if score >= 80 {
		return "fa-cloud-sun"
	} else if score >= 70 {
		return "fa-cloud"
	} else if score >= 60 {
		return "fa-cloud-rain"
	} else {
		return "fa-cloud-showers-heavy"
	}
}

func getResultClass(result validator.ResultMessage) string {
	cls := string(result.Severity)
	if result.Success {
		cls += " success"
	} else {
		cls += " failure"
	}
	return cls
}

func getWeatherText(counts validator.CountSummary) string {
	score := counts.GetScore()
	if score >= 90 {
		return "Smooth sailing"
	} else if score >= 80 {
		return "Mostly smooth sailing"
	} else if score >= 70 {
		return "Smooth sailing within sight"
	} else if score >= 60 {
		return "A little stormy"
	} else {
		return "Storms ahead, be careful"
	}
}

func getIcon(rm validator.ResultMessage) string {
	if rm.Success {
		return "fas fa-check"
	} else if rm.Severity == config.SeverityWarning {
		return "fas fa-exclamation"
	} else {
		return "fas fa-times"
	}
}

func getCategoryLink(category string) string {
	return "https://polaris.docs.fairwinds.com/checks/" + strings.ToLower(category)
}

func getCategoryInfo(category string) string {
	switch category {
	case "Reliability":
		return fmt.Sprintf(`
			Kubernetes is built to reliabily run highly available applications.
			Polaris includes a number of checks to ensure that you are maximizing
			the reliability potential of Kubernetes.
		`)
	case "Efficiency":
		return fmt.Sprintf(`
			Configuring resource requests and limits for workloads running in Kubernetes
			helps ensure that every container will have access to all the resources it
			needs. These are also a crucial part of cluster autoscaling logic, as new
			nodes are only spun up when there is insufficient capacity on existing
			infrastructure for new pod(s). By default, Polaris validates that resource
			requests and limits are set, it also includes optional functionality to ensure
			these requests and limits fall within specified ranges.
		`)
	case "Security":
		return fmt.Sprintf(`
			Kubernetes provides a great deal of configurability when it comes to the
			security of your workloads. A key principle here involves limiting the level
			of access any individual workload has. Polaris has validations for a number of
			best practices, mostly focused on ensuring that unnecessary access has not
			been granted to an application workload.
		`)
	default:
		return ""
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
