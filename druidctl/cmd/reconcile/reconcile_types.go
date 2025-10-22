package reconcile

import (
	"context"
	"time"

	client "github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
)

type reconcileContext interface {
	validate() error
	execute(context.Context) error
}

// reconcileCommandContext holds state and functionality specific to the reconcile command
type reconcileCommandContext struct {
	*types.CommandContext
	etcdClient    client.EtcdClientInterface
	waitTillReady bool
	timeout       time.Duration
}

func newReconcileCommandContext(cmdCtx *types.CommandContext, etcdClient client.EtcdClientInterface, waitTillReady bool, timeout time.Duration) *reconcileCommandContext {
	return &reconcileCommandContext{
		CommandContext: cmdCtx,
		etcdClient:     etcdClient,
		waitTillReady:  waitTillReady,
		timeout:        timeout,
	}
}

// suspendReconcileCommandContext holds state and functionality specific to the suspend-reconcile command
type suspendReconcileCommandContext struct {
	*types.CommandContext
	etcdClient client.EtcdClientInterface
}

func newSuspendReconcileCommandContext(cmdCtx *types.CommandContext, etcdClient client.EtcdClientInterface) *suspendReconcileCommandContext {
	return &suspendReconcileCommandContext{
		CommandContext: cmdCtx,
		etcdClient:     etcdClient,
	}
}

// resumeReconcileCommandContext holds state and functionality specific to the resume-reconcile command
type resumeReconcileCommandContext struct {
	*types.CommandContext
	etcdClient client.EtcdClientInterface
}

func newResumeReconcileCommandContext(cmdCtx *types.CommandContext, etcdClient client.EtcdClientInterface) *resumeReconcileCommandContext {
	return &resumeReconcileCommandContext{
		CommandContext: cmdCtx,
		etcdClient:     etcdClient,
	}
}
