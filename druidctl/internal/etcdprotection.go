package core

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/cli/types"
)

// AddDisableProtectionAnnotation adds the disable protection annotation to the Etcd resource. It makes the resources vulnerable
func AddDisableProtectionAnnotation(ctx context.Context, resourceProtectionCtx *types.ResourceProtectionCommandContext) error {
	etcdList, err := GetEtcdList(ctx, resourceProtectionCtx.EtcdClient, resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace, resourceProtectionCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if resourceProtectionCtx.Verbose {
		resourceProtectionCtx.Output.Info(fmt.Sprintf("Fetched %d etcd resources for AddDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if resourceProtectionCtx.Verbose {
			resourceProtectionCtx.Output.Info("Processing set disable protection annotation for etcd", etcd.Name, etcd.Namespace)
		}
		etcdModifier := func(e *druidv1alpha1.Etcd) {
			if e.Annotations == nil {
				e.Annotations = map[string]string{}
			}
			e.Annotations[druidv1alpha1.DisableEtcdComponentProtectionAnnotation] = ""
		}
		if err := resourceProtectionCtx.EtcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		resourceProtectionCtx.Output.Success("Added protection annotation to etcd", etcd.Name, etcd.Namespace)
	}
	return nil
}

// RemoveDisableProtectionAnnotation removes the disable protection annotation from the Etcd resource. It protects the resources
func RemoveDisableProtectionAnnotation(ctx context.Context, resourceProtectionCtx *types.ResourceProtectionCommandContext) error {

	etcdList, err := GetEtcdList(ctx, resourceProtectionCtx.EtcdClient, resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace, resourceProtectionCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if resourceProtectionCtx.Verbose {
		resourceProtectionCtx.Output.Info(fmt.Sprintf("Fetched %d etcd resources for RemoveDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if resourceProtectionCtx.Verbose {
			resourceProtectionCtx.Output.Info("Processing remove disable protection annotation for etcd", etcd.Name, etcd.Namespace)
		}
		if etcd.Annotations == nil {
			return fmt.Errorf("no annotation found to remove in ns/etcd: %s/%s", etcd.Namespace, etcd.Name)
		}
		etcdModifier := func(e *druidv1alpha1.Etcd) {
			if e.Annotations != nil {
				delete(e.Annotations, druidv1alpha1.DisableEtcdComponentProtectionAnnotation)
			}
		}
		if err := resourceProtectionCtx.EtcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		resourceProtectionCtx.Output.Success("Removed protection annotation from etcd", etcd.Name, etcd.Namespace)
	}
	return nil
}
