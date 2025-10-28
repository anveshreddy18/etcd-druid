package resourceprotection

import (
	"github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
)

type resourceProtectionCommandContext struct {
	*types.GlobalOptions
	etcdClient client.EtcdClientInterface
}

func newResourceProtectionCommandContext(options *types.GlobalOptions, etcdClient client.EtcdClientInterface) *resourceProtectionCommandContext {
	return &resourceProtectionCommandContext{
		GlobalOptions: options,
		etcdClient:    etcdClient,
	}
}
