package cmd

import (
	"github.com/gardener/etcd-druid/druidctl/cli/types"
	"github.com/gardener/etcd-druid/druidctl/pkg/banner"
	"github.com/spf13/cobra"
)

// Global options instance
var options *types.Options

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
	options = types.NewOptions()
	options.AddFlags(rootCmd)

	originalPreRun := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		banner.ShowBanner(options.DisableBanner)

		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
	}

	// Add subcommands
	rootCmd.AddCommand(NewReconcileCommand(options))
	rootCmd.AddCommand(newAddProtectionCommand(options))
	rootCmd.AddCommand(newRemoveProtectionCommand(options))
	return rootCmd.Execute()
}
