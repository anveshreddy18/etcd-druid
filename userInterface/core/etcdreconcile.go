package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/userInterface/pkg/output"
)

type EtcdReconciliationService struct {
	etcdClient    EtcdClientI
	waitTillReady bool
	timeout       time.Duration
	verbose       bool
}

type ReconcileResult struct {
	Etcd     *druidv1alpha1.Etcd
	Success  bool
	Error    error
	Duration time.Duration
}

func NewEtcdReconciliationService(etcdClient EtcdClientI, waitTillReady bool, timeout time.Duration, verbose bool) *EtcdReconciliationService {
	return &EtcdReconciliationService{
		etcdClient:    etcdClient,
		waitTillReady: waitTillReady,
		timeout:       timeout,
		verbose:       verbose,
	}
}

// There are two types of reconciles, one where you add the reconcile annotation and call it a day.
// Another where you wait till all the changes done to the Etcd resource have successfully reconciled and post reconciliation
// all the etcd cluster members are Ready

func (s *EtcdReconciliationService) ReconcileEtcd(ctx context.Context, name, namespace string, allNamespaces bool) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	etcdList, err := GetEtcdList(ctx, s.etcdClient, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	results := make([]*ReconcileResult, 0, len(etcdList.Items))
	resultChan := make(chan *ReconcileResult, len(etcdList.Items))

	wg := sync.WaitGroup{}

	// Reconcile each Etcd resource
	for _, etcd := range etcdList.Items {
		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			startTime := time.Now()
			err := s.reconcileEtcd(ctx, &etcd)
			resultChan <- &ReconcileResult{
				Etcd:     &etcd,
				Success:  err == nil,
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
		if result.Success {
			output.Success(fmt.Sprintf("Reconciliation successful for etcd '%s/%s' in %s", result.Etcd.Namespace, result.Etcd.Name, result.Duration))
		} else {
			output.EtcdOperationError(fmt.Sprintf("Reconciliation failed for etcd '%s/%s' in %s", result.Etcd.Namespace, result.Etcd.Name, result.Duration), result.Error)
		}
	}

	// If any reconciliation failed, return an error
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("one or more reconciliations failed")
		}
	}

	return nil
}

func (s *EtcdReconciliationService) reconcileEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd) error {
	output.StartedProgressMessage(fmt.Sprintf("Starting reconciliation for etcd '%s/%s'", etcd.Namespace, etcd.Name))

	// first reconcile the Etcd resource
	if err := s.reconcileEtcdResource(ctx, etcd); err != nil {
		return err
	}

	if s.waitTillReady {
		if err := s.waitForEtcdReady(ctx, etcd); err != nil {
			return fmt.Errorf("error waiting for Etcd '%s/%s' to be ready: %w", etcd.Namespace, etcd.Name, err)
		}
	}
	return nil
}

func (s *EtcdReconciliationService) reconcileEtcdResource(ctx context.Context, etcd *druidv1alpha1.Etcd) error {
	if etcd.Annotations == nil {
		etcd.Annotations = make(map[string]string)
	}
	etcd.Annotations[druidv1alpha1.DruidOperationAnnotation] = druidv1alpha1.DruidOperationReconcile
	// fetch the latest etcd and use that to update after that
	latestEtcd, err := s.etcdClient.GetEtcd(ctx, etcd.Namespace, etcd.Name)
	if err != nil {
		return fmt.Errorf("unable to get latest etcd object for '%s/%s': %w", etcd.Namespace, etcd.Name, err)
	}
	latestEtcd.Annotations[druidv1alpha1.DruidOperationAnnotation] = druidv1alpha1.DruidOperationReconcile
	updatedEtcd, err := s.etcdClient.UpdateEtcd(ctx, latestEtcd)
	if err != nil {
		return fmt.Errorf("unable to update etcd object '%s/%s': %w", etcd.Namespace, etcd.Name, err)
	}
	output.Info(fmt.Sprintf("Triggered reconciliation for etcd '%s/%s'", updatedEtcd.Namespace, updatedEtcd.Name))
	return nil
}

