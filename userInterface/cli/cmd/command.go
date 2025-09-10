package cmd

import (
	"fmt"

	"github.com/gardener/etcd-druid/userInterface/core"
	"github.com/gardener/etcd-druid/userInterface/pkg/output"
	"github.com/spf13/cobra"
)

// CommandContext holds common state and functionality for all commands
type CommandContext struct {
	EtcdClient    core.EtcdClientI
	ResourceName  string
	Namespace     string
	AllNamespaces bool
	Verbose       bool
}

func NewCommandContext(cmd *cobra.Command, args []string, options *Options) (*CommandContext, error) {
	// Get common flags from options
	allNs := options.AllNamespaces
	verbose := options.Verbose

	// Set output verbosity
	output.SetVerbose(verbose)

	// Handle resource name and namespace
	resourceName := ""
	namespace := ""
	var err error

	if len(args) > 0 {
		resourceName = args[0]
	}
	if namespace, _, err = options.ConfigFlags.ToRawKubeConfigLoader().Namespace(); err != nil {
		output.Error(fmt.Sprintf("Failed to get namespace: %v", err))
	}

	// Create etcd client
	clientFactory := core.NewClientFactory(options.ConfigFlags)
	etcdClient, err := clientFactory.CreateTypedEtcdClient()
	if err != nil {
		output.Error(fmt.Sprintf("Unable to create etcd client: %v", err))
		return nil, err
	}

	return &CommandContext{
		EtcdClient:    etcdClient,
		ResourceName:  resourceName,
		Namespace:     namespace,
		AllNamespaces: allNs,
		Verbose:       verbose,
	}, nil
}

func (c *CommandContext) Validate() error {
	if c.AllNamespaces {
		if c.Namespace != "default" {
			return fmt.Errorf("cannot specify --namespace/-n with --all-namespaces/-A")
		}
		if c.ResourceName != "" {
			return fmt.Errorf("cannot specify a resource name with --all-namespaces/-A")
		}
	} else {
		if c.ResourceName == "" {
			return fmt.Errorf("etcd resource name is required when not using --all-namespaces")
		}
	}
	return nil
}
