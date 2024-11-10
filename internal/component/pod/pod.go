package pod

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gardener/etcd-druid/internal/component"
	"github.com/gardener/etcd-druid/internal/component/statefulset"
	"sigs.k8s.io/controller-runtime/pkg/client"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	druiderr "github.com/gardener/etcd-druid/internal/errors"
	appsv1 "k8s.io/api/apps/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ErrListPods  druidv1alpha1.ErrorCode = "ERR_LIST_PODS"
	ErrDeletePod druidv1alpha1.ErrorCode = "ERR_DELETE_POD"
	ErrEvictPod  druidv1alpha1.ErrorCode = "ERR_EVICT_POD"
	ErrGetLease  druidv1alpha1.ErrorCode = "ERR_GET_LEASE"
)

type _resource struct {
	client client.Client
}

func New(client client.Client) component.Operator {
	return &_resource{
		client: client,
	}
}

type podWithScore struct {
	pod   corev1.Pod
	score int
}

// GetExistingResourceNames is not required for now. We may need to fill this for testing
func (r _resource) GetExistingResourceNames(ctx component.OperatorContext, etcdObjMeta metav1.ObjectMeta) ([]string, error) {
	return nil, nil
}

// TriggerDelete is a no-op for pod component as deletion of the pods is taken care of by sts controller itself.
func (r _resource) TriggerDelete(ctx component.OperatorContext, etcdObjMeta metav1.ObjectMeta) error {
	return nil
}

// This gets called for every reconciliation of the etcd controller.
// We should ideally have a flag to turn off this reconciliation ( whenever someone sets an annotation on Etcd to turn this off? Think through )

// Here we basically have our whole OnDelete process that we thought of. For now don't include any edge cases.
// Write plain and simple to test logic. Test, Enhance, Iterate.

// first fetch all pods with the controller hash label being same as sts updateRevision
// - If any of the updated ones are unhealthy, then don't move forward. Just requeue this request.
// - Once all the updated becomes healthy => we list all the non-updated ones, doesn't matter which version they are in.
// Out of the non-updated, pick the unhealthy first, then learner(if there is one), follower, and last leader.

// If all of them are healthy, then we don't requeue again. We will just wait until the next natural reconciliation process.

