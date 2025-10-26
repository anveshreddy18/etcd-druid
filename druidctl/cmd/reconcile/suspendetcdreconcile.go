package reconcile

import (
	"context"
	"fmt"
	"sync"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/pkg/utils"
)

type suspendReconcileResult struct {
	Etcd  *druidv1alpha1.Etcd
	Error error
}

func (suspendCtx *suspendReconcileCommandContext) validate() error {
	// add validation logic if needed
	if suspendCtx.Formatter != nil {
		return fmt.Errorf("output formatting is not supported for suspend-reconcile command")
	}
	return nil
}

// execute adds the suspend reconcile annotation to the Etcd resource.
func (suspendCtx *suspendReconcileCommandContext) execute(ctx context.Context) error {
	etcdList, err := utils.GetEtcdList(ctx, suspendCtx.etcdClient, suspendCtx.ResourceName, suspendCtx.Namespace, suspendCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if suspendCtx.Verbose {
		suspendCtx.Logger.Info("Fetched etcd resources for SuspendEtcdReconcile", fmt.Sprintf("%d", len(etcdList.Items)))
	}

	results := make([]*suspendReconcileResult, 0, len(etcdList.Items))
	var wg sync.WaitGroup

	for _, etcd := range etcdList.Items {
		if suspendCtx.Verbose {
			suspendCtx.Logger.Info("Processing suspend reconcile for etcd", etcd.Name, etcd.Namespace)
		}

		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			err := suspendEtcdReconcile(ctx, etcd, suspendCtx)
			results = append(results, &suspendReconcileResult{
				Etcd:  &etcd,
				Error: err,
			})
		}(etcd)
	}

	wg.Wait()

	failedEtcds := make([]string, 0)
	for _, result := range results {
		if result.Error == nil {
			suspendCtx.Logger.Success("Suspended reconciliation for etcd", result.Etcd.Name, result.Etcd.Namespace)
		} else {
			suspendCtx.Logger.Error("Failed to suspend reconciliation for etcd", result.Error, result.Etcd.Name, result.Etcd.Namespace)
			failedEtcds = append(failedEtcds, fmt.Sprintf("%s/%s", result.Etcd.Namespace, result.Etcd.Name))
		}
	}
	if len(failedEtcds) > 0 {
		suspendCtx.Logger.Warning("Failed to suspend reconciliation for etcd resources", failedEtcds...)
		return fmt.Errorf("suspending reconciliation failed for etcd resources: %v", failedEtcds)
	}
	suspendCtx.Logger.Success("Suspended reconciliation for all etcd resources")
	return nil
}

func suspendEtcdReconcile(ctx context.Context, etcd druidv1alpha1.Etcd, suspendCtx *suspendReconcileCommandContext) error {
	suspendCtx.Logger.Start("Starting to suspend reconciliation for etcd", etcd.Name, etcd.Namespace)

	etcdModifier := func(e *druidv1alpha1.Etcd) {
		if e.Annotations == nil {
			e.Annotations = make(map[string]string)
		}
		e.Annotations[druidv1alpha1.SuspendEtcdSpecReconcileAnnotation] = "true"
	}
	if err := suspendCtx.etcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
		return fmt.Errorf("unable to update etcd object: %w", err)
	}
	return nil
}
