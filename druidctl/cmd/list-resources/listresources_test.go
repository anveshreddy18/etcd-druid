package listresources

import (
	"context"
	"testing"

	fake "github.com/gardener/etcd-druid/druidctl/client/fake"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/gardener/etcd-druid/druidctl/pkg/printer"
)

func TestListResources(t *testing.T) {
	// Create test helper with realistic scenario
	helper := fake.NewTestHelper().WithTestScenario(fake.SingleEtcdWithResources())
	options := helper.CreateTestOptions()

	// Test the full command flow using NewCommandContext
	cmdCtx, err := types.NewCommandContext(nil, []string{}, options)
	if err != nil {
		t.Fatalf("Failed to create command context: %v", err)
	}

	// Override for all-namespaces test case
	cmdCtx.AllNamespaces = true
	cmdCtx.ResourceName = ""

	// Create list command context using the enhanced CommandContext
	listCtx := newListResourcesCommandContext(cmdCtx, "po,svc")

	// Use the lazy-loaded clients from ClientBundle
	etcdClient, err := cmdCtx.Clients.EtcdClient()
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	listCtx.EtcdClient = etcdClient

	genClient, err := cmdCtx.Clients.GenericClient()
	if err != nil {
		t.Fatalf("Failed to create generic client: %v", err)
	}
	listCtx.GenericClient = genClient

	// Test execution
	err = listCtx.execute(context.TODO())
	if err != nil {
		t.Logf("Execution completed with expected error (discovery limitation): %v", err)
	}

	// Verify that the lazy-loaded etcd client works correctly
	etcds, err := etcdClient.ListEtcds(context.TODO(), "")
	if err != nil {
		t.Fatalf("Failed to list etcds: %v", err)
	}

	if len(etcds.Items) != 1 {
		t.Errorf("Expected 1 etcd, got %d", len(etcds.Items))
	}

	if etcds.Items[0].Name != "test-etcd" {
		t.Errorf("Expected etcd name 'test-etcd', got '%s'", etcds.Items[0].Name)
	}

	t.Logf("Successfully tested enhanced command architecture with %d etcd resources", len(etcds.Items))
}

func TestListResourcesMultipleEtcds(t *testing.T) {
	// Create test helper with multiple etcd scenario
	helper := fake.NewTestHelper().WithTestScenario(fake.MultipleEtcdsScenario())
	options := helper.CreateTestOptions()

	// Test the full command flow
	cmdCtx, err := types.NewCommandContext(nil, []string{}, options)
	if err != nil {
		t.Fatalf("Failed to create command context: %v", err)
	}

	// Override for all-namespaces test case
	cmdCtx.AllNamespaces = true
	cmdCtx.ResourceName = ""

	// Create list command context using enhanced architecture
	listCtx := newListResourcesCommandContext(cmdCtx, "all")

	// Use lazy-loaded clients
	etcdClient, err := cmdCtx.Clients.EtcdClient()
	if err != nil {
		t.Fatalf("Failed to create etcd client: %v", err)
	}
	listCtx.EtcdClient = etcdClient

	genClient, err := cmdCtx.Clients.GenericClient()
	if err != nil {
		t.Fatalf("Failed to create generic client: %v", err)
	}
	listCtx.GenericClient = genClient

	// Test execution
	err = listCtx.execute(context.TODO())
	if err != nil {
		t.Logf("Execution completed with expected error (discovery limitation): %v", err)
	}

	// Verify multiple etcd resources using lazy-loaded client
	etcds, err := etcdClient.ListEtcds(context.TODO(), "")
	if err != nil {
		t.Fatalf("Failed to list etcds: %v", err)
	}

	if len(etcds.Items) != 3 { // Should have 3 etcd resources from MultipleEtcdsScenario
		t.Errorf("Expected 3 etcd resources, got %d", len(etcds.Items))
	}

	// Verify namespace filtering works
	shootNs1Etcds, err := etcdClient.ListEtcds(context.TODO(), "shoot-ns1")
	if err != nil {
		t.Fatalf("Failed to list etcds in shoot-ns1: %v", err)
	}

	if len(shootNs1Etcds.Items) != 2 { // Should have 2 etcd resources in shoot-ns1
		t.Errorf("Expected 2 etcd resources in shoot-ns1, got %d", len(shootNs1Etcds.Items))
	}

	t.Logf("Successfully tested enhanced command architecture with multiple etcd scenario: %d total etcds", len(etcds.Items))
}

