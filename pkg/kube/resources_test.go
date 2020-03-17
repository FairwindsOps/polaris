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

	assert.Equal(t, 1, len(resources.Deployments), "Should have a deployment")
	assert.Equal(t, "ubuntu", resources.Deployments[0].Spec.Template.Spec.Containers[0].Name)

	assert.Equal(t, 1, len(resources.StatefulSets), "Should have a stateful set")
	assert.Equal(t, "nginx", resources.StatefulSets[0].Spec.Template.Spec.Containers[0].Name)

	assert.Equal(t, 1, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "two", resources.Namespaces[0].ObjectMeta.Name)

	assert.Equal(t, 2, len(resources.Pods), "Should have two pods")
	assert.Equal(t, "", resources.Pods[0].ObjectMeta.Namespace, "Should have one pod in default namespace")
	assert.Equal(t, "two", resources.Pods[1].ObjectMeta.Namespace, "Should have one pod in namespace 'two'")
}

func TestGetMultipleResourceFromSingleFile(t *testing.T) {
	resources, err := CreateResourceProviderFromPath("./test_files/test_2/multi.yaml")

	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Path", resources.SourceType, "Should have type Path")
	assert.Equal(t, "./test_files/test_2/multi.yaml", resources.SourceName, "Should have filename as name")
	assert.Equal(t, "unknown", resources.ServerVersion, "Server version should be unknown")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Deployments), "Should have a deployment")
	assert.Equal(t, "dashboard", resources.Deployments[0].Spec.Template.Spec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}

func TestGetMultipleResourceFromBadFile(t *testing.T) {
	_, err := CreateResourceProviderFromPath("./test_files/test_3")
	assert.NotEqual(t, nil, err, "CreateResource From Path should fail with bad yaml")
}

func TestGetResourceFromAPI(t *testing.T) {
	k8s := test.SetupTestAPI()
	k8s = test.SetupAddControllers(k8s, "test")
	resources, err := CreateResourceProviderFromAPI(k8s, "test", nil)
	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "Cluster", resources.SourceType, "Should have type Path")
	assert.Equal(t, "test", resources.SourceName, "Should have source name")
	assert.IsType(t, time.Now(), resources.CreationTime, "Creation time should be set")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")
	assert.Equal(t, 0, len(resources.Pods), "Should have a pod")

}
