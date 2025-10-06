package core

import (
	"context"
	"fmt"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	druidclientet "github.com/gardener/etcd-druid/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	cached "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/util/retry"
)

// GetEtcd fetches a single Etcd resource by name and namespace.
func (a *EtcdClient) GetEtcd(ctx context.Context, namespace, name string) (*druidv1alpha1.Etcd, error) {
	return a.client.Etcds(namespace).Get(ctx, name, metav1.GetOptions{})
}

// UpdateEtcd updates the given Etcd resource and returns the updated object.
func (a *EtcdClient) UpdateEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd, etcdModifier func(*druidv1alpha1.Etcd)) error {
	backoff := wait.Backoff{
		Steps:    5,
		Duration: 100 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
		Cap:      5 * time.Second,
	}
	return retry.OnError(backoff, func(err error) bool {
		return errors.IsConflict(err) || errors.IsServerTimeout(err) || errors.IsTooManyRequests(err)
	}, func() error {
		latestEtcd, err := a.GetEtcd(ctx, etcd.Namespace, etcd.Name)
		if err != nil {
			return fmt.Errorf("unable to fetch latest etcd object: %w", err)
		}
		updatedEtcd := latestEtcd.DeepCopy()
		etcdModifier(updatedEtcd)
		_, err = a.client.Etcds(updatedEtcd.Namespace).Update(ctx, updatedEtcd, metav1.UpdateOptions{})
		return err
	})
}

// ListEtcds lists all Etcd resources in the specified namespace. If namespace is empty, it lists across all namespaces.
func (a *EtcdClient) ListEtcds(ctx context.Context, namespace string) (*druidv1alpha1.EtcdList, error) {
	etcdList, err := a.client.Etcds(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return etcdList, nil
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

// CreateTypedEtcdClient creates and returns an EtcdClient Interface
func (f *ClientFactory) CreateTypedEtcdClient() (EtcdClientInterface, error) {
	clientSet, err := CreateTypedClientSet(f.configFlags)
	if err != nil {
		return nil, err
	}
	return NewEtcdClient(clientSet.DruidV1alpha1()), nil
}

// CreateGenericClient builds a composite GenericClient consisting of typed kube client,
// dynamic client, discovery client, and a cached RESTMapper. This is the primary entry point
// for commands that need to work with arbitrary resource types like built-ins and CRDs.
func (f *ClientFactory) CreateGenericClient() (GenericClient, error) {
	config, err := f.configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Build a deferred RESTMapper backed by cached discovery to resolve kinds/resources and short names.
	cachedDisco := cached.NewMemCacheClient(discoClient)
	deferredMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDisco)
	expander := restmapper.NewShortcutExpander(deferredMapper, cachedDisco, func(string) {})

	return &genericClient{
		kube:       kubeClient,
		dynamic:    dynClient,
		discovery:  discoClient,
		restMapper: expander,
	}, nil
}
