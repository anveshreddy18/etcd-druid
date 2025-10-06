package cmd

import (
	"context"
	"time"

	"github.com/gardener/etcd-druid/druidctl/cli/types"
	core "github.com/gardener/etcd-druid/druidctl/internal"
	"github.com/spf13/cobra"
)

// newReconcileCommand creates the reconcile command
func newReconcileCommand(options *types.Options) *cobra.Command {
	reconcileCommandCtx := types.NewReconcileCommandContext(nil, false, 5*time.Minute)

	reconcileCmd := &cobra.Command{
		Use:   "reconcile <etcd-resource-name> --wait-till-ready(optional flag) --timeout(optional flag)",
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

			// Create typed etcd client
			etcdClient, err := reconcileCommandCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				reconcileCommandCtx.Output.Error("Unable to create etcd client: ", err)
				return err
			}
			reconcileCommandCtx.EtcdClient = etcdClient

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

// newSuspendReconcileCommand creates a new suspend reconcile command.
func newSuspendReconcileCommand(options *types.Options) *cobra.Command {
	suspendReconcileCommandCtx := types.NewSuspendReconcileCommandContext(nil)

	suspendReconcileCmd := &cobra.Command{
		Use:   "suspend-reconcile <etcd-resource-name>",
		Short: "Suspend reconciliation for the mentioned etcd resource",
		Long:  `Suspend reconciliation for the mentioned etcd resource.`,
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

			suspendReconcileCommandCtx.CommandContext = cmdCtx
			if err := suspendReconcileCommandCtx.Validate(); err != nil {
				return err
			}

			// Create typed etcd client
			etcdClient, err := suspendReconcileCommandCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				suspendReconcileCommandCtx.Output.Error("Unable to create etcd client: ", err)
				return err
			}
			suspendReconcileCommandCtx.EtcdClient = etcdClient

			if suspendReconcileCommandCtx.AllNamespaces {
				suspendReconcileCommandCtx.Output.Info("Suspending reconciliation for Etcd resources across all namespaces")
			} else {
				suspendReconcileCommandCtx.Output.Info("Suspending reconciliation for Etcd resource", suspendReconcileCommandCtx.ResourceName, suspendReconcileCommandCtx.Namespace)
			}

			if err := core.SuspendEtcdReconcile(context.TODO(), suspendReconcileCommandCtx); err != nil {
				suspendReconcileCommandCtx.Output.Error("Suspending reconciliation failed", err)
				return err
			}

			suspendReconcileCommandCtx.Output.Success("Suspending reconciliation completed successfully")
			return nil
		},
	}

	return suspendReconcileCmd
}

// newResumeReconcileCommand creates a new resume reconcile command.
func newResumeReconcileCommand(options *types.Options) *cobra.Command {
	resumeReconcileCommandCtx := types.NewResumeReconcileCommandContext(nil)

	resumeReconcileCmd := &cobra.Command{
		Use:   "resume-reconcile <etcd-resource-name>",
		Short: "Resume reconciliation for the mentioned etcd resource(s)",
		Long:  `Resume reconciliation for the mentioned etcd resource(s).`,
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

			resumeReconcileCommandCtx.CommandContext = cmdCtx
			if err := resumeReconcileCommandCtx.Validate(); err != nil {
				return err
			}

			// Create typed etcd client
			etcdClient, err := resumeReconcileCommandCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				resumeReconcileCommandCtx.Output.Error("Unable to create etcd client: ", err)
				return err
			}
			resumeReconcileCommandCtx.EtcdClient = etcdClient

			if resumeReconcileCommandCtx.AllNamespaces {
				resumeReconcileCommandCtx.Output.Info("Resuming reconciliation for Etcd resources across all namespaces")
			} else {
				resumeReconcileCommandCtx.Output.Info("Resuming reconciliation for Etcd resource", resumeReconcileCommandCtx.ResourceName, resumeReconcileCommandCtx.Namespace)
			}

			if err := core.ResumeEtcdReconcile(context.TODO(), resumeReconcileCommandCtx); err != nil {
				resumeReconcileCommandCtx.Output.Error("Resuming reconciliation failed", err)
				return err
			}

			resumeReconcileCommandCtx.Output.Success("Resuming reconciliation completed successfully")
			return nil
		},
	}

	return resumeReconcileCmd
}
