// Copyright 2022 FairwindsOps Inc
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

package webhook

import (
	"context"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/sirupsen/logrus"
	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
)

// Mutator mutate k8s resources.
type Mutator struct {
	Client  client.Client
	Config  config.Configuration
	decoder *admission.Decoder
}

// NewMutateWebhook creates a mutating admission webhook for the apiType.
func NewMutateWebhook(mgr manager.Manager, c config.Configuration) {
	path := "/mutate"

	mutator := Mutator{
		Client:  mgr.GetClient(),
		decoder: admission.NewDecoder(runtime.NewScheme()),
		Config:  c,
	}
	mgr.GetWebhookServer().Register(path, &webhook.Admission{Handler: &mutator})
}

func (m *Mutator) mutate(req admission.Request) ([]jsonpatch.Operation, error) {
	results, kubeResources, err := GetValidatedResults(req.AdmissionRequest.Kind.Kind, m.decoder, req, m.Config)
	if err != nil {
		logrus.Errorf("Error while validating resource: %v", err)
		return nil, err
	}
	if results == nil {
		logrus.Infof("Not mutating owned pod")
		return nil, nil
	}
	patches := mutation.GetMutationsFromResult(results)
	originalYaml, err := yaml.JSONToYAML(kubeResources.OriginalObjectJSON)
	if err != nil {
		logrus.Errorf("Failed to convert JSON to YAML: %v", err)
		return nil, err
	}
	mutatedYamlStr, err := mutation.ApplyAllMutations(string(originalYaml), patches)
	if err != nil {
		logrus.Errorf("Failed to apply mutations: %v", err)
		return nil, err
	}

	mutatedJson, err := yaml.YAMLToJSON([]byte(mutatedYamlStr))
	if err != nil {
		logrus.Errorf("Failed to convert YAML to JSON: %v", err)
		return nil, err
	}

	ops, err := jsonpatch.CreatePatch(kubeResources.OriginalObjectJSON, mutatedJson)
	if err != nil {
		logrus.Errorf("Failed to create patch from mutation: %v", err)
		return nil, err
	}
	return ops, nil
}

// Handle for Validator to run validation checks.
func (m *Mutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logrus.Info("Starting mutation request")
	patches, err := m.mutate(req)
	if err != nil {
		logrus.Errorf("Error while getting mutations: %v", err)
		return admission.Errored(403, err)
	}
	if patches == nil {
		logrus.Infof("No patches generated")
		return admission.Allowed("Allowed")
	}
	logrus.Infof("Generated %d patches", len(patches))
	return admission.Patched("", patches...)
}
