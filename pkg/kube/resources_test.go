package kube

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetResourcesFromPath(t *testing.T) {
	resources, err := CreateResourceProviderFromPath("./test_files/test_1")

	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Path", resources.SourceType, "Should have type Path")
	assert.Equal(t, "./test_files/test_1", resources.SourceName, "Should have filename as name")
	assert.Equal(t, "unknown", resources.ServerVersion, "Server version should be unknown")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "two", resources.Namespaces[0].ObjectMeta.Name)

	assert.Equal(t, 9, len(resources.Controllers), "Should have eight controllers")
	namespaceCount := map[string]int{}
	for _, controller := range resources.Controllers {
		namespaceCount[controller.ObjectMeta.GetNamespace()]++
	}
	assert.Equal(t, 8, namespaceCount[""])
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

	assert.Equal(t, 1, len(resources.Controllers), "Should have one controller")
	assert.Equal(t, "dashboard", resources.Controllers[0].PodSpec.Containers[0].Name)

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
	resources := &ResourceProvider{
		ServerVersion: "unknown",
		SourceType:    "Path",
		SourceName:    "-",
		Nodes:         []corev1.Node{},
		Namespaces:    []corev1.Namespace{},
		Controllers:   []GenericWorkload{},
	}
	err = addResourcesFromReader(reader, resources)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Controllers), "Should have one controller")
	assert.Equal(t, "dashboard", resources.Controllers[0].PodSpec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}

func TestGetResourceFromAPI(t *testing.T) {
	k8s, dynamicInterface := test.SetupTestAPI(test.GetMockControllers("test")...)
	resources, err := CreateResourceProviderFromAPI(context.Background(), k8s, "test", &dynamicInterface)
	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Cluster", resources.SourceType, "Should have type Path")
	assert.Equal(t, "test", resources.SourceName, "Should have source name")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")
	assert.Equal(t, 0, len(resources.Ingresses), "Should not have any ingresses")
	assert.Equal(t, 5, len(resources.Controllers), "Should have 5 controllers")

	expectedNames := map[string]bool{
		"deploy":      false,
		"job":         false,
		"cronjob":     false,
		"statefulset": false,
		"daemonset":   false,
	}
	for _, ctrl := range resources.Controllers {
		expectedNames[ctrl.ObjectMeta.GetName()] = true
	}
	for name, val := range expectedNames {
		assert.Equal(t, true, val, name)
	}
}
