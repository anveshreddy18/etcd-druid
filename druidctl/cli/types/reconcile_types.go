package types

import (
	"time"

	client "github.com/gardener/etcd-druid/druidctl/client"
)

// ReconcileCommandContext holds state and functionality specific to the reconcile command
type ReconcileCommandContext struct {
	*CommandContext
	EtcdClient    client.EtcdClientInterface
	WaitTillReady bool
	Timeout       time.Duration
}

func NewReconcileCommandContext(cmdCtx *CommandContext, waitTillReady bool, timeout time.Duration) *ReconcileCommandContext {
	return &ReconcileCommandContext{
		CommandContext: cmdCtx,
		WaitTillReady:  waitTillReady,
		Timeout:        timeout,
	}
}

func (r *ReconcileCommandContext) Validate() error {
	// add validation logic if needed
	return nil
}

// SuspendReconcileCommandContext holds state and functionality specific to the suspend-reconcile command
type SuspendReconcileCommandContext struct {
	*CommandContext
	EtcdClient client.EtcdClientInterface
}

func NewSuspendReconcileCommandContext(cmdCtx *CommandContext) *SuspendReconcileCommandContext {
	return &SuspendReconcileCommandContext{
		CommandContext: cmdCtx,
	}
}

func (s *SuspendReconcileCommandContext) Validate() error {
	// add validation logic if needed
	return nil
}

// ResumeReconcileCommandContext holds state and functionality specific to the resume-reconcile command
type ResumeReconcileCommandContext struct {
	*CommandContext
	EtcdClient client.EtcdClientInterface
}

func NewResumeReconcileCommandContext(cmdCtx *CommandContext) *ResumeReconcileCommandContext {
	return &ResumeReconcileCommandContext{
		CommandContext: cmdCtx,
	}
}

func (r *ResumeReconcileCommandContext) Validate() error {
	// add validation logic if needed
	return nil
}
