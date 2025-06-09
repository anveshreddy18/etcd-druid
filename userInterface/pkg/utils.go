// Utility functions and shared types for userInterface module.
package pkg

import (
	"fmt"

	// druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	clientSet "github.com/gardener/etcd-druid/client/clientset/versioned"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func CreateTypedClientSet(configFlags *genericclioptions.ConfigFlags) (*clientSet.Clientset, error) {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	// Create a Kubernetes clientset
	clientset, err := clientSet.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}
	return clientset, nil
}

// function to get the etcd custom resource and annotate it and patch it or something?