// first fetch the status.updatedRevision of the statefulset named etcd.ObjectMeta.Name and store it as a string.
func (r _resource) Sync(ctx component.OperatorContext, etcd *druidv1alpha1.Etcd) error {
	ctx.Logger.Info("Running Pod Sync")

	time.Sleep(2 * time.Second)

	objKey := client.ObjectKey{
		Name:      etcd.ObjectMeta.Name,
		Namespace: etcd.ObjectMeta.Namespace,
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            etcd.ObjectMeta.Name,
			Namespace:       etcd.ObjectMeta.Namespace,
			OwnerReferences: []metav1.OwnerReference{druidv1alpha1.GetAsOwnerReference(etcd.ObjectMeta)},
		},
	}

	err := r.client.Get(context.TODO(), objKey, sts)
	if err != nil {
		return druiderr.WrapError(
			err,
			statefulset.ErrGetStatefulSet,
			"Get",
			fmt.Sprintf("Error getting StatefulSet: %v for etcd: %v", objKey, druidv1alpha1.GetNamespaceName(etcd.ObjectMeta)))
	}
	latestStsRev := sts.Status.UpdateRevision
	fmt.Printf("Anvesh:: Latest sts revision: %v\n", latestStsRev)

	// fetch all the pods under the statefulset
	podList := corev1.PodList{}
	err = r.client.List(context.TODO(), &podList, client.InNamespace(etcd.ObjectMeta.Namespace), client.MatchingLabels{"app.kubernetes.io/name": etcd.ObjectMeta.Name})
	if err != nil {
		return druiderr.WrapError(
			err,
			ErrListPods,
			"List",
			fmt.Sprintf("Error listing pods : %v", latestStsRev))
	}

	// separate out updated and non-updated pods from the podList
	nonUpdatedPodList := []corev1.Pod{}
	updatedPodList := []corev1.Pod{}
	for _, pod := range podList.Items {
		if pod.Labels["controller-revision-hash"] == latestStsRev {
			updatedPodList = append(updatedPodList, pod)
		} else {
			nonUpdatedPodList = append(nonUpdatedPodList, pod)
		}
	}

	// if all pods are up to date, then there is no need to proceed further
	if len(updatedPodList) == int(etcd.Spec.Replicas) {
		return nil
	}

	// delete all the unhealthy pods from the `nonUpdatedPodList`
	// failure to delete any returns a ErrDeletePod err
	nonUpdatedHealthyPodList := []corev1.Pod{}
	for _, pod := range nonUpdatedPodList {
		if !isEtcdContainerReady(&pod) {
			err := r.client.Delete(context.TODO(), &pod)
			if err != nil {
				return druiderr.WrapError(
					fmt.Errorf("unable to delete unhealthy pod %s", pod.Name),
					ErrDeletePod,
					"Delete",
					fmt.Sprintf("unable to delete unhealthy pod %s", pod.Name))
			}
			// if the delete is successful, we add the pod into updatedPodList
			updatedPodList = append(updatedPodList, pod)
		} else {
			// if pod is healthy, we add it to the nonUpdatedHealthyPodList
			nonUpdatedHealthyPodList = append(nonUpdatedHealthyPodList, pod)
		}
	}

	// check if all the updated pods are ready, if not requeue
	unhealthyPodList := []corev1.Pod{}
	for _, pod := range updatedPodList {
		if !isEtcdContainerReady(&pod) {
			fmt.Println("Anvesh:: Waiting for the updated pods to become healthy: ", unhealthyPodList)
			return druiderr.WrapError(
				fmt.Errorf("waiting for the updated pods %v to become healthy", unhealthyPodList),
				druiderr.ErrRequeueAfter,
				"Sync",
				fmt.Sprintf("unhealthy pods found in updated pods: %v", unhealthyPodList))
		}
	}

	// Now that there are no unhealthy pods ( both updated and nonUpdated), we can proceed to make evict calls in the order of least preference
	fmt.Printf("Anvesh:: Non Updated Healthy Pod List: %v\n", nonUpdatedHealthyPodList)

	preferredPodToEvict := podWithScore{score: 4}
	for _, pod := range nonUpdatedHealthyPodList {
		lease := coordinationv1.Lease{}
		err := r.client.Get(context.TODO(), client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, &lease)
		if err != nil {
			return druiderr.WrapError(
				err,
				ErrGetLease,
				"Get",
				fmt.Sprintf("Error getting lease for pod: %v", pod.Name))
		}
		score := getScoreFromLease(&lease)
		if score < preferredPodToEvict.score {
			preferredPodToEvict = podWithScore{pod, score}
		}
	}

	// Print the preferredPodToEvict
	fmt.Printf("Anvesh:: Preferred Pod to Evict: %s\n", preferredPodToEvict.pod.Name)
	// print it's score as well
	fmt.Printf("Anvesh:: Score of the Preferred Pod to Evict: %d\n", preferredPodToEvict.score)

	// evict the pod with the least score
	if preferredPodToEvict.score != 4 {
		pod := preferredPodToEvict.pod
		err := r.client.SubResource("eviction").Create(context.TODO(), &pod, &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{},
		})
		if err != nil {
			allEtcdContainersReady, eerr := r.areAllEtcdContainersReady(etcd.ObjectMeta)
			if eerr != nil {
				return eerr
			}
			if allEtcdContainersReady {
				// make the forced delete call to the pod
				// reason being, pdb is blocking the eviction of pod because of `backup-restore` container being unhealthy
				// we don't want to consider `backup-restore` container readiness for the eviction of the pod
				fmt.Println("Anvesh:: Force deleting the pod: ", pod.Name)
				err := r.client.Delete(context.TODO(), &pod)
				if err != nil {
					return druiderr.WrapError(
						err,
						ErrDeletePod,
						"Delete",
						fmt.Sprintf("Error force deleting pod: %v", pod.Name))
				}
				return druiderr.WrapError(
					fmt.Errorf("pod %s is deleted", pod.Name),
					druiderr.ErrRequeueAfter,
					"Sync",
					fmt.Sprintf("Pod %s is force deleted as PDB is blocking the eviction by considering `backup-restore` container readiness", pod.Name))
			}
			return druiderr.WrapError(
				err,
				ErrEvictPod,
				"Evict",
				fmt.Sprintf("Error evicting pod: %v", pod.Name))
		}
		return druiderr.WrapError(
			fmt.Errorf("pod %s is evicted", pod.Name),
			druiderr.ErrRequeueAfter,
			"Sync",
			fmt.Sprintf("Pod %s is evicted", pod.Name))
	}
	return nil
}

// PreSync is a no-op for the configmap component
func (r _resource) PreSync(_ component.OperatorContext, _ *druidv1alpha1.Etcd) error { return nil }

// isEtcdContainerReady checks if the `etcd` container in the pod is ready.
func isEtcdContainerReady(pod *corev1.Pod) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == "etcd" {
			return containerStatus.Ready
		}
	}
	return false
}

// areAllEtcdContainersReady fetches and checks if all the pods under the statefulset are ready.
func (r _resource) areAllEtcdContainersReady(etcdObjMeta metav1.ObjectMeta) (bool, error) {
	podList := corev1.PodList{}
	err := r.client.List(context.TODO(), &podList, client.InNamespace(etcdObjMeta.Namespace), client.MatchingLabels{"app.kubernetes.io/name": etcdObjMeta.Name})
	if err != nil {
		return false, druiderr.WrapError(
			err,
			ErrListPods,
			"List",
			fmt.Sprintf("Error listing pods : %v", etcdObjMeta.Name))
	}
	for _, pod := range podList.Items {
		if !isEtcdContainerReady(&pod) {
			return false, nil
		}
	}
	return true, nil
}

// getScoreFromLease returns the score of the pod based on the role of the pod defined in the corresponding lease.
func getScoreFromLease(lease *coordinationv1.Lease) int {
	holderIdentity := lease.Spec.HolderIdentity
	if holderIdentity == nil {
		return 0
	}
	role := strings.Split(*holderIdentity, ":")[1]
	switch role {
	case "Learner":
		return 1
	case "Member":
		return 2
	case "Leader":
		return 3
	default:
		return 0
	}
}
