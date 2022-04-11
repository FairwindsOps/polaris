// Copyright 2022 FairwindsOps, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
)

func TestGetResourcesFromPath(t *testing.T) {
	provider, err := CreateResourceProviderFromPath("./test_files/test_1")

	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Path", provider.SourceType, "Should have type Path")
	assert.Equal(t, "./test_files/test_1", provider.SourceName, "Should have filename as name")
	assert.Equal(t, "unknown", provider.ServerVersion, "Server version should be unknown")
	assert.IsType(t, time.Now(), provider.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(provider.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(provider.Namespaces), "Should have a namespace")
	assert.Equal(t, "two", provider.Namespaces[0].ObjectMeta.Name)

	namespaceCount := map[string]int{}
	for _, resources := range provider.Resources {
		for _, controller := range resources {
			namespaceCount[controller.ObjectMeta.GetNamespace()]++
		}
	}
	assert.Equal(t, 11, provider.Resources.GetLength())
	assert.Equal(t, 10, namespaceCount[""])
	assert.Equal(t, 1, namespaceCount["two"])
}

func TestGetMultipleResourceFromSingleFile(t *testing.T) {
	resources, err := CreateResourceProviderFromPath("./test_files/test_2/multi.yaml")

	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Path", resources.SourceType, "Should have type Path")
	assert.Equal(t, "./test_files/test_2/multi.yaml", resources.SourceName, "Should have filename as name")
	assert.Equal(t, "unknown", resources.ServerVersion, "Server version should be unknown")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Resources["extensions/Deployment"]), "Should have one controller")
	assert.Equal(t, "dashboard", resources.Resources["extensions/Deployment"][0].PodSpec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}

func TestGetMultipleResourceFromBadFile(t *testing.T) {
	_, err := CreateResourceProviderFromPath("./test_files/test_3")
	assert.Equal(t, nil, err, "CreateResource From Path should not fail with bad yaml")
}

func TestAddResourcesFromReader(t *testing.T) {
	contents, err := ioutil.ReadFile("./test_files/test_2/multi.yaml")
	assert.NoError(t, err)
	reader := bytes.NewBuffer(contents)
	resources := newResourceProvider("unknown", "Path", "-")
	err = resources.addResourcesFromReader(reader)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Resources["extensions/Deployment"]), "Should have one controller")
	assert.Equal(t, "dashboard", resources.Resources["extensions/Deployment"][0].PodSpec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}

func TestGetResourceFromAPI(t *testing.T) {
	k8s, dynamicInterface := test.SetupTestAPI(test.GetMockControllers("test")...)

	expectedNames := map[string]bool{
		"deploy":      false,
		"job":         false,
		"cronjob":     false,
		"statefulset": false,
		"daemonset":   false,
	}

	tests := []struct {
		name        string
		config      conf.Configuration
		want        *ResourceProvider
		wantErr     bool
		clusterName string
	}{
		{
			name:        "standard",
			config:      conf.Configuration{},
			clusterName: "test1",
			want: &ResourceProvider{
				SourceType:   "Cluster",
				SourceName:   "test1",
				CreationTime: time.Now(),
			},
		},
		{
			name: "namespaced",
			config: conf.Configuration{
				Namespace: "test",
			},
			clusterName: "test2",
			want: &ResourceProvider{
				SourceType:   "ClusterNamespace",
				SourceName:   "test2",
				CreationTime: time.Now(),
			},
		},
		{
			name: "namespace does not exist",
			config: conf.Configuration{
				Namespace: "test3",
			},
			clusterName: "test3",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := CreateResourceProviderFromAPI(context.Background(), k8s, tt.clusterName, &dynamicInterface, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.SourceType, resources.SourceType)
				assert.Equal(t, tt.want.SourceName, resources.SourceName)
				assert.IsType(t, tt.want.CreationTime, resources.CreationTime)
				assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")
				assert.Equal(t, 5, len(resources.Resources), "Should have 5 controllers")

				for _, controllers := range resources.Resources {
					for _, ctrl := range controllers {
						expectedNames[ctrl.ObjectMeta.GetName()] = true
					}
				}
				for name, val := range expectedNames {
					assert.Equal(t, true, val, name)
				}
			}
		})
	}
}
