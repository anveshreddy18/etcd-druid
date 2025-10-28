package cmd

import (
	"fmt"
	"os"

	"github.com/gardener/etcd-druid/druidctl/client"
	"github.com/gardener/etcd-druid/druidctl/pkg"
	"github.com/gardener/etcd-druid/druidctl/pkg/log"
	"github.com/gardener/etcd-druid/druidctl/pkg/printer"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// GlobalOptions holds all global options and configuration for the CLI
type GlobalOptions struct {
	// Common options
	Verbose       bool
	AllNamespaces bool
	Namespace     string
	ResourceName  string
	DisableBanner bool
	OutputFormat  printer.OutputFormat
	LogType       log.LogType

	// IO options
	Logger    log.Logger
	Formatter printer.Formatter
	IOStreams genericiooptions.IOStreams

	// client options
	ConfigFlags   *genericclioptions.ConfigFlags
	ClientFactory client.Factory
	Clients       *ClientBundle
}

// NewOptions returns a new Options instance with default values
func NewOptions() *GlobalOptions {
	factory := client.NewClientFactory(pkg.GetConfigFlags())
	return &GlobalOptions{
		OutputFormat:  printer.OutputTypeNone,
		LogType:       log.LogTypeCharm,
		ConfigFlags:   pkg.GetConfigFlags(),
		ClientFactory: factory,
		Clients:       NewClientBundle(factory),
		IOStreams:     genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}
}

// AddFlags adds flags to the specified command
func (o *GlobalOptions) AddFlags(cmd *cobra.Command) {
	o.ConfigFlags.AddFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces")
	cmd.PersistentFlags().BoolVar(&o.DisableBanner, "no-banner", false, "Disable the CLI banner")

	var outputFormatStr string
	cmd.PersistentFlags().StringVarP(&outputFormatStr, "output", "o", "", "Output format. One of: json, yaml")
	if outputFormatStr != "" {
		o.OutputFormat = printer.OutputFormat(outputFormatStr)
	}
}

func (o *GlobalOptions) Complete(cmd *cobra.Command, args []string) error {
	// Initialize Logger with proper IOStreams integration
	o.Logger = log.NewLogger(o.LogType)
	o.Logger.SetVerbose(o.Verbose)

	// Initialize Formatter
	var err error
	o.Formatter, err = printer.NewFormatter(o.OutputFormat)
	if err != nil {
		o.Logger.Error("Failed to create formatter: ", err)
		return err
	}

	// Fill in ResourceName and Namespace
	if len(args) > 0 {
		o.ResourceName = args[0]
	}
	if o.Namespace, _, err = o.ConfigFlags.ToRawKubeConfigLoader().Namespace(); err != nil {
		o.Logger.Error("Failed to get namespace: ", err)
	}
	if o.Namespace == "" {
		o.Namespace = "default"
	}
	return nil
}

func (o *GlobalOptions) Validate() error {
	if o.AllNamespaces {
		if o.Namespace != "default" {
			return fmt.Errorf("cannot specify --namespace/-n with --all-namespaces/-A")
		}
		if o.ResourceName != "" {
			return fmt.Errorf("cannot specify a resource name with --all-namespaces/-A")
		}
	} else {
		if o.ResourceName == "" {
			return fmt.Errorf("etcd resource name is required when not using --all-namespaces")
		}
	}
	return nil
}
