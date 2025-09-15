package core

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	druidclientet "github.com/gardener/etcd-druid/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// GetEtcd fetches a single Etcd resource by name and namespace.
func (a *EtcdClient) GetEtcd(ctx context.Context, namespace, name string) (*druidv1alpha1.Etcd, error) {
	return a.client.Etcds(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateEtcd updates the given Etcd resource and returns the updated object.
func (a *EtcdClient) UpdateEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd) (*druidv1alpha1.Etcd, error) {
	return a.client.Etcds(etcd.Namespace).Update(ctx, etcd, metav1.UpdateOptions{})
}

// ListEtcds lists all Etcd resources in the specified namespace. If namespace is empty, it lists across all namespaces.
func (a *EtcdClient) ListEtcds(ctx context.Context, namespace string) (*druidv1alpha1.EtcdList, error) {
	etcdList, err := a.client.Etcds(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return etcdList, nil
}

// CreateTypedEtcdClient creates and returns an EtcdClient Interface
func (f *ClientFactory) CreateTypedEtcdClient() (EtcdClientInterface, error) {
	clientSet, err := CreateTypedClientSet(f.configFlags)
	if err != nil {
		return nil, err
	}
	return NewEtcdClient(clientSet.DruidV1alpha1()), nil
}

// CreateTypedClientSet creates and returns a typed Kubernetes clientset using the provided config flags.
func CreateTypedClientSet(configFlags *genericclioptions.ConfigFlags) (*druidclientet.Clientset, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	// Create a typed Kubernetes clientset for Druid managed resources
	druidclientset, err := druidclientet.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}
	return druidclientset, nil
}
