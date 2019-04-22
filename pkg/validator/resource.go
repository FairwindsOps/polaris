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

package validator

import (
	"errors"
	"fmt"
	conf "github.com/reactiveops/fairwinds/pkg/config"
)

// ResourceValidation contains methods shared by PodValidation and ContainerValidation
type ResourceValidation struct {
	Summary   *ResultSummary
	Errors    []*ResultMessage
	Warnings  []*ResultMessage
	Successes []*ResultMessage
}

func (rv *ResourceValidation) messages() []*ResultMessage {
	messages := []*ResultMessage{}
	messages = append(messages, rv.Errors...)
	messages = append(messages, rv.Warnings...)
	messages = append(messages, rv.Successes...)
	return messages
}

func (rv *ResourceValidation) addFailure(message string, severity conf.Severity, category string) {
	if severity == conf.SeverityError {
		rv.addError(message, category)
	} else if severity == conf.SeverityWarning {
		rv.addWarning(message, category)
	} else {
		errMsg := fmt.Sprintf("Invalid severity: %s", severity)
		log.Error(errors.New(errMsg), errMsg)
	}
}

func (rv *ResourceValidation) addError(message string, category string) {
	rv.Summary.Errors++
	rv.Errors = append(rv.Errors, &ResultMessage{
		Message:  message,
		Type:     MessageTypeError,
		Category: category,
	})
}

func (rv *ResourceValidation) addWarning(message string, category string) {
	rv.Summary.Warnings++
	rv.Warnings = append(rv.Warnings, &ResultMessage{
		Message:  message,
		Type:     MessageTypeWarning,
		Category: category,
	})
}

func (rv *ResourceValidation) addSuccess(message string, category string) {
	rv.Summary.Successes++
	rv.Successes = append(rv.Successes, &ResultMessage{
		Message:  message,
		Type:     MessageTypeSuccess,
		Category: category,
	})
}
