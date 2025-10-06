package cmd

import (
	"context"

	"github.com/gardener/etcd-druid/druidctl/cli/types"
	core "github.com/gardener/etcd-druid/druidctl/internal"
	"github.com/spf13/cobra"
)

// ⚠️ ⚠️ ⚠️ ⚠️ IMPLEMENTATION UNDERWAY FOR LIST RESOURCES ⚠️ ⚠️ ⚠️ ⚠️

// List all the managed resources for an etcd cluster

// Basically the filter should work with any type of resource
// pods,sts,svc,cm,pvc,secret,lease,pdb,role,rolebinding,svcaccount,etc
// identifier: resources have the label: app.kubernetes.io/part-of: <etcd-name>
// out of these listed resources above, all of them have this label.
// How many does not have OwnerReferences? pvc,pods,

// so it doesn't matter which resources the user asks for, we do a validation to confirm that such a resource exists first, if not, we throw an error.
// Now we first proceed to list all the Etcd resources ( either in a single ns or across all ns)
// For each etcd, we go through the list of resource types asked for, and find any resources that are managed by this particular etcd, as can be queried with the label `app.kubernetes.io/part-of: <etcd-name>`

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

	return listResourcesCmd
}
