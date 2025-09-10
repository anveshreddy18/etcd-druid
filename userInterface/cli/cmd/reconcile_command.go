package cmd

import (
	"context"
	"time"

	"github.com/gardener/etcd-druid/userInterface/core"
	"github.com/gardener/etcd-druid/userInterface/pkg/output"
	"github.com/spf13/cobra"
)

type ReconcileCommandContext struct {
	*CommandContext
	WaitTillReady bool
	Timeout       time.Duration
}

func (r *ReconcileCommandContext) Validate() error {
	// add validation logic if any
	return nil
}

// NewReconcileCommand creates the reconcile command
func NewReconcileCommand(options *Options) *cobra.Command {
	reconcileCtx := &ReconcileCommandContext{
		WaitTillReady: false,
		Timeout:       5 * time.Minute,
	}

	reconcileCmd := &cobra.Command{
		Use:   "reconcile <etcd-resource-name> --wait-till-ready(optional flag)",
		Short: "Reconcile the mentioned etcd resource",
		Long:  `Reconcile the mentioned etcd resource. If the flag --wait-till-ready is set, then reconcile only after the Etcd CR is considered ready`,
		Args:  cobra.MaximumNArgs(1),
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

			// Create reconcile command context with the command context
			reconcileCommandCtx := &ReconcileCommandContext{
				CommandContext: cmdCtx,
				WaitTillReady:  reconcileCtx.WaitTillReady,
				Timeout:        reconcileCtx.Timeout,
			}

			// Validate reconcile command context
			if err := reconcileCommandCtx.Validate(); err != nil {
				return err
			}

			// Show operation start
			output.EtcdOperation("Reconciling", reconcileCommandCtx.ResourceName, reconcileCommandCtx.Namespace, reconcileCommandCtx.AllNamespaces)

			service := core.NewEtcdReconciliationService(
				reconcileCommandCtx.EtcdClient,
				reconcileCommandCtx.WaitTillReady,
				reconcileCommandCtx.Timeout,
				reconcileCommandCtx.Verbose,
			)
			if err := service.ReconcileEtcd(context.TODO(), reconcileCommandCtx.ResourceName, reconcileCommandCtx.Namespace, reconcileCommandCtx.AllNamespaces); err != nil {
				output.EtcdOperationError("Reconciliation", err)
				return err
			}

			output.EtcdOperationSuccess("Reconciliation")
			return nil
		},
	}

	// Add command-specific flags
	reconcileCmd.Flags().BoolVarP(&reconcileCtx.WaitTillReady, "wait-till-ready", "w", false,
		"Wait until the Etcd resource is ready before reconciling")
	reconcileCmd.Flags().DurationVarP(&reconcileCtx.Timeout, "timeout", "t", 5*time.Minute,
		"Timeout for the reconciliation process")

	return reconcileCmd
}
