package core

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/userInterface/pkg/output"
)

// EtcdProtectionService encapsulates logic for managing Etcd resource protection annotations.
type EtcdProtectionService struct {
	etcdClient EtcdClientI
	verbose    bool
}

// NewEtcdProtectionService creates a new service for Etcd protection annotation management.
func NewEtcdProtectionService(etcdClient EtcdClientI, verbose bool) *EtcdProtectionService {
	return &EtcdProtectionService{etcdClient: etcdClient, verbose: verbose}
}

// AddDisableProtectionAnnotation adds the disable protection annotation to the Etcd resource. It makes the resources vulnerable
func (s *EtcdProtectionService) AddDisableProtectionAnnotation(ctx context.Context, name, namespace string, allNamespaces bool) error {
	etcdList, err := GetEtcdList(ctx, s.etcdClient, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	if s.verbose {
		output.Info(fmt.Sprintf("Fetched %d etcd resources for AddDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if s.verbose {
			output.Info(fmt.Sprintf("Processing etcd: %s/%s", etcd.Namespace, etcd.Name))
		}
		if etcd.Annotations == nil {
			etcd.Annotations = map[string]string{}
			if s.verbose {
				output.Info(fmt.Sprintf("Initialized annotations map for etcd: %s/%s", etcd.Namespace, etcd.Name))
			}
		}
		etcd.Annotations[druidv1alpha1.DisableEtcdComponentProtectionAnnotation] = ""
		if s.verbose {
			output.Info(fmt.Sprintf("Set disable protection annotation for etcd: %s/%s", etcd.Namespace, etcd.Name))
		}
		updatedEtcd, err := s.etcdClient.UpdateEtcd(ctx, &etcd)
		if err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		output.Success(fmt.Sprintf("Added protection annotation to etcd '%s'", updatedEtcd.Name))
	}
	return nil
}

// RemoveDisableProtectionAnnotation removes the disable protection annotation from the Etcd resource. It protects the resources
func (s *EtcdProtectionService) RemoveDisableProtectionAnnotation(ctx context.Context, name, namespace string, allNamespaces bool) error {

	etcdList, err := GetEtcdList(ctx, s.etcdClient, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	if s.verbose {
		output.Info(fmt.Sprintf("Fetched %d etcd resources for RemoveDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if s.verbose {
			output.Info(fmt.Sprintf("Processing etcd: %s/%s", etcd.Namespace, etcd.Name))
		}
		if etcd.Annotations == nil {
			return fmt.Errorf("no annotation found to remove in ns/etcd: %s/%s", etcd.Namespace, etcd.Name)
		}
		delete(etcd.Annotations, druidv1alpha1.DisableEtcdComponentProtectionAnnotation)
		if s.verbose {
			output.Info(fmt.Sprintf("Removed disable protection annotation for etcd: %s/%s", etcd.Namespace, etcd.Name))
		}
		updatedEtcd, err := s.etcdClient.UpdateEtcd(ctx, &etcd)
		if err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		output.Success(fmt.Sprintf("Removed protection annotation from etcd '%s/%s'", updatedEtcd.Namespace, updatedEtcd.Name))
	}
	return nil
}
