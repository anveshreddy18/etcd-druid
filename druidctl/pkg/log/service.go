package log

import (
	"github.com/gardener/etcd-druid/druidctl/pkg/log/charm"
)

func DefaultService() Logger {
	return charm.NewCharmService()
}

func NewLogger(serviceType LogType) Logger {
	switch serviceType {
	case LogTypeCharm:
		return charm.NewCharmService()
	default:
		return DefaultService()
	}
}
