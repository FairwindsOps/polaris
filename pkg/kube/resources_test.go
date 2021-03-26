package kube

import (
	"bytes"
	"context"
	"fmt"
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
	for kind, resources := range provider.Resources {
		fmt.Println("found", kind, len(resources))
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

	assert.Equal(t, 1, len(resources.Resources["Deployment"]), "Should have one controller")
	assert.Equal(t, "dashboard", resources.Resources["Deployment"][0].PodSpec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}

func TestGetMultipleResourceFromBadFile(t *testing.T) {
	_, err := CreateResourceProviderFromPath("./test_files/test_3")
	assert.NotEqual(t, nil, err, "CreateResource From Path should fail with bad yaml")
}

func TestAddResourcesFromReader(t *testing.T) {
	contents, err := ioutil.ReadFile("./test_files/test_2/multi.yaml")
	assert.NoError(t, err)
	reader := bytes.NewBuffer(contents)
	resources := newResourceProvider("unknown", "Path", "-")
	err = resources.addResourcesFromReader(reader)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Resources["Deployment"]), "Should have one controller")
	assert.Equal(t, "dashboard", resources.Resources["Deployment"][0].PodSpec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}

func TestGetResourceFromAPI(t *testing.T) {
	k8s, dynamicInterface := test.SetupTestAPI(test.GetMockControllers("test")...)
	resources, err := CreateResourceProviderFromAPI(context.Background(), k8s, "test", &dynamicInterface, conf.Configuration{})
	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Cluster", resources.SourceType, "Should have type Path")
	assert.Equal(t, "test", resources.SourceName, "Should have source name")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")
	assert.Equal(t, 5, len(resources.Resources), "Should have 5 controllers")

	expectedNames := map[string]bool{
		"deploy":      false,
		"job":         false,
		"cronjob":     false,
		"statefulset": false,
		"daemonset":   false,
	}
	for _, controllers := range resources.Resources {
		for _, ctrl := range controllers {
			expectedNames[ctrl.ObjectMeta.GetName()] = true
		}
	}
	for name, val := range expectedNames {
		assert.Equal(t, true, val, name)
	}
}
