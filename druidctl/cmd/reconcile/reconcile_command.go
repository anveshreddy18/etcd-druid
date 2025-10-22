package reconcile

import (
	"context"
	"fmt"
	"strings"
	"time"

	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
	"github.com/spf13/cobra"
)

const (
	defaultTimeout = 5 * time.Minute
)

func newReconcileBaseCommand(
	use string,
	short string,
	long string,
	options *types.Options,
	createReconcileContext func(*types.CommandContext) (reconcileContext, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
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

			reconcileContext, err := createReconcileContext(cmdCtx)
			if err != nil {
				return err
			}
			if err := reconcileContext.validate(); err != nil {
				cmdCtx.Logger.Error(fmt.Sprintf("%s validation failed", getOperationName(use)), err)
				return err
			}

			if cmdCtx.AllNamespaces {
				cmdCtx.Logger.Info(fmt.Sprintf("%s Etcd resources across all namespaces", getOperationName(use)))
			} else {
				cmdCtx.Logger.Info(fmt.Sprintf("%s Etcd resource", getOperationName(use)), cmdCtx.ResourceName, cmdCtx.Namespace)
			}

			if err := reconcileContext.execute(context.TODO()); err != nil {
				cmdCtx.Logger.Error(fmt.Sprintf("%s failed", getOperationName(use)), err)
				return err
			}
			cmdCtx.Logger.Success(fmt.Sprintf("%s completed successfully", getOperationName(use)))
			return nil
		},
	}
	return cmd
}

// NewReconcileCommand creates the reconcile command
func NewReconcileCommand(options *types.Options) *cobra.Command {
	var waitTillReady bool
	var timeout time.Duration = defaultTimeout

	reconcileCmd := newReconcileBaseCommand(
		"reconcile <etcd-resource-name> --wait-till-ready(optional flag) --timeout(optional flag)",
		"Reconcile the mentioned etcd resource",
		`Reconcile the mentioned etcd resource. If the flag --wait-till-ready is set, then reconcile only after the Etcd CR is considered ready`,
		options,
		func(cmdCtx *types.CommandContext) (reconcileContext, error) {
			etcdClient, err := cmdCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				cmdCtx.Logger.Error("Unable to create etcd client: ", err)
				return nil, err
			}

			reconcileContext := newReconcileCommandContext(cmdCtx, etcdClient, waitTillReady, timeout)
			return reconcileContext, nil
		},
	)

	// Add command-specific flags
	reconcileCmd.Flags().BoolVarP(&waitTillReady, "wait-till-ready", "w", false,
		"Wait until the Etcd resource is ready before reconciling")
	reconcileCmd.Flags().DurationVarP(&timeout, "timeout", "t", defaultTimeout,
		"Timeout for the reconciliation process")

	return reconcileCmd
}

// NewSuspendReconcileCommand creates a new suspend reconcile command.
func NewSuspendReconcileCommand(options *types.Options) *cobra.Command {
	suspendReconcileCmd := newReconcileBaseCommand(
		"suspend-reconcile <etcd-resource-name>",
		"Suspend reconciliation for the mentioned etcd resource",
		"Suspend reconciliation for the mentioned etcd resource.",
		options,
		func(cmdCtx *types.CommandContext) (reconcileContext, error) {
			etcdClient, err := cmdCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				cmdCtx.Logger.Error("Unable to create etcd client: ", err)
				return nil, err
			}

			suspendReconcileContext := newSuspendReconcileCommandContext(cmdCtx, etcdClient)
			return suspendReconcileContext, nil
		},
	)

	return suspendReconcileCmd
}

// NewResumeReconcileCommand creates a new resume reconcile command.
func NewResumeReconcileCommand(options *types.Options) *cobra.Command {
	resumeReconcileCmd := newReconcileBaseCommand(
		"resume-reconcile <etcd-resource-name>",
		"Resume reconciliation for the mentioned etcd resource",
		"Resume reconciliation for the mentioned etcd resource.",
		options,
		func(cmdCtx *types.CommandContext) (reconcileContext, error) {
			etcdClient, err := cmdCtx.ClientFactory.CreateTypedEtcdClient()
			if err != nil {
				cmdCtx.Logger.Error("Unable to create etcd client: ", err)
				return nil, err
			}

			resumeReconcileContext := newResumeReconcileCommandContext(cmdCtx, etcdClient)
			return resumeReconcileContext, nil
		},
	)

	return resumeReconcileCmd
}

func getOperationName(commandUse string) string {
	command := strings.Split(commandUse, " ")[0]
	switch command {
	case "suspend-reconcile":
		return "Suspending reconciliation for"
	case "resume-reconcile":
		return "Resuming reconciliation for"
	case "reconcile":
		return "Reconciling"
	}
	return "Processing"
}
