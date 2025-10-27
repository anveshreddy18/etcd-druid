package resourceprotection

import (
	"github.com/gardener/etcd-druid/druidctl/client"
	types "github.com/gardener/etcd-druid/druidctl/cmd/types"
)

type resourceProtectionCommandContext struct {
	*types.CommandContext
	etcdClient client.EtcdClientInterface
}

func newResourceProtectionCommandContext(cmdCtx *types.CommandContext, etcdClient client.EtcdClientInterface) *resourceProtectionCommandContext {
	return &resourceProtectionCommandContext{
		CommandContext: cmdCtx,
		etcdClient:     etcdClient,
	}
}
