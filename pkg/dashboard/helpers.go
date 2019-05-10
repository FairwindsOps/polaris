// Copyright 2019 ReactiveOps
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
	"github.com/reactiveops/polaris/pkg/validator"
	"strings"
)

func getWarningWidth(counts validator.CountSummary, fullWidth int) uint {
	return uint(float64(counts.Successes+counts.Warnings) / float64(counts.Successes+counts.Warnings+counts.Errors) * float64(fullWidth))
}

func getSuccessWidth(counts validator.CountSummary, fullWidth int) uint {
	return uint(float64(counts.Successes) / float64(counts.Successes+counts.Warnings+counts.Errors) * float64(fullWidth))
}

func getCategoryLink(category string) string {
	return strings.Replace(strings.ToLower(category), " ", "-", -1)
}

func getGrade(rs validator.ResultSummary) string {
	score := getScore(rs)
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

func getScore(rs validator.ResultSummary) uint {
	total := (rs.Totals.Successes * 2) + rs.Totals.Warnings + (rs.Totals.Errors * 2)
	return uint((float64(rs.Totals.Successes*2) / float64(total)) * 100)
}

func getWeatherIcon(rs validator.ResultSummary) string {
	score := getScore(rs)
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

func getWeatherText(rs validator.ResultSummary) string {
	score := getScore(rs)
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
	switch rm.Type {
	case "success":
		return "fas fa-check"
	case "warning":
		return "fas fa-exclamation"
	default:
		return "fas fa-times"
	}
}
