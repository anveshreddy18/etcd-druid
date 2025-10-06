package types

type ListResourcesCommandContext struct {
	*CommandContext
	Filter string
}

func NewListResourcesCommandContext(cmdCtx *CommandContext, filter string) *ListResourcesCommandContext {
	return &ListResourcesCommandContext{
		CommandContext: cmdCtx,
		Filter:         filter,
	}
}

func (l *ListResourcesCommandContext) Validate() error {
	// add validation logic if needed
	return nil
}
