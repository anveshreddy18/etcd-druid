// Utility functions and shared types for druidctl module.
package tui

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	druidclientet "github.com/gardener/etcd-druid/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

// CreateGenericClientSet returns a client-go kubernetes.Interface for native resources
func CreateGenericClientSet(configFlags *genericclioptions.ConfigFlags) (kubernetes.Interface, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}
	// Create a Generic Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create generic kubernetes clientset: %w", err)
	}
	return clientset, nil
}

// ListAllEtcds lists all Etcd resources across all namespaces.
func ListAllEtcds(ctx context.Context, cs *druidclientet.Clientset) ([]druidv1alpha1.Etcd, error) {
	etcds, err := cs.DruidV1alpha1().Etcds("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Etcds: %w", err)
	}
	return etcds.Items, nil
}
