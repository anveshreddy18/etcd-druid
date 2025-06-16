package core

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/client/clientset/versioned/typed/core/v1alpha1"
	userInterfacePkg "github.com/gardener/etcd-druid/userInterface/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EtcdProtectionService encapsulates logic for managing Etcd resource protection annotations.
type EtcdProtectionService struct {
	client v1alpha1.DruidV1alpha1Interface
}

// NewEtcdProtectionService creates a new service for Etcd protection annotation management.
func NewEtcdProtectionService(client v1alpha1.DruidV1alpha1Interface) *EtcdProtectionService {
	return &EtcdProtectionService{client: client}
}

// AddDisableProtectionAnnotation adds the disable protection annotation to the Etcd resource. It makes the resources vulnerable
func (s *EtcdProtectionService) AddDisableProtectionAnnotation(ctx context.Context, name, namespace string, allNamespaces bool) error {
	etcdList, err := userInterfacePkg.GetEtcdList(ctx, s.client, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	for _, etcd := range etcdList.Items {
		if etcd.Annotations == nil {
			etcd.Annotations = map[string]string{}
		}
		etcd.Annotations[druidv1alpha1.DisableEtcdComponentProtectionAnnotation] = ""
		updatedEtcd, err := s.client.Etcds(namespace).Update(ctx, &etcd, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		fmt.Println("Added protection annotation to Etcd:", updatedEtcd.Name)
	}
	return nil
}

// RemoveDisableProtectionAnnotation removes the disable protection annotation from the Etcd resource. It protects the resources
func (s *EtcdProtectionService) RemoveDisableProtectionAnnotation(ctx context.Context, name, namespace string, allNamespaces bool) error {

	etcdList, err := userInterfacePkg.GetEtcdList(ctx, s.client, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	for _, etcd := range etcdList.Items {
		if etcd.Annotations == nil {
			return fmt.Errorf("no annotation found to remove in ns/etcd: %s/%s", etcd.Namespace, etcd.Name)
		}
		delete(etcd.Annotations, druidv1alpha1.DisableEtcdComponentProtectionAnnotation)
		updatedEtcd, err := s.client.Etcds(namespace).Update(ctx, &etcd, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		fmt.Println("Removed protection annotation from Etcd:", updatedEtcd.Name)
	}
	return nil
}
