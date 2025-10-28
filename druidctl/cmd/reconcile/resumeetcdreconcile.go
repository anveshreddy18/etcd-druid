package reconcile

import (
	"context"
	"fmt"
	"strings"
	"sync"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/pkg/utils"
)

type resumeReconcileResult struct {
	Etcd  *druidv1alpha1.Etcd
	Error error
}

func (resumeCtx *resumeReconcileCommandContext) validate() error {
	if resumeCtx.Formatter != nil {
		return fmt.Errorf("output formatting is not supported for resume-reconcile command")
	}
	return nil
}

// execute removes the suspend reconcile annotation from the Etcd resource.
func (resumeCtx *resumeReconcileCommandContext) execute(ctx context.Context) error {
	etcdList, err := utils.GetEtcdList(ctx, resumeCtx.etcdClient, resumeCtx.ResourceName, resumeCtx.Namespace, resumeCtx.AllNamespaces)
	if err != nil {
		return err
	}

	if resumeCtx.Verbose {
		resumeCtx.Logger.Info("Fetched etcd resources for ResumeEtcdReconcile", fmt.Sprintf("%d", len(etcdList.Items)))
	}

	results := make([]*resumeReconcileResult, 0, len(etcdList.Items))
	var wg sync.WaitGroup

	for _, etcd := range etcdList.Items {
		if resumeCtx.Verbose {
			resumeCtx.Logger.Info("Processing resume reconcile for etcd", etcd.Name, etcd.Namespace)
		}

		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			err := resumeEtcdReconcile(ctx, etcd, resumeCtx)
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
			resumeCtx.Logger.Success("Resumed reconciliation for etcd", result.Etcd.Name, result.Etcd.Namespace)
		} else {
			resumeCtx.Logger.Error("Failed to resume reconciliation for etcd", result.Error, result.Etcd.Name, result.Etcd.Namespace)
			failedEtcds = append(failedEtcds, fmt.Sprintf("%s/%s", result.Etcd.Namespace, result.Etcd.Name))
		}
	}
	if len(failedEtcds) > 0 {
		resumeCtx.Logger.Warning("Failed to resume reconciliation for etcd resources", failedEtcds...)
		return fmt.Errorf("failed to resume reconciliation for etcd resources: %s", strings.Join(failedEtcds, ", "))
	}
	resumeCtx.Logger.Success("Resumed reconciliation for all etcd resources")
	return nil
}

func resumeEtcdReconcile(ctx context.Context, etcd druidv1alpha1.Etcd, resumeCtx *resumeReconcileCommandContext) error {
	resumeCtx.Logger.Start("Starting to resume reconciliation for etcd", etcd.Name, etcd.Namespace)

	etcdModifier := func(e *druidv1alpha1.Etcd) {
		if e.Annotations != nil {
			delete(e.Annotations, druidv1alpha1.SuspendEtcdSpecReconcileAnnotation)
		}
	}
	if err := resumeCtx.etcdClient.UpdateEtcd(ctx, &etcd, etcdModifier); err != nil {
		return fmt.Errorf("unable to update etcd object: %w", err)
	}
	return nil
}
