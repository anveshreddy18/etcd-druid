package cmd

import (
	"github.com/gardener/etcd-druid/druidctl/pkg"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Options holds all global options and configuration for the CLI
type Options struct {
	ConfigFlags   *genericclioptions.ConfigFlags
	Verbose       bool
	AllNamespaces bool
	DisableBanner bool
	OutputFormat  string
}

// NewOptions returns a new Options instance with default values
func NewOptions() *Options {
	return &Options{
		ConfigFlags: pkg.GetConfigFlags(),
	}
}

// AddFlags adds flags to the specified command
func (o *Options) AddFlags(cmd *cobra.Command) {
	o.ConfigFlags.AddFlags(cmd.PersistentFlags())
	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", false,
		"If present, list the requested object(s) across all namespaces")
	cmd.PersistentFlags().BoolVar(&o.DisableBanner, "no-banner", false, "Disable the CLI banner")
	cmd.PersistentFlags().StringVarP(&o.OutputFormat, "output", "o", "", "Output format. One of: json, yaml")
}
