package core

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/userInterface/pkg/output"
)

// EtcdProtectionService encapsulates logic for managing Etcd resource protection annotations.
type EtcdProtectionService struct {
	etcdClient EtcdClientInterface
	verbose    bool
	output     output.OutputService
}

// NewEtcdProtectionService creates a new service for Etcd protection annotation management.
func NewEtcdProtectionService(etcdClient EtcdClientInterface, verbose bool, output output.OutputService) *EtcdProtectionService {
	return &EtcdProtectionService{etcdClient: etcdClient, verbose: verbose, output: output}
}

// AddDisableProtectionAnnotation adds the disable protection annotation to the Etcd resource. It makes the resources vulnerable
func (s *EtcdProtectionService) AddDisableProtectionAnnotation(ctx context.Context, name, namespace string, allNamespaces bool) error {
	etcdList, err := GetEtcdList(ctx, s.etcdClient, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	if s.verbose {
		s.output.Info(fmt.Sprintf("Fetched %d etcd resources for AddDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if s.verbose {
			s.output.Info("Processing etcd", etcd.Name, etcd.Namespace)
		}
		if etcd.Annotations == nil {
			etcd.Annotations = map[string]string{}
			if s.verbose {
				s.output.Info("Initialized annotations map for etcd", etcd.Name, etcd.Namespace)
			}
		}
		etcd.Annotations[druidv1alpha1.DisableEtcdComponentProtectionAnnotation] = ""
		if s.verbose {
			s.output.Info("Set disable protection annotation for etcd", etcd.Name, etcd.Namespace)
		}
		updatedEtcd, err := s.etcdClient.UpdateEtcd(ctx, &etcd)
		if err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		s.output.Success("Added protection annotation to etcd", updatedEtcd.Name, updatedEtcd.Namespace)
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
		s.output.Info(fmt.Sprintf("Fetched %d etcd resources for RemoveDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if s.verbose {
			s.output.Info("Processing etcd", etcd.Name, etcd.Namespace)
		}
		if etcd.Annotations == nil {
			return fmt.Errorf("no annotation found to remove in ns/etcd: %s/%s", etcd.Namespace, etcd.Name)
		}
		delete(etcd.Annotations, druidv1alpha1.DisableEtcdComponentProtectionAnnotation)
		if s.verbose {
			s.output.Info("Removed disable protection annotation for etcd", etcd.Name, etcd.Namespace)
		}
		updatedEtcd, err := s.etcdClient.UpdateEtcd(ctx, &etcd)
		if err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		s.output.Success("Removed protection annotation from etcd", updatedEtcd.Name, updatedEtcd.Namespace)
	}
	return nil
}
