package cmd

import (
	"context"
	"time"

	"github.com/gardener/etcd-druid/druidctl/cli/types"
	core "github.com/gardener/etcd-druid/druidctl/internal"
	"github.com/spf13/cobra"
)

// NewReconcileCommand creates the reconcile command
func NewReconcileCommand(options *types.Options) *cobra.Command {
	reconcileCommandCtx := types.NewReconcileCommandContext(nil, false, 5*time.Minute)

	reconcileCmd := &cobra.Command{
		Use:   "reconcile <etcd-resource-name> --wait-till-ready(optional flag)",
		Short: "Reconcile the mentioned etcd resource",
		Long:  `Reconcile the mentioned etcd resource. If the flag --wait-till-ready is set, then reconcile only after the Etcd CR is considered ready`,
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

			reconcileCommandCtx.CommandContext = cmdCtx
			if err := reconcileCommandCtx.Validate(); err != nil {
				return err
			}

			if reconcileCommandCtx.AllNamespaces {
				reconcileCommandCtx.Output.Info("Reconciling Etcd resources across all namespaces")
			} else {
				reconcileCommandCtx.Output.Info("Reconciling Etcd resource", reconcileCommandCtx.ResourceName, reconcileCommandCtx.Namespace)
			}

			if err := core.ReconcileEtcd(context.TODO(), reconcileCommandCtx); err != nil {
				reconcileCommandCtx.Output.Error("Reconciliation failed", err)
				return err
			}

			reconcileCommandCtx.Output.Success("Reconciliation completed successfully")
			return nil
		},
	}

	// Add command-specific flags
	reconcileCmd.Flags().BoolVarP(&reconcileCommandCtx.WaitTillReady, "wait-till-ready", "w", false,
		"Wait until the Etcd resource is ready before reconciling")
	reconcileCmd.Flags().DurationVarP(&reconcileCommandCtx.Timeout, "timeout", "t", 5*time.Minute,
		"Timeout for the reconciliation process")

	return reconcileCmd
}
