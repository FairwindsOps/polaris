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

package validator

import (
	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/sirupsen/logrus"
)

// ResourceValidation contains methods shared by PodValidation and ContainerValidation
type ResourceValidation struct {
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

func (rv *ResourceValidation) summary() *ResultSummary {
	counts := CountSummary{
		Errors:    uint(len(rv.Errors)),
		Warnings:  uint(len(rv.Warnings)),
		Successes: uint(len(rv.Successes)),
	}
	byCategory := CategorySummary{}
	for _, msg := range rv.messages() {
		if _, ok := byCategory[msg.Category]; !ok {
			byCategory[msg.Category] = &CountSummary{}
		}
		if msg.Type == MessageTypeError {
			byCategory[msg.Category].Errors++
		} else if msg.Type == MessageTypeWarning {
			byCategory[msg.Category].Warnings++
		} else if msg.Type == MessageTypeSuccess {
			byCategory[msg.Category].Successes++
		}
	}
	return &ResultSummary{
		Totals:     counts,
		ByCategory: byCategory,
	}
}

func (rv *ResourceValidation) addMessage(message ResultMessage) {
	if message.Type == MessageTypeError {
		rv.Errors = append(rv.Errors, &message)
	} else if message.Type == MessageTypeWarning {
		rv.Warnings = append(rv.Warnings, &message)
	} else if message.Type == MessageTypeSuccess {
		rv.Successes = append(rv.Successes, &message)
	} else {
		panic("Bad message type")
	}
}

func (rv *ResourceValidation) addFailure(message string, severity conf.Severity, category string, id string) {
	if severity == conf.SeverityError {
		rv.addError(message, category, id)
	} else if severity == conf.SeverityWarning {
		rv.addWarning(message, category, id)
	} else {
		logrus.Errorf("Invalid severity: %s", severity)
	}
}

func (rv *ResourceValidation) addError(message string, category string, id string) {
	rv.Errors = append(rv.Errors, &ResultMessage{
		ID:       id,
		Message:  message,
		Type:     MessageTypeError,
		Category: category,
	})
}

func (rv *ResourceValidation) addWarning(message string, category string, id string) {
	rv.Warnings = append(rv.Warnings, &ResultMessage{
		ID:       id,
		Message:  message,
		Type:     MessageTypeWarning,
		Category: category,
	})
}

func (rv *ResourceValidation) addSuccess(message string, category string, id string) {
	rv.Successes = append(rv.Successes, &ResultMessage{
		ID:       id,
		Message:  message,
		Type:     MessageTypeSuccess,
		Category: category,
	})
}
