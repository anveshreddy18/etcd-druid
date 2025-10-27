package fake

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/client"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type TestFactory struct {
	configFlags *genericclioptions.TestConfigFlags
}

func NewTestFactory(configFlags *genericclioptions.TestConfigFlags) *TestFactory {
	return &TestFactory{configFlags: configFlags}
}

func (f *TestFactory) CreateTypedEtcdClient() (client.EtcdClientInterface, error) {
	return &FakeEtcdClient{}, nil
}

func (f *TestFactory) CreateGenericClient() (client.GenericClientInterface, error) {
	return &FakeGenericClient{}, nil
}

type FakeEtcdClient struct {
	// Add fields to store fake data as needed
	// this is a temporary field for now, we need to orchestrate fake data properly later
	etcds map[string]*druidv1alpha1.Etcd
}

func (c *FakeEtcdClient) GetEtcd(ctx context.Context, namespace, name string) (*druidv1alpha1.Etcd, error) {
	etcd, exists := c.etcds[fmt.Sprintf("%s/%s", namespace, name)]
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{}, name)
	}
	return etcd, nil
}

func (c *FakeEtcdClient) UpdateEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd, etcdModifier func(*druidv1alpha1.Etcd)) error {
	key := fmt.Sprintf("%s/%s", etcd.Namespace, etcd.Name)
	existingEtcd, exists := c.etcds[key]
	if !exists {
		return errors.NewNotFound(schema.GroupResource{}, etcd.Name)
	}
	etcdModifier(existingEtcd)
	c.etcds[key] = existingEtcd
	return nil
}

func (c *FakeEtcdClient) ListEtcds(ctx context.Context, namespace string) (*druidv1alpha1.EtcdList, error) {
	etcdList := &druidv1alpha1.EtcdList{}
	for _, etcd := range c.etcds {
		if namespace == "" || etcd.Namespace == namespace {
			etcdList.Items = append(etcdList.Items, *etcd)
		}
	}
	return etcdList, nil
}

type FakeGenericClient struct {
	// Add fields to store fake data as needed
}

func (c *FakeGenericClient) Kube() kubernetes.Interface {
	return nil
}

func (c *FakeGenericClient) Dynamic() dynamic.Interface {
	return nil
}

func (c *FakeGenericClient) Discovery() discovery.DiscoveryInterface {
	return nil
}

func (c *FakeGenericClient) RESTMapper() meta.RESTMapper {
	return nil
}
