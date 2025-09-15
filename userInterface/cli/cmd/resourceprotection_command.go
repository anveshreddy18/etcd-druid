package cmd

import (
	"context"

	"github.com/gardener/etcd-druid/userInterface/core"
	"github.com/spf13/cobra"
)

type ResourceProtectionCommandContext struct {
	*CommandContext
}

func (r *ResourceProtectionCommandContext) Validate() error {
	// add validation logic if any
	return nil
}

// Create add-component-protection subcommand
func newAddProtectionCommand(options *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "add-component-protection <etcd-resource-name>",
		Short: "Adds resource protection to all managed components for a given etcd cluster",
		Long: `Adds resource protection to all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create command context with all common functionality
			cmdCtx, err := NewCommandContext(cmd, args, options)
			if err != nil {
				return err
			}

			// Validate command context
			if err := cmdCtx.Validate(); err != nil {
				return err
			}

			// create resource protection command context
			resourceProtectionCtx := &ResourceProtectionCommandContext{
				CommandContext: cmdCtx,
			}

			// Validate command context
			if err := resourceProtectionCtx.Validate(); err != nil {
				return err
			}

			if resourceProtectionCtx.AllNamespaces {
				resourceProtectionCtx.Output.Info("Adding component protection to all namespaces")
			} else {
				resourceProtectionCtx.Output.Info("Adding component protection to Etcd", resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace)
			}

			service := core.NewEtcdProtectionService(resourceProtectionCtx.EtcdClient, resourceProtectionCtx.Verbose, resourceProtectionCtx.Output)
			if err := service.AddDisableProtectionAnnotation(context.TODO(), resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace, resourceProtectionCtx.AllNamespaces); err != nil {
				resourceProtectionCtx.Output.Error("Add component protection failed", err)
				return err
			}

			resourceProtectionCtx.Output.Success("Component protection added successfully")
			return nil
		},
	}
}

// Create remove-component-protection subcommand
func newRemoveProtectionCommand(options *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-component-protection <etcd-resource-name>",
		Short: "Removes resource protection for all managed components for a given etcd cluster",
		Long: `Removes resource protection for all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create command context with all common functionality
			cmdCtx, err := NewCommandContext(cmd, args, options)
			if err != nil {
				return err
			}
			// Validate command context
			if err := cmdCtx.Validate(); err != nil {
				return err
			}

			// create resource protection command context
			resourceProtectionCtx := &ResourceProtectionCommandContext{
				CommandContext: cmdCtx,
			}

			// Validate command context
			if err := resourceProtectionCtx.Validate(); err != nil {
				return err
			}

			if resourceProtectionCtx.AllNamespaces {
				resourceProtectionCtx.Output.Info("Removing component protection from Etcds across all namespaces")
			} else {
				resourceProtectionCtx.Output.Info("Removing component protection from Etcd", resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace)
			}

			service := core.NewEtcdProtectionService(resourceProtectionCtx.EtcdClient, resourceProtectionCtx.Verbose, resourceProtectionCtx.Output)
			if err := service.RemoveDisableProtectionAnnotation(context.TODO(), resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace, resourceProtectionCtx.AllNamespaces); err != nil {
				resourceProtectionCtx.Output.Error("Remove component protection failed", err)
				return err
			}

			resourceProtectionCtx.Output.Success("Component protection removed successfully")
			return nil
		},
	}
}
