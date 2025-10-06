package main

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Use the user's home directory to find the kubeconfig file
	// home := homedir.HomeDir()
	kubeconfig := filepath.Join("/Users/i586337/work/etcd-druid/hack/kind/kubeconfig")

	// Create a Kubernetes client config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	// Create a discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		panic(err)
	}

	// Get a list of all server resources
	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		panic(err)
	}

	// Print the short names
	for _, apiResourceList := range apiResourceLists {
		for _, resource := range apiResourceList.APIResources {
			if len(resource.ShortNames) > 0 {
				fmt.Printf("Resource: %-20s Short Name(s): %v\n", resource.Name, resource.ShortNames)
			}
		}
	}
}
