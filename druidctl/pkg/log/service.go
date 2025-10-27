package log

import (
	"github.com/gardener/etcd-druid/druidctl/pkg/log/charm"
)

func DefaultService() Logger {
	return charm.NewCharmService()
}

func NewLogger(logType LogType) Logger {
	switch logType {
	case LogTypeCharm:
		return charm.NewCharmService()
	default:
		return DefaultService()
	}
}
