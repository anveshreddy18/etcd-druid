package core

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EtcdProtectionService encapsulates logic for managing Etcd resource protection annotations.
type EtcdProtectionService struct {
	Client druidEtcdClient
}

// @anveshreddy18 --- I feel this interface is unnecessary here. Should be removed after fair bit of thinking.
type druidEtcdClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*druidv1alpha1.Etcd, error)
	Update(ctx context.Context, etcd *druidv1alpha1.Etcd, opts metav1.UpdateOptions) (*druidv1alpha1.Etcd, error)
}

// NewEtcdProtectionService creates a new service for Etcd protection annotation management.
func NewEtcdProtectionService(client druidEtcdClient) *EtcdProtectionService {
	return &EtcdProtectionService{Client: client}
}

// AddProtectionAnnotation adds the protection annotation to the Etcd resource.
func (s *EtcdProtectionService) AddProtectionAnnotation(ctx context.Context, name string) (*druidv1alpha1.Etcd, error) {
	etcd, err := s.Client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get etcd object: %w", err)
	}
	if etcd.Annotations == nil {
		etcd.Annotations = map[string]string{}
	}
	etcd.Annotations[druidv1alpha1.DisableEtcdComponentProtectionAnnotation] = ""
	updated, err := s.Client.Update(ctx, etcd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to update etcd object: %w", err)
	}
	return updated, nil
}

// RemoveProtectionAnnotation removes the protection annotation from the Etcd resource.
func (s *EtcdProtectionService) RemoveProtectionAnnotation(ctx context.Context, name string) (*druidv1alpha1.Etcd, error) {
	etcd, err := s.Client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get etcd object: %w", err)
	}
	if etcd.Annotations == nil {
		return etcd, nil // nothing to remove
	}
	delete(etcd.Annotations, druidv1alpha1.DisableEtcdComponentProtectionAnnotation)
	updated, err := s.Client.Update(ctx, etcd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to update etcd object: %w", err)
	}
	return updated, nil
}
