package output

import (
	"github.com/gardener/etcd-druid/druidctl/pkg/output/charm"
)

func DefaultService() OutputService {
	return charm.NewCharmService()
}

func NewService(serviceType OutputType) OutputService {
	switch serviceType {
	case OutputTypeCharm:
		return charm.NewCharmService()
	default:
		return DefaultService()
	}
}
