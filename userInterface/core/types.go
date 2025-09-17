package core

import (
	"context"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/client/clientset/versioned/typed/core/v1alpha1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type EtcdClientInterface interface {
	GetEtcd(ctx context.Context, namespace, name string) (*druidv1alpha1.Etcd, error)
	UpdateEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd, etcdModifier func(*druidv1alpha1.Etcd)) error
	ListEtcds(ctx context.Context, namespace string) (*druidv1alpha1.EtcdList, error)
}

type EtcdClient struct {
	client v1alpha1.DruidV1alpha1Interface
}

func NewEtcdClient(client v1alpha1.DruidV1alpha1Interface) EtcdClientInterface {
	return &EtcdClient{client: client}
}

type ClientFactory struct {
	configFlags *genericclioptions.ConfigFlags
}

func NewClientFactory(configFlags *genericclioptions.ConfigFlags) *ClientFactory {
	return &ClientFactory{configFlags: configFlags}
}
