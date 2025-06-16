package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/core/v1alpha1"
	"github.com/gardener/etcd-druid/client/clientset/versioned/typed/core/v1alpha1"
	"github.com/gardener/etcd-druid/userInterface/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EtcdReconciliationService struct {
	client        v1alpha1.DruidV1alpha1Interface
	waitTillReady bool
	timeout       time.Duration
}

func NewEtcdReconciliationService(client v1alpha1.DruidV1alpha1Interface, waitTillReady bool, timeout time.Duration) *EtcdReconciliationService {
	return &EtcdReconciliationService{
		client:        client,
		waitTillReady: waitTillReady,
		timeout:       timeout,
	}
}

// There are two types of reconciles, one where you add the reconcile annotation and call it a day.
// Another where you wait till all the changes done to the Etcd resource have successfully reconciled and post reconciliation
// all the etcd cluster members are Ready

func (s *EtcdReconciliationService) ReconcileEtcd(ctx context.Context, name, namespace string, allNamespaces bool) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	etcdList, err := pkg.GetEtcdList(ctx, s.client, name, namespace, allNamespaces)
	if err != nil {
		return err
	}

	// Reconcile each Etcd resource
	wg := sync.WaitGroup{}
	for _, etcd := range etcdList.Items {
		wg.Add(1)
		go func(etcd druidv1alpha1.Etcd) {
			defer wg.Done()
			if err := s.reconcileEtcd(ctx, &etcd); err != nil {
				fmt.Println("Error reconciling Etcd:", etcd.Name, "Error:", err)
			}
		}(etcd)
	}
	wg.Wait()
	return nil
}

func (s *EtcdReconciliationService) reconcileEtcd(ctx context.Context, etcd *druidv1alpha1.Etcd) error {
	// first reconcile the Etcd resource
	fmt.Println("Reconciling Etcd:", etcd.Name)

	if err := s.reconcileEtcdResource(ctx, etcd); err != nil {
		return err
	}

	if s.waitTillReady {
		readyCh := make(chan struct{})
		go func() {
			if err := s.waitForEtcdReady(ctx, etcd); err != nil {
				fmt.Println("Error waiting for Etcd to be ready: %w", err)
			}
			close(readyCh)
		}()

		// Wait for the Etcd resource to be ready or for the context to be done
		select {
		case <-readyCh:
			fmt.Println("Etcd is ready:", etcd.Name)
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.timeout):
			return fmt.Errorf("timed out waiting for Etcd to be ready: %s", etcd.Name)
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
	latestEtcd, err := s.client.Etcds(etcd.Namespace).Get(ctx, etcd.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get latest etcd object: %w", err)
	}
	latestEtcd.Annotations[druidv1alpha1.DruidOperationAnnotation] = druidv1alpha1.DruidOperationReconcile
	updatedEtcd, err := s.client.Etcds(etcd.Namespace).Update(ctx, latestEtcd, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update etcd object: %w", err)
	}
	fmt.Println("Triggered reconciliation for Etcd:", updatedEtcd.Name)
	return nil
}

func (s *EtcdReconciliationService) waitForEtcdReady(ctx context.Context, etcd *druidv1alpha1.Etcd) error {
	fmt.Println("Waiting for Etcd to be ready, ns:", etcd.Namespace, "name:", etcd.Name)

	conditions := []druidv1alpha1.ConditionType{
		druidv1alpha1.ConditionTypeAllMembersUpdated,
		druidv1alpha1.ConditionTypeAllMembersReady,
	}
	// For the Etcd to be considered ready, the conditions in the conditions slice must all be set to true
	// use a ticker to periodically check the conditions
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if all conditions are met
			latestEtcd, err := s.client.Etcds(etcd.Namespace).Get(ctx, etcd.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get latest Etcd: %w", err)
			}
			if s.areEtcdConditionsMet(latestEtcd, conditions) {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *EtcdReconciliationService) areEtcdConditionsMet(etcd *druidv1alpha1.Etcd, conditions []druidv1alpha1.ConditionType) bool {
	for _, condition := range conditions {
		fmt.Println("Checking condition:", condition, "for Etcd:", etcd.Name)
		if !s.isEtcdConditionTrue(etcd, condition) {
			return false
		}
		fmt.Println("Condition met:", condition, "for Etcd:", etcd.Name)
	}
	return true
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
