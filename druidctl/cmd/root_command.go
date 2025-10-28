package cmd

import (
	listresources "github.com/gardener/etcd-druid/druidctl/cmd/list-resources"
	"github.com/gardener/etcd-druid/druidctl/cmd/reconcile"
	resourceprotection "github.com/gardener/etcd-druid/druidctl/cmd/resource-protection"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/gardener/etcd-druid/druidctl/pkg/banner"
	"github.com/spf13/cobra"
)

// Global options instance
var options *types.GlobalOptions

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
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		banner.ShowBanner(rootCmd, cmd, options.DisableBanner)
		options.Complete(cmd, args)
		if err := options.Validate(); err != nil {
			options.Logger.Error("Validation failed: ", err)
			return err
		}
		if originalPreRun != nil {
			originalPreRun(cmd, args)
		}
		return nil
	}

	// Add subcommands
	rootCmd.AddCommand(reconcile.NewReconcileCommand(options))
	rootCmd.AddCommand(resourceprotection.NewAddProtectionCommand(options))
	rootCmd.AddCommand(resourceprotection.NewRemoveProtectionCommand(options))
	rootCmd.AddCommand(reconcile.NewSuspendReconcileCommand(options))
	rootCmd.AddCommand(reconcile.NewResumeReconcileCommand(options))
	rootCmd.AddCommand(listresources.NewListResourcesCommand(options))
	return rootCmd.Execute()
}
