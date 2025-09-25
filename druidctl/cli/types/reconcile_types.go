package types

import "time"

// ReconcileCommandContext holds state and functionality specific to the reconcile command
type ReconcileCommandContext struct {
	*CommandContext
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

func Validate(r *ReconcileCommandContext) error {
	// add validation logic if needed
	return nil
}
