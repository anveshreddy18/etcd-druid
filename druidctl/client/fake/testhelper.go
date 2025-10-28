package fake

import (
	"context"
	"strings"
	"testing"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/gardener/etcd-druid/druidctl/pkg/log"
	"github.com/gardener/etcd-druid/druidctl/pkg/printer"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// TestHelper provides utilities for creating test environments
type TestHelper struct {
	etcdObjects []runtime.Object
	k8sObjects  []runtime.Object
	streams     genericiooptions.IOStreams
}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	streams, _, _, _ := genericiooptions.NewTestIOStreams()
	return &TestHelper{
		etcdObjects: make([]runtime.Object, 0),
		k8sObjects:  make([]runtime.Object, 0),
		streams:     streams,
	}
}

// WithEtcdObjects adds etcd objects to the test environment
func (h *TestHelper) WithEtcdObjects(objects []runtime.Object) *TestHelper {
	h.etcdObjects = append(h.etcdObjects, objects...)
	return h
}

// WithK8sObjects adds k8s objects to the test environment
func (h *TestHelper) WithK8sObjects(objects []runtime.Object) *TestHelper {
	h.k8sObjects = append(h.k8sObjects, objects...)
	return h
}

// WithTestScenario adds objects from a test scenario builder
func (h *TestHelper) WithTestScenario(builder *TestDataBuilder) *TestHelper {
	etcdObjs, k8sObjs := builder.Build()
	h.etcdObjects = append(h.etcdObjects, etcdObjs...)
	h.k8sObjects = append(h.k8sObjects, k8sObjs...)
	return h
}

// CreateTestOptions creates Options configured for testing
func (h *TestHelper) CreateTestOptions() *types.GlobalOptions {
	testFactory := NewTestFactoryWithData(h.etcdObjects, h.k8sObjects)

	return &types.GlobalOptions{
		OutputFormat:  printer.OutputTypeNone,
		LogType:       log.LogTypeCharm,
		ConfigFlags:   nil, // Will be handled gracefully by NewCommandContext
		ClientFactory: testFactory,
		IOStreams:     h.streams,
	}
}

// CreateTestCommandContext creates a CommandContext for testing
func (h *TestHelper) CreateTestCommandContext(resourceName string, allNamespaces bool) *types.ClientBundle {
	testFactory := NewTestFactoryWithData(h.etcdObjects, h.k8sObjects)
	return types.NewClientBundle(testFactory)
}

// GetOutputBuffer returns the output buffer for assertions
func (h *TestHelper) GetOutputBuffer() genericiooptions.IOStreams {
	return h.streams
}

// TestAssertions provides common assertion helpers for command testing
type TestAssertions struct {
	t *testing.T
}

// NewTestAssertions creates assertion helpers for the given test
func NewTestAssertions(t *testing.T) *TestAssertions {
	return &TestAssertions{t: t}
}

// AssertEtcdCount verifies the expected number of etcd resources
func (a *TestAssertions) AssertEtcdCount(client client.EtcdClientInterface, namespace string, expected int) {
	etcds, err := client.ListEtcds(context.TODO(), namespace)
	if err != nil {
		a.t.Fatalf("Failed to list etcds: %v", err)
	}
	if len(etcds.Items) != expected {
		a.t.Errorf("Expected %d etcd resources, got %d", expected, len(etcds.Items))
	}
}

// AssertEtcdExists verifies that a specific etcd resource exists
func (a *TestAssertions) AssertEtcdExists(client client.EtcdClientInterface, namespace, name string) *druidv1alpha1.Etcd {
	etcd, err := client.GetEtcd(context.TODO(), namespace, name)
	if err != nil {
		a.t.Fatalf("Expected etcd %s/%s to exist, but got error: %v", namespace, name, err)
	}
	if etcd.Name != name {
		a.t.Errorf("Expected etcd name '%s', got '%s'", name, etcd.Name)
	}
	return etcd
}

// AssertEtcdNotFound verifies that a specific etcd resource does not exist
func (a *TestAssertions) AssertEtcdNotFound(client client.EtcdClientInterface, namespace, name string) {
	_, err := client.GetEtcd(context.TODO(), namespace, name)
	if err == nil {
		a.t.Errorf("Expected etcd %s/%s to not exist, but it was found", namespace, name)
	}
}

// AssertNoError verifies that no error occurred
func (a *TestAssertions) AssertNoError(err error, message string) {
	if err != nil {
		a.t.Fatalf("%s: %v", message, err)
	}
}

// AssertError verifies that an error occurred
func (a *TestAssertions) AssertError(err error, message string) {
	if err == nil {
		a.t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertContains verifies that a string contains a substring
func (a *TestAssertions) AssertContains(str, substr, message string) {
	if !strings.Contains(str, substr) {
		a.t.Errorf("%s: expected '%s' to contain '%s'", message, str, substr)
	}
}
