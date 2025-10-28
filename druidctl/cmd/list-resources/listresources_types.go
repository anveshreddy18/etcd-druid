package listresources

import (
	client "github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
)

type listResourcesCommandContext struct {
	*types.GlobalOptions
	EtcdClient    client.EtcdClientInterface
	GenericClient client.GenericClientInterface
	Filter        string
}

func newListResourcesCommandContext(options *types.GlobalOptions, filter string) *listResourcesCommandContext {
	return &listResourcesCommandContext{
		GlobalOptions: options,
		Filter:        filter,
	}
}
