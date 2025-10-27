package cmd

import (
	"os"

	"github.com/gardener/etcd-druid/druidctl/client"
	"github.com/gardener/etcd-druid/druidctl/pkg"
	"github.com/gardener/etcd-druid/druidctl/pkg/log"
	"github.com/gardener/etcd-druid/druidctl/pkg/printer"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// Options holds all global options and configuration for the CLI
type Options struct {
	Verbose       bool
	AllNamespaces bool
	DisableBanner bool
	OutputFormat  printer.OutputFormat
	LogType       log.LogType

	ConfigFlags   *genericclioptions.ConfigFlags
	ClientFactory client.Factory

	genericiooptions.IOStreams
}

// NewOptions returns a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		OutputFormat:  printer.OutputTypeNone,
		LogType:       log.LogTypeCharm,
		ConfigFlags:   pkg.GetConfigFlags(),
		ClientFactory: client.NewClientFactory(pkg.GetConfigFlags()),
		IOStreams:     genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}
}

// AddFlags adds flags to the specified command
func (o *Options) AddFlags(cmd *cobra.Command) {
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
