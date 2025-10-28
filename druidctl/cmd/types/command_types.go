package cmd

import (
	"fmt"

	"github.com/gardener/etcd-druid/druidctl/client"
	"github.com/gardener/etcd-druid/druidctl/pkg/log"
	"github.com/gardener/etcd-druid/druidctl/pkg/printer"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// CommandContext holds common state and functionality for all commands
type CommandContext struct {
	ResourceName  string
	Namespace     string
	AllNamespaces bool
	Verbose       bool
	Logger        log.Logger
	Formatter     printer.Formatter
	Factory       client.Factory
	IOStreams     genericiooptions.IOStreams
	Clients       *ClientBundle // Lazy-loaded clients
}

// ClientBundle provides lazy-loaded clients to improve performance
type ClientBundle struct {
	factory    client.Factory
	etcdClient client.EtcdClientInterface
	genClient  client.GenericClientInterface
}

// NewClientBundle creates a new ClientBundle with the given factory
func NewClientBundle(factory client.Factory) *ClientBundle {
	return &ClientBundle{factory: factory}
}

// EtcdClient returns the etcd client, creating it if necessary
func (c *ClientBundle) EtcdClient() (client.EtcdClientInterface, error) {
	if c.etcdClient == nil {
		var err error
		c.etcdClient, err = c.factory.CreateEtcdClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create etcd client: %w", err)
		}
	}
	return c.etcdClient, nil
}

// GenericClient returns the generic client, creating it if necessary
func (c *ClientBundle) GenericClient() (client.GenericClientInterface, error) {
	if c.genClient == nil {
		var err error
		c.genClient, err = c.factory.CreateGenericClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create generic client: %w", err)
		}
	}
	return c.genClient, nil
}

func NewCommandContext(cmd *cobra.Command, args []string, options *GlobalOptions) (*CommandContext, error) {
	// Get common flags from options
	allNs := options.AllNamespaces
	verbose := options.Verbose

	outputLogger := log.NewLogger(options.LogType)
	outputLogger.SetVerbose(verbose)
	outputLogger.SetOutput(options.IOStreams.Out)

	var err error
	formatter, err := printer.NewFormatter(options.OutputFormat)
	if err != nil {
		outputLogger.Error("Failed to create formatter: ", err)
		return nil, err
	}

	resourceName := ""
	namespace := ""

	if len(args) > 0 {
		resourceName = args[0]
	}

	// Handle namespace resolution - with fallback for testing
	if options.ConfigFlags != nil {
		if namespace, _, err = options.ConfigFlags.ToRawKubeConfigLoader().Namespace(); err != nil {
			outputLogger.Error("Failed to get namespace: ", err)
		}
	}
	if namespace == "" {
		namespace = "default" // Fallback for testing
	}

	return &CommandContext{
		ResourceName:  resourceName,
		Namespace:     namespace,
		AllNamespaces: allNs,
		Verbose:       verbose,
		Logger:        outputLogger,
		Formatter:     formatter,
		Factory:       options.ClientFactory,
		IOStreams:     options.IOStreams,
		Clients:       NewClientBundle(options.ClientFactory),
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
