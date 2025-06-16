package cmd

import (
	"github.com/gardener/etcd-druid/userInterface/pkg"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	configFlags   *genericclioptions.ConfigFlags
	verbose       bool
	allNamespaces bool
)

var rootCmd = &cobra.Command{
	Use:   "druid [command] [resource] [flags]",
	Short: "CLI for etcd-druid operator",
	Long:  `This is a command line interface for Druid. It allows you to interact with Druid using various commands and flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			cmd.Println("Verbose mode enabled")
		}
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	configFlags = pkg.GetConfigFlags()
	configFlags.AddFlags(rootCmd.PersistentFlags())
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "perform operation on all namespaces")
}
