package listresources

import (
	"context"
	"fmt"

	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/spf13/cobra"
)

const defaultFilter = "all"

var (
	example = `
		# List all managed resources for the etcd resource named 'my-etcd' in the 'default' namespace
		druidctl list-resources my-etcd --namespace default

		# List all managed resources for all etcd resources across all namespaces
		druidctl list-resources --all-namespaces

		# List only the Secrets and ConfigMaps managed resources for the etcd resource named 'my-etcd' in the 'default' namespace
		druidctl list-resources my-etcd --namespace default --filter=secrets,configmaps

		# List all managed resources for the etcd resource named 'my-etcd' in the 'default' namespace in JSON format
		druidctl list-resources my-etcd --namespace default --output=json
		
		# List all managed resources for all etcd resources across all namespaces in YAML format
		druidctl list-resources --all-namespaces --output=yaml
	`
)

// NewListResourcesCommand creates the list-resources command
func NewListResourcesCommand(options *types.Options) *cobra.Command {
	listResourcesCommandCtx := newListResourcesCommandContext(nil, defaultFilter)

	listResourcesCmd := &cobra.Command{
		Use:     "list-resources <etcd-resource-name> --filter=<comma separated types> (optional flag) --output=<output-format> (optional flag)",
		Short:   "List managed resources for an etcd cluster filtered by the specified types",
		Long:    `List managed resources for an etcd cluster filtered by the specified types. If no types are specified, all managed resources will be listed.`,
		Args:    cobra.MaximumNArgs(1),
		Example: example,
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
			if err := listResourcesCommandCtx.validate(); err != nil {
				return err
			}

			// Create typed etcd client
			etcdClient, err := listResourcesCommandCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				listResourcesCommandCtx.Logger.Error("Unable to create etcd client: ", err)
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
				listResourcesCommandCtx.Logger.Info("Listing all Managed resources for Etcds across all namespaces")
			} else {
				listResourcesCommandCtx.Logger.Info("Listing Managed resources for Etcds in namespace", listResourcesCommandCtx.Namespace)
			}

			if err := listResourcesCommandCtx.execute(context.TODO()); err != nil {
				listResourcesCommandCtx.Logger.Error("Listing Managed resources for Etcds failed", err)
				return err
			}

			listResourcesCommandCtx.Logger.Success("Listing Managed resources for Etcds completed successfully")
			return nil
		},
	}

	listResourcesCmd.Flags().StringVarP(&listResourcesCommandCtx.Filter, "filter", "f", defaultFilter, "Comma-separated list of resource types to include (short or full names). Use 'all' for a curated default set.")

	return listResourcesCmd
}
