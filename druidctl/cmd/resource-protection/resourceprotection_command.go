package resourceprotection

import (
	"context"

	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/spf13/cobra"
)

// Create add-component-protection subcommand
func NewAddProtectionCommand(options *types.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "add-component-protection <etcd-resource-name>",
		Short: "Adds resource protection to all managed components for a given etcd cluster",
		Long: `Adds resource protection to all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceProtectionCtx, err := getResourceProtection(cmd, args, options)
			if err != nil {
				return err
			}

			if resourceProtectionCtx.AllNamespaces {
				resourceProtectionCtx.Logger.Info("Adding component protection to all namespaces")
			} else {
				resourceProtectionCtx.Logger.Info("Adding component protection to Etcd", resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace)
			}

			if err := resourceProtectionCtx.addDisableProtectionAnnotation(context.TODO()); err != nil {
				resourceProtectionCtx.Logger.Error("Add component protection failed", err)
				return err
			}

			resourceProtectionCtx.Logger.Success("Component protection added successfully")
			return nil
		},
	}
}

// Create remove-component-protection subcommand
func NewRemoveProtectionCommand(options *types.Options) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-component-protection <etcd-resource-name>",
		Short: "Removes resource protection for all managed components for a given etcd cluster",
		Long: `Removes resource protection for all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceProtectionCtx, err := getResourceProtection(cmd, args, options)
			if err != nil {
				return err
			}

			if resourceProtectionCtx.AllNamespaces {
				resourceProtectionCtx.Logger.Info("Removing component protection from Etcds across all namespaces")
			} else {
				resourceProtectionCtx.Logger.Info("Removing component protection from Etcd", resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace)
			}

			if err := resourceProtectionCtx.removeDisableProtectionAnnotation(context.TODO()); err != nil {
				resourceProtectionCtx.Logger.Error("Remove component protection failed", err)
				return err
			}

			resourceProtectionCtx.Logger.Success("Component protection removed successfully")
			return nil
		},
	}
}

func getResourceProtection(cmd *cobra.Command, args []string, options *types.Options) (*resourceProtectionCommandContext, error) {
	// Create command context with all common functionality
	cmdCtx, err := types.NewCommandContext(cmd, args, options)
	if err != nil {
		return nil, err
	}
	if err := cmdCtx.Validate(); err != nil {
		return nil, err
	}

	// Create typed etcd client
	etcdClient, err := cmdCtx.ClientFactory.CreateTypedEtcdClient()
	if err != nil {
		cmdCtx.Logger.Error("Unable to create etcd client: ", err)
		return nil, err
	}

	resourceProtectionCtx := newResourceProtectionCommandContext(cmdCtx, etcdClient)
	if err := resourceProtectionCtx.validate(); err != nil {
		return nil, err
	}
	return resourceProtectionCtx, nil
}
