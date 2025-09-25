package cmd

import (
	"context"

	"github.com/gardener/etcd-druid/druidctl/cli/types"
	core "github.com/gardener/etcd-druid/druidctl/internal"
	"github.com/spf13/cobra"
)

// Create add-component-protection subcommand
func newAddProtectionCommand(options *types.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "add-component-protection <etcd-resource-name>",
		Short: "Adds resource protection to all managed components for a given etcd cluster",
		Long: `Adds resource protection to all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx, err := types.NewCommandContext(cmd, args, options)
			if err != nil {
				return err
			}
			if err := cmdCtx.Validate(); err != nil {
				return err
			}

			resourceProtectionCtx := types.NewResourceProtectionCommandContext(cmdCtx)
			if err := resourceProtectionCtx.Validate(); err != nil {
				return err
			}

			if resourceProtectionCtx.AllNamespaces {
				resourceProtectionCtx.Output.Info("Adding component protection to all namespaces")
			} else {
				resourceProtectionCtx.Output.Info("Adding component protection to Etcd", resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace)
			}

			if err := core.AddDisableProtectionAnnotation(context.TODO(), resourceProtectionCtx); err != nil {
				resourceProtectionCtx.Output.Error("Add component protection failed", err)
				return err
			}

			resourceProtectionCtx.Output.Success("Component protection added successfully")
			return nil
		},
	}
}

// Create remove-component-protection subcommand
func newRemoveProtectionCommand(options *types.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-component-protection <etcd-resource-name>",
		Short: "Removes resource protection for all managed components for a given etcd cluster",
		Long: `Removes resource protection for all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx, err := types.NewCommandContext(cmd, args, options)
			if err != nil {
				return err
			}
			if err := cmdCtx.Validate(); err != nil {
				return err
			}

			resourceProtectionCtx := types.NewResourceProtectionCommandContext(cmdCtx)
			if err := resourceProtectionCtx.Validate(); err != nil {
				return err
			}

			if resourceProtectionCtx.AllNamespaces {
				resourceProtectionCtx.Output.Info("Removing component protection from Etcds across all namespaces")
			} else {
				resourceProtectionCtx.Output.Info("Removing component protection from Etcd", resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace)
			}

			if err := core.RemoveDisableProtectionAnnotation(context.TODO(), resourceProtectionCtx); err != nil {
				resourceProtectionCtx.Output.Error("Remove component protection failed", err)
				return err
			}

			resourceProtectionCtx.Output.Success("Component protection removed successfully")
			return nil
		},
	}
}
