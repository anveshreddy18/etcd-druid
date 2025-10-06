package types

import (
	client "github.com/gardener/etcd-druid/druidctl/client"
)

type ListResourcesCommandContext struct {
	*CommandContext
	EtcdClient    client.EtcdClientInterface
	GenericClient client.GenericClient
	Filter        string
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
