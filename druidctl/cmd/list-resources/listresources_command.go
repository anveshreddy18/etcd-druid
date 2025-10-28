package listresources

import (
	"context"

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
func NewListResourcesCommand(options *types.GlobalOptions) *cobra.Command {
	listResourcesCommandCtx := newListResourcesCommandContext(options, defaultFilter)

	listResourcesCmd := &cobra.Command{
		Use:     "list-resources <etcd-resource-name> --filter=<comma separated types> (optional flag) --output=<output-format> (optional flag)",
		Short:   "List managed resources for an etcd cluster filtered by the specified types",
		Long:    `List managed resources for an etcd cluster filtered by the specified types. If no types are specified, all managed resources will be listed.`,
		Args:    cobra.MaximumNArgs(1),
		Example: example,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := listResourcesCommandCtx.validate(); err != nil {
				return err
			}
			options.Logger.SetOutput(options.IOStreams.Out)

			etcdClient, err := options.Clients.EtcdClient()
			if err != nil {
				options.Logger.SetOutput(options.IOStreams.ErrOut)
				options.Logger.Error("Unable to create etcd client: ", err)
				return err
			}
			listResourcesCommandCtx.EtcdClient = etcdClient

			genClient, err := options.Clients.GenericClient()
			if err != nil {
				options.Logger.SetOutput(options.IOStreams.ErrOut)
				options.Logger.Error("Unable to create generic kube clients: ", err)
				return err
			}
			listResourcesCommandCtx.GenericClient = genClient

			if options.AllNamespaces {
				options.Logger.Info("Listing all Managed resources for Etcds across all namespaces")
			} else {
				options.Logger.Info("Listing Managed resources for Etcds in namespace", options.Namespace)
			}

			if err := listResourcesCommandCtx.execute(context.TODO()); err != nil {
				options.Logger.SetOutput(options.IOStreams.ErrOut)
				options.Logger.Error("Listing Managed resources for Etcds failed", err)
				return err
			}

			options.Logger.Success("Listing Managed resources for Etcds completed successfully")
			return nil
		},
	}

	listResourcesCmd.Flags().StringVarP(&listResourcesCommandCtx.Filter, "filter", "f", defaultFilter, "Comma-separated list of resource types to include (short or full names). Use 'all' for a curated default set.")

	return listResourcesCmd
}