func (s *EtcdReconciliationService) waitForEtcdReady(ctx context.Context, etcd *druidv1alpha1.Etcd) error {
	output.ProgressMessage(fmt.Sprintf("Waiting for etcd '%s/%s' to be ready...", etcd.Namespace, etcd.Name))

	conditions := []druidv1alpha1.ConditionType{
		druidv1alpha1.ConditionTypeAllMembersUpdated,
		druidv1alpha1.ConditionTypeAllMembersReady,
	}
	// For the Etcd to be considered ready, the conditions in the conditions slice must all be set to true
	// use a checkTicker to periodically check the conditions
	progressTicker := time.NewTicker(10 * time.Second)
	defer progressTicker.Stop()

	checkTicker := time.NewTicker(3 * time.Second)
	defer checkTicker.Stop()

	for {
		select {
		case <-progressTicker.C:
			// Check the progress
			output.ProgressMessage(fmt.Sprintf("Still waiting for etcd '%s/%s' to be ready...", etcd.Namespace, etcd.Name))
		case <-checkTicker.C:
			// Check if all conditions are met
			ready, err := s.checkEtcdConditions(ctx, etcd, conditions)
			if err != nil {
				output.Error(fmt.Sprintf("Error checking conditions for Etcd '%s/%s': %v", etcd.Namespace, etcd.Name, err))
			}
			if ready {
				output.Success(fmt.Sprintf("Etcd '%s/%s' is now ready", etcd.Namespace, etcd.Name))
				return nil
			}
		case <-ctx.Done():
			// context canceled
			return fmt.Errorf("context canceled while waiting for Etcd '%s/%s' to be ready: %w", etcd.Namespace, etcd.Name, ctx.Err())
		}
	}
}

func (s *EtcdReconciliationService) checkEtcdConditions(ctx context.Context, etcd *druidv1alpha1.Etcd, conditions []druidv1alpha1.ConditionType) (bool, error) {
	latestEtcd, err := s.etcdClient.GetEtcd(ctx, etcd.Namespace, etcd.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get latest Etcd '%s/%s': %w", etcd.Namespace, etcd.Name, err)
	}

	failingConditions := []druidv1alpha1.ConditionType{}
	for _, condition := range conditions {
		if !s.isEtcdConditionTrue(latestEtcd, condition) {
			failingConditions = append(failingConditions, condition)
		}
	}
	if len(failingConditions) > 0 {
		if s.verbose {
			output.Info(fmt.Sprintf("Etcd '%s/%s' is not ready. Failing conditions: %v", latestEtcd.Namespace, latestEtcd.Name, failingConditions))
		}
		return false, nil
	}
	return true, nil
}

func (s *EtcdReconciliationService) isEtcdConditionTrue(etcd *druidv1alpha1.Etcd, condition druidv1alpha1.ConditionType) bool {
	// Check if the specified condition is true for the Etcd resource
	for _, cond := range etcd.Status.Conditions {
		if cond.Type == condition && cond.Status == druidv1alpha1.ConditionTrue {
			return true
		}
	}
	return false
}

func GetEtcdList(ctx context.Context, cl EtcdClientI, name, namespace string, allNamespaces bool) (*druidv1alpha1.EtcdList, error) {
	etcdList := &druidv1alpha1.EtcdList{}
	var err error
	if allNamespaces {
		// list all Etcd custom resources present in the entire cluster across all namespaces.
		etcdList, err = cl.ListEtcds(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("unable to list etcd objects: %w", err)
		}
	} else {
		etcd, err := cl.GetEtcd(ctx, namespace, name)
		if err != nil {
			return nil, fmt.Errorf("unable to get etcd object: %w", err)
		}
		etcdList.Items = append(etcdList.Items, *etcd)
	}
	return etcdList, nil
}
