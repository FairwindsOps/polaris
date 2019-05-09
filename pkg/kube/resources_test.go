package kube

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetResourcesFromPath(t *testing.T) {
	resources, err := CreateResourceProviderFromPath("./test_files/test_1")

	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "unknown", resources.ServerVersion, "Server version should be unknown")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Deployments), "Should have a deployment")
	assert.Equal(t, "ubuntu", resources.Deployments[0].Spec.Template.Spec.Containers[0].Name)

	assert.Equal(t, 1, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "two", resources.Namespaces[0].ObjectMeta.Name)

	assert.Equal(t, 2, len(resources.Pods), "Should have two pods")
	assert.Equal(t, "", resources.Pods[0].ObjectMeta.Namespace, "Should have one pod in default namespace")
	assert.Equal(t, "two", resources.Pods[1].ObjectMeta.Namespace, "Should have one pod in namespace 'two'")
}

func TestGetMultipleResourceFromSingleFile(t *testing.T) {
	resources, err := CreateResourceProviderFromPath("./test_files/test_2/multi.yaml")

	assert.Equal(t, nil, err, "Error should be nil")

	assert.Equal(t, "unknown", resources.ServerVersion, "Server version should be unknown")

	assert.Equal(t, 0, len(resources.Nodes), "Should not have any nodes")

	assert.Equal(t, 1, len(resources.Deployments), "Should have a deployment")
	assert.Equal(t, "dashboard", resources.Deployments[0].Spec.Template.Spec.Containers[0].Name)

	assert.Equal(t, 2, len(resources.Namespaces), "Should have a namespace")
	assert.Equal(t, "polaris", resources.Namespaces[0].ObjectMeta.Name)
	assert.Equal(t, "polaris-2", resources.Namespaces[1].ObjectMeta.Name)
}
