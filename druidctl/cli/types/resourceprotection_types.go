package types

import (
	client "github.com/gardener/etcd-druid/druidctl/client"
)

type ResourceProtectionCommandContext struct {
	*CommandContext
	EtcdClient client.EtcdClientInterface
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
