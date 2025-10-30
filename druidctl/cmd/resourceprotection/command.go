package resourceprotection

import (
	"context"

	cmdutils "github.com/gardener/etcd-druid/druidctl/cmd/utils"
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
func NewAddProtectionCommand(options *cmdutils.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "add-component-protection <etcd-resource-name>",
		Short: "Adds resource protection to all managed components for a given etcd cluster",
		Long: `Adds resource protection to all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Example: addProtectionExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceProtectionOptions := newResourceProtectionOptions(options)
			resourceProtectionCmdCtx := &resourceProtectionCmdCtx{
				resourceProtectionOptions: resourceProtectionOptions,
			}

			if err := resourceProtectionCmdCtx.validate(); err != nil {
				cmd.Help()
				return err
			}

			if err := resourceProtectionCmdCtx.complete(options); err != nil {
				return err
			}

			if err := resourceProtectionCmdCtx.removeDisableProtectionAnnotation(context.TODO()); err != nil {
				options.Logger.Error(options.IOStreams.ErrOut, "Add component protection failed", err)
				return err
			}

			options.Logger.Success(options.IOStreams.Out, "Component protection added successfully")
			return nil
		},
	}
}

// Create remove-component-protection subcommand
func NewRemoveProtectionCommand(options *cmdutils.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-component-protection <etcd-resource-name>",
		Short: "Removes resource protection for all managed components for a given etcd cluster",
		Long: `Removes resource protection for all managed components for a given etcd cluster.
			   NOTE: This will only have effect if resource protection webhook has been enabled when deploying etcd-druid.`,
		Example: removeProtectionExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceProtectionOptions := newResourceProtectionOptions(options)
			resourceProtectionCmdCtx := &resourceProtectionCmdCtx{
				resourceProtectionOptions: resourceProtectionOptions,
			}

			if err := resourceProtectionCmdCtx.validate(); err != nil {
				return err
			}

			if err := resourceProtectionCmdCtx.complete(options); err != nil {
				return err
			}

			if err := resourceProtectionCmdCtx.addDisableProtectionAnnotation(context.TODO()); err != nil {
				options.Logger.Error(options.IOStreams.ErrOut, "Remove component protection failed", err)
				return err
			}

			options.Logger.Success(options.IOStreams.Out, "Component protection removed successfully")
			return nil
		},
	}
}
