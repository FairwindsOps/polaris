package kube

import (
	"testing"
	"time"

	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, 8, len(resources.Controllers), "Should have eight controllers")
	namespaceCount := map[string]int{}
	for _, controller := range resources.Controllers {
		namespaceCount[controller.ObjectMeta.GetNamespace()]++
	}
	assert.Equal(t, 7, namespaceCount[""], "Should have seven controller in default namespace")
	assert.Equal(t, 1, namespaceCount["two"], "Should have one controller in namespace 'two'")
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

func TestGetResourceFromAPI(t *testing.T) {
	k8s, dynamicInterface := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	// TODO find a way to mock out the dynamic client
	// and create fake pods in order to find all of the controllers.
	resources, err := CreateResourceProviderFromAPI(k8s, "test", &dynamicInterface)
	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Cluster", resources.SourceType, "Should have type Path")
	assert.Equal(t, "test", resources.SourceName, "Should have source name")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")
	assert.Equal(t, 1, len(resources.Controllers), "Should have 1 controller")

	assert.Equal(t, "", resources.Controllers[0].ObjectMeta.GetName())
}
