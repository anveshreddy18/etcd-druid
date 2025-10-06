package core

import (
	"context"
	"fmt"
	"sync"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/cli/types"
)

type suspendReconcileResult struct {
	Etcd  *druidv1alpha1.Etcd
	Error error
}

// SuspendEtcdReconcile adds the suspend reconcile annotation to the Etcd resource.
func SuspendEtcdReconcile(ctx context.Context, suspendCtx *types.SuspendReconcileCommandContext) error {
	etcdList, err := GetEtcdList(ctx, suspendCtx.EtcdClient, suspendCtx.ResourceName, suspendCtx.Namespace, suspendCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if suspendCtx.Verbose {
		suspendCtx.Output.Info("Fetched etcd resources for SuspendEtcdReconcile", fmt.Sprintf("%d", len(etcdList.Items)))
	}

	results := make([]*suspendReconcileResult, 0, len(etcdList.Items))
	var wg sync.WaitGroup

	for _, etcd := range etcdList.Items {
		if suspendCtx.Verbose {
			suspendCtx.Output.Info("Processing suspend reconcile for etcd", etcd.Name, etcd.Namespace)
		}

		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			err := suspendReconcile(ctx, etcd, suspendCtx)
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
			suspendCtx.Output.Success("Suspended reconciliation for etcd", result.Etcd.Name, result.Etcd.Namespace)
		} else {
			suspendCtx.Output.Error("Failed to suspend reconciliation for etcd", result.Error, result.Etcd.Name, result.Etcd.Namespace)
			failedEtcds = append(failedEtcds, fmt.Sprintf("%s/%s", result.Etcd.Namespace, result.Etcd.Name))
		}
	}
	if len(failedEtcds) > 0 {
		suspendCtx.Output.Warning("Failed to suspend reconciliation for etcd resources", failedEtcds...)
		return fmt.Errorf("suspending reconciliation failed for etcd resources: %v", failedEtcds)
	}
	suspendCtx.Output.Success("Suspended reconciliation for all etcd resources")
	return nil
}

func suspendReconcile(ctx context.Context, etcd druidv1alpha1.Etcd, suspendCtx *types.SuspendReconcileCommandContext) error {
	suspendCtx.Output.Start("Starting to suspend reconciliation for etcd", etcd.Name, etcd.Namespace)

	etcdModifier := func(e *druidv1alpha1.Etcd) {
		if e.Annotations == nil {
			e.Annotations = make(map[string]string)
		}
		e.Annotations[druidv1alpha1.SuspendEtcdSpecReconcileAnnotation] = "true"
	}
	if err := suspendCtx.EtcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
		return fmt.Errorf("unable to update etcd object: %w", err)
	}
	return nil
}
