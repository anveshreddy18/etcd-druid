package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/druidctl/cli/types"
)

type reconcileResult struct {
	Etcd     *druidv1alpha1.Etcd
	Error    error
	Duration time.Duration
}

// There are two types of reconciles, one where you add the reconcile annotation and exit.
// Another where you wait till all the changes done to the Etcd resource have successfully reconciled and post reconciliation
// all the etcd cluster members are Ready

func ReconcileEtcd(ctx context.Context, reconcileCommandCtx *types.ReconcileCommandContext) error {
	ctx, cancel := context.WithTimeout(ctx, reconcileCommandCtx.Timeout)
	defer cancel()
	etcdList, err := GetEtcdList(ctx, reconcileCommandCtx.EtcdClient, reconcileCommandCtx.ResourceName, reconcileCommandCtx.Namespace, reconcileCommandCtx.AllNamespaces)
	if err != nil {
		return err
	}

	results := make([]*reconcileResult, 0, len(etcdList.Items))
	resultChan := make(chan *reconcileResult, len(etcdList.Items))

	wg := sync.WaitGroup{}

	// Reconcile each Etcd resource
	for _, etcd := range etcdList.Items {
		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			startTime := time.Now()
			err := reconcileEtcd(ctx, &etcd, reconcileCommandCtx)
			resultChan <- &reconcileResult{
				Etcd:     &etcd,
				Error:    err,
				Duration: time.Since(startTime),
			}
		}(etcd)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results = append(results, result)
		if result.Error == nil {
			reconcileCommandCtx.Output.Success(fmt.Sprintf("Reconciliation successful in %s", shortDuration(result.Duration)), result.Etcd.Name, result.Etcd.Namespace)
		} else {
			reconcileCommandCtx.Output.Error("Reconciliation failed", result.Error, result.Etcd.Name, result.Etcd.Namespace)
		}
	}

	// If any reconciliation failed, return an error
	for _, result := range results {
		if result.Error != nil {
			return fmt.Errorf("one or more reconciliations failed")
		}
	}

	return nil
}

func reconcileEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd, reconcileCommandCtx *types.ReconcileCommandContext) error {
	reconcileCommandCtx.Output.Start("Starting reconciliation for etcd", etcd.Name, etcd.Namespace)

	// first reconcile the Etcd resource
	if err := reconcileEtcdResource(ctx, etcd, reconcileCommandCtx); err != nil {
		return err
	}

	//  check if the reconciliation is suspended, if yes, then return error as we cannot proceed
	if _, suspended := etcd.Annotations[druidv1alpha1.SuspendEtcdSpecReconcileAnnotation]; suspended {
		return fmt.Errorf("reconciliation is suspended for Etcd, cannot proceed")
	}

	if reconcileCommandCtx.WaitTillReady {
		if err := waitForEtcdReady(ctx, etcd, reconcileCommandCtx); err != nil {
			return fmt.Errorf("error waiting for Etcd to be ready: %w", err)
		}
	}
	return nil
}

func reconcileEtcdResource(ctx context.Context, etcd *druidv1alpha1.Etcd, reconcileCommandCtx *types.ReconcileCommandContext) error {
	etcdModifier := func(e *druidv1alpha1.Etcd) {
		if e.Annotations == nil {
			e.Annotations = make(map[string]string)
		}
		e.Annotations[druidv1alpha1.DruidOperationAnnotation] = druidv1alpha1.DruidOperationReconcile
	}
	if err := reconcileCommandCtx.EtcdClient.UpdateEtcd(ctx, etcd, etcdModifier); err != nil {
		return fmt.Errorf("unable to update etcd object '%s/%s': %w", etcd.Namespace, etcd.Name, err)
	}
	reconcileCommandCtx.Output.Info("Triggered reconciliation for etcd", etcd.Name, etcd.Namespace)
	return nil
}

func waitForEtcdReady(ctx context.Context, etcd *druidv1alpha1.Etcd, reconcileCommandCtx *types.ReconcileCommandContext) error {
	reconcileCommandCtx.Output.Progress("Waiting for etcd to be ready...", etcd.Name, etcd.Namespace)

	// For the Etcd to be considered ready, the conditions in the conditions slice must all be set to true
	conditions := []druidv1alpha1.ConditionType{
		druidv1alpha1.ConditionTypeAllMembersUpdated,
		druidv1alpha1.ConditionTypeAllMembersReady,
	}
	// use a checkTicker to periodically check the conditions
	progressTicker := time.NewTicker(10 * time.Second)
	defer progressTicker.Stop()

	checkTicker := time.NewTicker(3 * time.Second)
	defer checkTicker.Stop()

	for {
		select {
		case <-progressTicker.C:
			// Check the progress
			reconcileCommandCtx.Output.Progress("Still waiting for etcd to be ready...", etcd.Name, etcd.Namespace)
		case <-checkTicker.C:
			// Check if all conditions are met
			ready, err := checkEtcdConditions(ctx, etcd, conditions, reconcileCommandCtx)
			if err != nil {
				reconcileCommandCtx.Output.Warning("Warning : failed checking conditions for Etcd", err.Error(), etcd.Name, etcd.Namespace)
			}
			if ready {
				reconcileCommandCtx.Output.Success("Etcd is now ready", etcd.Name, etcd.Namespace)
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("context canceled while waiting for Etcd to be ready: %w", ctx.Err())
		}
	}
}

func checkEtcdConditions(ctx context.Context, etcd *druidv1alpha1.Etcd, conditions []druidv1alpha1.ConditionType, reconcileCommandCtx *types.ReconcileCommandContext) (bool, error) {
	latestEtcd, err := reconcileCommandCtx.EtcdClient.GetEtcd(ctx, etcd.Namespace, etcd.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get latest Etcd: %w", err)
	}

	failingConditions := []druidv1alpha1.ConditionType{}
	for _, condition := range conditions {
		if !isEtcdConditionTrue(latestEtcd, condition) {
			failingConditions = append(failingConditions, condition)
		}
	}
	if len(failingConditions) > 0 {
		if reconcileCommandCtx.Verbose {
			reconcileCommandCtx.Output.Warning(fmt.Sprintf("Warning : Etcd is not ready. Failing conditions: %v", failingConditions), latestEtcd.Name, latestEtcd.Namespace)
		}
		return false, nil
	}
	return true, nil
}

func isEtcdConditionTrue(etcd *druidv1alpha1.Etcd, condition druidv1alpha1.ConditionType) bool {
	for _, cond := range etcd.Status.Conditions {
		if cond.Type == condition && cond.Status == druidv1alpha1.ConditionTrue {
			return true
		}
	}
	return false
}
