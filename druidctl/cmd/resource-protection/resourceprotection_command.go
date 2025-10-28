package resourceprotection

import (
	"context"

	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/spf13/cobra"
)

var (
	addProtectionExample = `
		# Add component protection to an Etcd resource named "my-etcd" in the default namespace
		druidctl add-component-protection my-etcd --namespace default
		
		# Add component protection to all Etcd resources in all namespaces
		druidctl add-component-protection my-etcd --all-namespaces`

	removeProtectionExample = `
		# Remove component protection from an Etcd resource named "my-etcd" in the default namespace
		druidctl remove-component-protection my-etcd --namespace default
		
		# Remove component protection from all Etcd resources in all namespaces
		druidctl remove-component-protection my-etcd --all-namespaces`
)

// Create add-component-protection subcommand
func NewAddProtectionCommand(options *types.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "add-component-protection <etcd-resource-name>",
		Short: "Adds resource protection to all managed components for a given etcd cluster",
		Long: `Adds resource protection to all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Example: addProtectionExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceProtectionCtx, err := getResourceProtection(options)
			if err != nil {
				return err
			}
			options.Logger.SetOutput(options.IOStreams.Out)
			if options.AllNamespaces {
				options.Logger.Info("Adding component protection to all namespaces")
			} else {
				options.Logger.Info("Adding component protection to Etcd", options.ResourceName, options.Namespace)
			}

			if err := resourceProtectionCtx.removeDisableProtectionAnnotation(context.TODO()); err != nil {
				options.Logger.SetOutput(options.IOStreams.ErrOut)
				options.Logger.Error("Add component protection failed", err)
				return err
			}

			options.Logger.Success("Component protection added successfully")
			return nil
		},
	}
}

// Create remove-component-protection subcommand
func NewRemoveProtectionCommand(options *types.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-component-protection <etcd-resource-name>",
		Short: "Removes resource protection for all managed components for a given etcd cluster",
		Long: `Removes resource protection for all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Example: removeProtectionExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceProtectionCtx, err := getResourceProtection(options)
			if err != nil {
				return err
			}
			options.Logger.SetOutput(options.IOStreams.Out)
			if options.AllNamespaces {
				options.Logger.Info("Removing component protection from Etcds across all namespaces")
			} else {
				options.Logger.Info("Removing component protection from Etcd", options.ResourceName, options.Namespace)
			}

			if err := resourceProtectionCtx.addDisableProtectionAnnotation(context.TODO()); err != nil {
				options.Logger.SetOutput(options.IOStreams.ErrOut)
				options.Logger.Error("Remove component protection failed", err)
				return err
			}

			options.Logger.Success("Component protection removed successfully")
			return nil
		},
	}
}

func getResourceProtection(options *types.GlobalOptions) (*resourceProtectionCommandContext, error) {
	etcdClient, err := options.Clients.EtcdClient()
	if err != nil {
		options.Logger.Error("Unable to create etcd client: ", err)
		return nil, err
	}

	resourceProtectionCtx := newResourceProtectionCommandContext(options, etcdClient)
	if err := resourceProtectionCtx.validate(); err != nil {
		return nil, err
	}
	return resourceProtectionCtx, nil
}
