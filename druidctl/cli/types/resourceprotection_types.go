package types

type ResourceProtectionCommandContext struct {
	*CommandContext
}

func NewResourceProtectionCommandContext(cmdCtx *CommandContext) *ResourceProtectionCommandContext {
	return &ResourceProtectionCommandContext{
		CommandContext: cmdCtx,
	}
}

func (r *ResourceProtectionCommandContext) Validate() error {
	// add validation logic if needed
	return nil
}
