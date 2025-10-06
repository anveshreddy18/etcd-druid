package output

import (
	"github.com/gardener/etcd-druid/druidctl/pkg/output/charm"
)

func DefaultService() Service {
	return charm.NewCharmService()
}

func NewService(serviceType OutputType) Service {
	switch serviceType {
	case OutputTypeCharm:
		return charm.NewCharmService()
	default:
		return DefaultService()
	}
}
