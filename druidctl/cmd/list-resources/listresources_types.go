package listresources

import (
	client "github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
)

type listResourcesCommandContext struct {
	*types.CommandContext
	EtcdClient    client.EtcdClientInterface
	GenericClient client.GenericClientInterface
	Filter        string
}

func newListResourcesCommandContext(cmdCtx *types.CommandContext, filter string) *listResourcesCommandContext {
	return &listResourcesCommandContext{
		CommandContext: cmdCtx,
		Filter:         filter,
	}
}
