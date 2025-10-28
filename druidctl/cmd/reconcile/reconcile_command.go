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

var (
	reconcileExample = `
		# Reconcile an Etcd resource named "my-etcd" in the default namespace
		druidctl reconcile my-etcd --namespace default
		
		# Reconcile an Etcd resource named "my-etcd" in all namespaces
		druidctl reconcile my-etcd --all-namespaces
		
		# Reconcile an Etcd resource named "my-etcd" in the default namespace and wait until it's ready
		druidctl reconcile my-etcd --namespace default --wait-till-ready

		# Reconcile an Etcd resource named "my-etcd" in the default namespace with a custom timeout
		druidctl reconcile my-etcd --namespace default --wait-till-ready --timeout=10m`

	suspendReconcileExample = `
		# Suspend reconciliation for an Etcd resource named "my-etcd" in the default namespace
		druidctl suspend-reconcile my-etcd --namespace default
		
		# Suspend reconciliation for an Etcd resource named "my-etcd" in all namespaces
		druidctl suspend-reconcile my-etcd --all-namespaces`

	resumeReconcileExample = `
		# Resume reconciliation for an Etcd resource named "my-etcd" in the default namespace
		druidctl resume-reconcile my-etcd --namespace default
		
		# Resume reconciliation for an Etcd resource named "my-etcd" in all namespaces
		druidctl resume-reconcile my-etcd --all-namespaces`
)

// group the Use, Short, Long and Example for the reconcile commands into a structure
type reconcileCommandInfo struct {
	use     string
	short   string
	long    string
	example string
}

func newReconcileBaseCommand(
	cmdInfo *reconcileCommandInfo,
	options *types.GlobalOptions,
	createReconcileContext func(*types.CommandContext) (reconcileContext, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     cmdInfo.use,
		Short:   cmdInfo.short,
		Long:    cmdInfo.long,
		Example: cmdInfo.example,
		Args:    cobra.MaximumNArgs(1),
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
				cmdCtx.Logger.Error(fmt.Sprintf("%s validation failed", getOperationName(cmdInfo.use)), err)
				return err
			}

			if cmdCtx.AllNamespaces {
				cmdCtx.Logger.Info(fmt.Sprintf("%s Etcd resources across all namespaces", getOperationName(cmdInfo.use)))
			} else {
				cmdCtx.Logger.Info(fmt.Sprintf("%s Etcd resource", getOperationName(cmdInfo.use)), cmdCtx.ResourceName, cmdCtx.Namespace)
			}

			if err := reconcileContext.execute(context.TODO()); err != nil {
				cmdCtx.Logger.Error(fmt.Sprintf("%s failed", getOperationName(cmdInfo.use)), err)
				return err
			}
			cmdCtx.Logger.Success(fmt.Sprintf("%s completed successfully", getOperationName(cmdInfo.use)))
			return nil
		},
	}
	return cmd
}

// NewReconcileCommand creates the reconcile command
func NewReconcileCommand(options *types.GlobalOptions) *cobra.Command {
	var waitTillReady bool
	var timeout time.Duration = defaultTimeout

	cmdInfo := &reconcileCommandInfo{
		use:     "reconcile <etcd-resource-name> --wait-till-ready(optional flag) --timeout(optional flag)",
		short:   "Reconcile the mentioned etcd resource",
		long:    "Reconcile the mentioned etcd resource. If the flag --wait-till-ready is set, then reconcile only after the Etcd CR is considered ready",
		example: reconcileExample,
	}

	reconcileCmd := newReconcileBaseCommand(
		cmdInfo,
		options,
		func(cmdCtx *types.CommandContext) (reconcileContext, error) {
			etcdClient, err := cmdCtx.Clients.EtcdClient()
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
func NewSuspendReconcileCommand(options *types.GlobalOptions) *cobra.Command {
	cmdInfo := &reconcileCommandInfo{
		use:     "suspend-reconcile <etcd-resource-name>",
		short:   "Suspend reconciliation for the mentioned etcd resource",
		long:    "Suspend reconciliation for the mentioned etcd resource.",
		example: suspendReconcileExample,
	}
	suspendReconcileCmd := newReconcileBaseCommand(
		cmdInfo,
		options,
		func(cmdCtx *types.CommandContext) (reconcileContext, error) {
			etcdClient, err := cmdCtx.Clients.EtcdClient()
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
func NewResumeReconcileCommand(options *types.GlobalOptions) *cobra.Command {
	cmdInfo := &reconcileCommandInfo{
		use:     "resume-reconcile <etcd-resource-name>",
		short:   "Resume reconciliation for the mentioned etcd resource",
		long:    "Resume reconciliation for the mentioned etcd resource.",
		example: resumeReconcileExample,
	}
	resumeReconcileCmd := newReconcileBaseCommand(
		cmdInfo,
		options,
		func(cmdCtx *types.CommandContext) (reconcileContext, error) {
			etcdClient, err := cmdCtx.Clients.EtcdClient()
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
		return "Suspending reconciliation"
	case "resume-reconcile":
		return "Resuming reconciliation"
	case "reconcile":
		return "Reconciling"
	}
	return "Processing"
}
