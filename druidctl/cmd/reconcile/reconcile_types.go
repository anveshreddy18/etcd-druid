package reconcile

import (
	"context"
	"time"

	"github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
)

type reconcileContext interface {
	validate() error
	execute(context.Context) error
}

// reconcileCommandContext holds state and functionality specific to the reconcile command
type reconcileCommandContext struct {
	*types.GlobalOptions
	etcdClient    client.EtcdClientInterface
	waitTillReady bool
	timeout       time.Duration
}

func newReconcileCommandContext(globalOpts *types.GlobalOptions, etcdClient client.EtcdClientInterface, waitTillReady bool, timeout time.Duration) *reconcileCommandContext {
	return &reconcileCommandContext{
		GlobalOptions: globalOpts,
		etcdClient:    etcdClient,
		waitTillReady: waitTillReady,
		timeout:       timeout,
	}
}

// suspendReconcileCommandContext holds state and functionality specific to the suspend-reconcile command
type suspendReconcileCommandContext struct {
	*types.GlobalOptions
	etcdClient client.EtcdClientInterface
}

func newSuspendReconcileCommandContext(globalOpts *types.GlobalOptions, etcdClient client.EtcdClientInterface) *suspendReconcileCommandContext {
	return &suspendReconcileCommandContext{
		GlobalOptions: globalOpts,
		etcdClient:    etcdClient,
	}
}

// resumeReconcileCommandContext holds state and functionality specific to the resume-reconcile command
type resumeReconcileCommandContext struct {
	*types.GlobalOptions
	etcdClient client.EtcdClientInterface
}

func newResumeReconcileCommandContext(globalOpts *types.GlobalOptions, etcdClient client.EtcdClientInterface) *resumeReconcileCommandContext {
	return &resumeReconcileCommandContext{
		GlobalOptions: globalOpts,
		etcdClient:    etcdClient,
	}
}
