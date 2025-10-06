package cmd

import (
	"context"
	"fmt"

	"github.com/gardener/etcd-druid/druidctl/cli/types"
	core "github.com/gardener/etcd-druid/druidctl/internal"
	"github.com/spf13/cobra"
)

const defaultFilter = "all"

// newListResourcesCommand creates the list-resources command
func newListResourcesCommand() *cobra.Command {
	listResourcesCommandCtx := types.NewListResourcesCommandContext(nil, defaultFilter)

	listResourcesCmd := &cobra.Command{
		Use:   "list-resources <etcd-resource-name> --filter=<comma separated types> (optional flag)",
		Short: "List managed resources for an etcd cluster filtered by the specified types",
		Long:  `List managed resources for an etcd cluster filtered by the specified types. If no types are specified, all managed resources will be listed.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create command context with all common functionality
			cmdCtx, err := types.NewCommandContext(cmd, args, options)
			if err != nil {
				return err
			}
			if err := cmdCtx.Validate(); err != nil {
				return err
			}

			listResourcesCommandCtx.CommandContext = cmdCtx
			if err := listResourcesCommandCtx.Validate(); err != nil {
				return err
			}

			// Create typed etcd client
			etcdClient, err := listResourcesCommandCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				listResourcesCommandCtx.Output.Error("Unable to create etcd client: ", err)
				return err
			}
			listResourcesCommandCtx.EtcdClient = etcdClient

			// Create generic etcd client
			genClient, err := listResourcesCommandCtx.ClientFactory.CreateGenericClient()
			if err != nil {
				return fmt.Errorf("failed to create generic kube clients: %w", err)
			}
			listResourcesCommandCtx.GenericClient = genClient

			if listResourcesCommandCtx.AllNamespaces {
				listResourcesCommandCtx.Output.Info("Listing all Managed resources for Etcds across all namespaces")
			} else {
				listResourcesCommandCtx.Output.Info("Listing Managed resources for Etcds in namespace", listResourcesCommandCtx.Namespace)
			}

			if err := core.ListManagedResources(context.TODO(), listResourcesCommandCtx); err != nil {
				listResourcesCommandCtx.Output.Error("Listing Managed resources for Etcds failed", err)
				return err
			}

			listResourcesCommandCtx.Output.Success("Listing Managed resources for Etcds completed successfully")
			return nil
		},
	}

	listResourcesCmd.Flags().StringVarP(&listResourcesCommandCtx.Filter, "filter", "f", defaultFilter, "Comma-separated list of resource types to include (short or full names). Use 'all' for a curated default set.")

	return listResourcesCmd
}