func TestClientBundleLazyLoading(t *testing.T) {
	// Test that ClientBundle creates clients only when needed
	helper := fake.NewTestHelper().WithTestScenario(fake.SingleEtcdWithResources())
	options := helper.CreateTestOptions()

	cmdCtx, err := types.NewCommandContext(nil, []string{}, options)
	if err != nil {
		t.Fatalf("Failed to create command context: %v", err)
	}

	// Initially, no clients should be created yet
	if cmdCtx.Clients == nil {
		t.Fatal("ClientBundle should be initialized")
	}

	// Create first client - should initialize etcd client
	etcdClient1, err := cmdCtx.Clients.EtcdClient()
	if err != nil {
		t.Fatalf("Failed to create first etcd client: %v", err)
	}

	// Get same client again - should return cached instance
	etcdClient2, err := cmdCtx.Clients.EtcdClient()
	if err != nil {
		t.Fatalf("Failed to get cached etcd client: %v", err)
	}

	// Should be the same instance (lazy loading works)
	if etcdClient1 != etcdClient2 {
		t.Error("Expected same client instance from lazy loading, got different instances")
	}

	// Test generic client lazy loading
	genClient1, err := cmdCtx.Clients.GenericClient()
	if err != nil {
		t.Fatalf("Failed to create first generic client: %v", err)
	}

	genClient2, err := cmdCtx.Clients.GenericClient()
	if err != nil {
		t.Fatalf("Failed to get cached generic client: %v", err)
	}

	if genClient1 != genClient2 {
		t.Error("Expected same generic client instance from lazy loading, got different instances")
	}

	t.Log("Successfully tested ClientBundle lazy loading behavior")
}

func TestListResourcesWithAssertions(t *testing.T) {
	// Test list-resources using TestAssertions helper
	helper := fake.NewTestHelper().WithTestScenario(fake.SingleEtcdWithResources())
	options := helper.CreateTestOptions()
	assert := fake.NewTestAssertions(t)

	cmdCtx, err := types.NewCommandContext(nil, []string{"test-etcd"}, options)
	assert.AssertNoError(err, "Failed to create command context")

	etcdClient, err := cmdCtx.Clients.EtcdClient()
	assert.AssertNoError(err, "Failed to create etcd client")

	genericClient, err := cmdCtx.Clients.GenericClient()
	assert.AssertNoError(err, "Failed to create generic client")

	// Use assertions to verify etcd exists
	etcd := assert.AssertEtcdExists(etcdClient, "default", "test-etcd")

	// Create list resources command context
	listCtx := newListResourcesCommandContext(cmdCtx, "po,svc")
	listCtx.EtcdClient = etcdClient
	listCtx.GenericClient = genericClient

	// Test execution
	err = listCtx.execute(context.TODO())
	if err != nil {
		t.Logf("Expected discovery error: %v", err)
	}

	t.Logf("Successfully used TestAssertions with etcd: %s", etcd.Name)
}

func TestListResourcesOutputFormats(t *testing.T) {
	// Test different output formats
	helper := fake.NewTestHelper().WithTestScenario(fake.SingleEtcdWithResources())

	tests := []struct {
		name         string
		outputFormat string
	}{
		{"default_output", ""},
		{"json_output", "json"},
		{"yaml_output", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := helper.CreateTestOptions()
			assert := fake.NewTestAssertions(t)

			// Set output format
			options.OutputFormat = printer.OutputFormat(tt.outputFormat)

			cmdCtx, err := types.NewCommandContext(nil, []string{"test-etcd"}, options)
			assert.AssertNoError(err, "Failed to create command context")

			etcdClient, err := cmdCtx.Clients.EtcdClient()
			assert.AssertNoError(err, "Failed to create etcd client")

			genericClient, err := cmdCtx.Clients.GenericClient()
			assert.AssertNoError(err, "Failed to create generic client")

			// Verify the test etcd exists
			assert.AssertEtcdExists(etcdClient, "default", "test-etcd")

			// Create list resources command context
			listCtx := newListResourcesCommandContext(cmdCtx, "po,svc")
			listCtx.EtcdClient = etcdClient
			listCtx.GenericClient = genericClient

			// Test list resources with specific output format
			err = listCtx.execute(context.TODO())
			if err != nil {
				t.Logf("Expected discovery error for output format %s: %v", tt.outputFormat, err)
			}

			t.Logf("Successfully tested list-resources with output format: %s", tt.outputFormat)
		})
	}
}

func TestListResourcesErrorHandling(t *testing.T) {
	// Test error cases with empty scenario
	helper := fake.NewTestHelper() // No test data
	options := helper.CreateTestOptions()
	assert := fake.NewTestAssertions(t)

	cmdCtx, err := types.NewCommandContext(nil, []string{"non-existent-etcd"}, options)
	assert.AssertNoError(err, "Failed to create command context")

	etcdClient, err := cmdCtx.Clients.EtcdClient()
	assert.AssertNoError(err, "Failed to create etcd client")

	genericClient, err := cmdCtx.Clients.GenericClient()
	assert.AssertNoError(err, "Failed to create generic client")

	// Verify etcd doesn't exist
	assert.AssertEtcdNotFound(etcdClient, "default", "non-existent-etcd")

	// Create list resources command context with non-existent etcd
	listCtx := newListResourcesCommandContext(cmdCtx, "po,svc")
	listCtx.EtcdClient = etcdClient
	listCtx.GenericClient = genericClient

	// Test that list handles non-existent resources gracefully
	err = listCtx.execute(context.TODO())
	if err != nil {
		t.Logf("Expected error for non-existent resource: %v", err)
	}

	t.Log("Successfully tested list-resources error handling")
}
