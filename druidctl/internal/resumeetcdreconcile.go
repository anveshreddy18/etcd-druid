package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/cli/types"
)

type resumeReconcileResult struct {
	Etcd  *druidv1alpha1.Etcd
	Error error
}

// ResumeEtcdReconcile removes the suspend reconcile annotation from the Etcd resource.
func ResumeEtcdReconcile(ctx context.Context, resumeCtx *types.ResumeReconcileCommandContext) error {
	etcdList, err := GetEtcdList(ctx, resumeCtx.EtcdClient, resumeCtx.ResourceName, resumeCtx.Namespace, resumeCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if resumeCtx.Verbose {
		resumeCtx.Output.Info("Fetched etcd resources for ResumeEtcdReconcile", fmt.Sprintf("%d", len(etcdList.Items)))
	}

	results := make([]*resumeReconcileResult, 0, len(etcdList.Items))
	var wg sync.WaitGroup

	for _, etcd := range etcdList.Items {
		if resumeCtx.Verbose {
			resumeCtx.Output.Info("Processing resume reconcile for etcd", etcd.Name, etcd.Namespace)
		}

		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			err := resumeReconcile(ctx, etcd, resumeCtx)
			results = append(results, &resumeReconcileResult{
				Etcd:  &etcd,
				Error: err,
			})
		}(etcd)
	}

	wg.Wait()

	failedEtcds := make([]string, 0)
	for _, result := range results {
		if result.Error == nil {
			resumeCtx.Output.Success("Resumed reconciliation for etcd", result.Etcd.Name, result.Etcd.Namespace)
		} else {
			resumeCtx.Output.Error("Failed to resume reconciliation for etcd", result.Error, result.Etcd.Name, result.Etcd.Namespace)
			failedEtcds = append(failedEtcds, fmt.Sprintf("%s/%s", result.Etcd.Namespace, result.Etcd.Name))
		}
	}
	if len(failedEtcds) > 0 {
		resumeCtx.Output.Warning("Failed to resume reconciliation for etcd resources", failedEtcds...)
		return fmt.Errorf("failed to resume reconciliation for etcd resources: %s", strings.Join(failedEtcds, ", "))
	}
	resumeCtx.Output.Success("Resumed reconciliation for all etcd resources")
	return nil
}

func resumeReconcile(ctx context.Context, etcd druidv1alpha1.Etcd, resumeCtx *types.ResumeReconcileCommandContext) error {
	resumeCtx.Output.Start("Starting to resume reconciliation for etcd", etcd.Name, etcd.Namespace)

	etcdModifier := func(e *druidv1alpha1.Etcd) {
		delete(e.Annotations, druidv1alpha1.SuspendEtcdSpecReconcileAnnotation)
	}
	if err := resumeCtx.EtcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
		return fmt.Errorf("unable to update etcd object: %w", err)
	}
	return nil
}
