package resourceprotection

import (
	"context"
	"fmt"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/pkg/utils"
)

func (r *resourceProtectionCommandContext) validate() error {
	// add validation logic if needed
	return nil
}

// addDisableProtectionAnnotation adds the disable protection annotation to the Etcd resource. It makes the resources vulnerable
func (resourceProtectionCtx *resourceProtectionCommandContext) addDisableProtectionAnnotation(ctx context.Context) error {
	etcdList, err := utils.GetEtcdList(ctx, resourceProtectionCtx.etcdClient, resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace, resourceProtectionCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if resourceProtectionCtx.Verbose {
		resourceProtectionCtx.Logger.Info(fmt.Sprintf("Fetched %d etcd resources for AddDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if resourceProtectionCtx.Verbose {
			resourceProtectionCtx.Logger.Info("Processing set disable protection annotation for etcd", etcd.Name, etcd.Namespace)
		}
		etcdModifier := func(e *druidv1alpha1.Etcd) {
			if e.Annotations == nil {
				e.Annotations = map[string]string{}
			}
			e.Annotations[druidv1alpha1.DisableEtcdComponentProtectionAnnotation] = ""
		}
		if err := resourceProtectionCtx.etcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		resourceProtectionCtx.Logger.Success("Added protection annotation to etcd", etcd.Name, etcd.Namespace)
	}
	return nil
}

// removeDisableProtectionAnnotation removes the disable protection annotation from the Etcd resource. It protects the resources
func (resourceProtectionCtx *resourceProtectionCommandContext) removeDisableProtectionAnnotation(ctx context.Context) error {

	etcdList, err := utils.GetEtcdList(ctx, resourceProtectionCtx.etcdClient, resourceProtectionCtx.ResourceName, resourceProtectionCtx.Namespace, resourceProtectionCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if resourceProtectionCtx.Verbose {
		resourceProtectionCtx.Logger.Info(fmt.Sprintf("Fetched %d etcd resources for RemoveDisableProtectionAnnotation", len(etcdList.Items)))
	}

	for _, etcd := range etcdList.Items {
		if resourceProtectionCtx.Verbose {
			resourceProtectionCtx.Logger.Info("Processing remove disable protection annotation for etcd", etcd.Name, etcd.Namespace)
		}
		if etcd.Annotations == nil {
			return fmt.Errorf("no annotation found to remove in ns/etcd: %s/%s", etcd.Namespace, etcd.Name)
		}
		etcdModifier := func(e *druidv1alpha1.Etcd) {
			if e.Annotations != nil {
				delete(e.Annotations, druidv1alpha1.DisableEtcdComponentProtectionAnnotation)
			}
		}
		if err := resourceProtectionCtx.etcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
			return fmt.Errorf("unable to update etcd object: %w", err)
		}
		resourceProtectionCtx.Logger.Success("Removed protection annotation from etcd", etcd.Name, etcd.Namespace)
	}
	return nil
}
