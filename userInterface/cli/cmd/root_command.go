package cmd

import (
	"github.com/gardener/etcd-druid/userInterface/pkg"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Options holds all global options and configuration for the CLI
type Options struct {
	// Common CLI options
	ConfigFlags   *genericclioptions.ConfigFlags
	Verbose       bool
	AllNamespaces bool
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
}

// Global options instance
var options *Options

var rootCmd = &cobra.Command{
	Use:   "druid [command] [resource] [flags]",
	Short: "CLI for etcd-druid operator",
	Long:  `This is a command line interface for Druid. It allows you to interact with Druid using various commands and flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		if options.Verbose {
			cmd.Println("Verbose mode enabled")
		}
		cmd.Help()
	},
}

// Execute runs the root command
func Execute() error {
	options = NewOptions()
	options.AddFlags(rootCmd)

	// Add subcommands
	rootCmd.AddCommand(NewReconcileCommand(options))
	rootCmd.AddCommand(newAddProtectionCommand(options))
	rootCmd.AddCommand(newRemoveProtectionCommand(options))
	// We'll add other commands as we update them

	return rootCmd.Execute()
}
