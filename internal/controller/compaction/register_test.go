// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package compaction

import (
	"context"
	"crypto/rand"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/gardener/etcd-druid/internal/utils"

	batchv1 "k8s.io/api/batch/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/gomega"
)

func TestSnapshotRevisionChangedForCreateEvents(t *testing.T) {
	tests := []struct {
		name                   string
		isObjectLease          bool
		objectName             string
		isHolderIdentitySet    bool
		shouldAllowCreateEvent bool
	}{
		{
			name:                   "object is not a lease object",
			isObjectLease:          false,
			objectName:             "not-a-lease",
			shouldAllowCreateEvent: false,
		},
		{
			name:                   "object is a lease object, but not a snapshot lease",
			isObjectLease:          true,
			objectName:             "different-lease",
			shouldAllowCreateEvent: false,
		},
		{
			name:                   "object is a new delta-snapshot lease, but holder identity is not set",
			isObjectLease:          true,
			objectName:             "etcd-test-delta-snap",
			isHolderIdentitySet:    false,
			shouldAllowCreateEvent: true,
		},
		{
			name:                   "object is a new delta-snapshot lease, and holder identity is set",
			isObjectLease:          true,
			objectName:             "etcd-test-delta-snap",
			isHolderIdentitySet:    true,
			shouldAllowCreateEvent: true,
		},
		{
			name:                   "object is a new full-snapshot lease, but holder identity is not set",
			isObjectLease:          true,
			objectName:             "etcd-test-full-snap",
			isHolderIdentitySet:    false,
			shouldAllowCreateEvent: true,
		},
		{
			name:                   "object is a new full-snapshot lease, and holder identity is set",
			isObjectLease:          true,
			objectName:             "etcd-test-full-snap",
			isHolderIdentitySet:    true,
			shouldAllowCreateEvent: true,
		},
	}

	g := NewWithT(t)
	t.Parallel()
	predicate := snapshotRevisionChanged()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			obj, _ := createObjectsForSnapshotLeasePredicate(g, test.objectName, test.isObjectLease, true, test.isHolderIdentitySet, false)
			g.Expect(predicate.Create(event.CreateEvent{Object: obj})).To(Equal(test.shouldAllowCreateEvent))
		})
	}
}

func TestSnapshotRevisionChangedForUpdateEvents(t *testing.T) {
	tests := []struct {
		name                    string
		isObjectLease           bool
		objectName              string
		isHolderIdentityChanged bool
		shouldAllowUpdateEvent  bool
	}{
		{
			name:                   "object is not a lease object",
			isObjectLease:          false,
			objectName:             "not-a-lease",
			shouldAllowUpdateEvent: false,
		},
		{
			name:                   "object is a lease object, but not a snapshot lease",
			isObjectLease:          true,
			objectName:             "different-lease",
			shouldAllowUpdateEvent: false,
		},
		{
			name:                    "object is a delta-snapshot lease, but holder identity is not changed",
			isObjectLease:           true,
			objectName:              "etcd-test-delta-snap",
			isHolderIdentityChanged: false,
			shouldAllowUpdateEvent:  false,
		},
		{
			name:                    "object is a delta-snapshot lease, and holder identity is changed",
			isObjectLease:           true,
			objectName:              "etcd-test-delta-snap",
			isHolderIdentityChanged: true,
			shouldAllowUpdateEvent:  true,
		},
		{
			name:                    "object is a full-snapshot lease, but holder identity is not changed",
			isObjectLease:           true,
			objectName:              "etcd-test-full-snap",
			isHolderIdentityChanged: false,
			shouldAllowUpdateEvent:  false,
		},
		{
			name:                    "object is a full-snapshot lease, and holder identity is changed",
			isObjectLease:           true,
			objectName:              "etcd-test-full-snap",
			isHolderIdentityChanged: true,
			shouldAllowUpdateEvent:  true,
		},
	}

	g := NewWithT(t)
	t.Parallel()
	predicate := snapshotRevisionChanged()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			obj, oldObj := createObjectsForSnapshotLeasePredicate(g, test.objectName, test.isObjectLease, false, true, test.isHolderIdentityChanged)
			g.Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldObj, ObjectNew: obj})).To(Equal(test.shouldAllowUpdateEvent))
		})
	}
}

func TestSnapshotRevisionChangedForDeleteEvents(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	predicate := snapshotRevisionChanged()
	obj, _ := createObjectsForSnapshotLeasePredicate(g, "etcd-test-delta-snap", true, true, true, true)
	g.Expect(predicate.Delete(event.DeleteEvent{Object: obj})).To(BeFalse())
}

func TestSnapshotRevisionChangedForGenericEvents(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()
	predicate := snapshotRevisionChanged()
	obj, _ := createObjectsForSnapshotLeasePredicate(g, "etcd-test-delta-snap", true, true, true, true)
	g.Expect(predicate.Generic(event.GenericEvent{Object: obj})).To(BeFalse())
}

func TestJobStatusChangedForUpdateEvents(t *testing.T) {
	tests := []struct {
		name                   string
		isObjectJob            bool
		isStatusChanged        bool
		shouldAllowUpdateEvent bool
	}{
		{
			name:                   "object is not a job",
			isObjectJob:            false,
			shouldAllowUpdateEvent: false,
		},
		{
			name:                   "object is a job, but status is not changed",
			isObjectJob:            true,
			isStatusChanged:        false,
			shouldAllowUpdateEvent: false,
		},
		{
			name:                   "object is a job, and status is changed",
			isObjectJob:            true,
			isStatusChanged:        true,
			shouldAllowUpdateEvent: true,
		},
	}

	g := NewWithT(t)
	t.Parallel()
	predicate := jobStatusChanged()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			obj, oldObj := createObjectsForJobStatusChangedPredicate(g, "etcd-test-compaction-job", test.isObjectJob, test.isStatusChanged)
			g.Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldObj, ObjectNew: obj})).To(Equal(test.shouldAllowUpdateEvent))
		})
	}
}

func TestEtcdCompactionAnnotation(t *testing.T) {
	test1EtcdAnnotation := map[string]string{
		"dummy-annotation": "dummy",
	}

	test2EtcdAnnotation := map[string]string{
		SafeToEvictKey: "false",
	}

	g := NewWithT(t)
	compactionAnnotation := getEtcdCompactionAnnotations(utils.MergeMaps(test1EtcdAnnotation, test2EtcdAnnotation))
	g.Expect(compactionAnnotation).To(Equal(test1EtcdAnnotation))
}

func TestJobCompletionStatusAndReason(t *testing.T) {
	tests := []struct {
		name           string
		jobConditions  []batchv1.JobCondition
		expectedStatus bool
		expectedReason string
	}{
		{
			name: "Job is successful with type Complete and reason CompletionsReached",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobComplete,
					Status: corev1.ConditionTrue,
					Reason: batchv1.JobReasonCompletionsReached,
				},
			},
			expectedStatus: true,
			expectedReason: batchv1.JobReasonCompletionsReached,
		},
		{
			name: "Job is successful with type SuccessCriteriaMet and reason CompletionsReached",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobSuccessCriteriaMet,
					Status: corev1.ConditionTrue,
					Reason: batchv1.JobReasonCompletionsReached,
				},
			},
			expectedStatus: true,
			expectedReason: batchv1.JobReasonCompletionsReached,
		},
		{
			name: "Job failed with type FailureTarget and reason DeadlineExceeded",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobFailureTarget,
					Status: corev1.ConditionTrue,
					Reason: batchv1.JobReasonDeadlineExceeded,
				},
			},
			expectedStatus: false,
			expectedReason: batchv1.JobReasonDeadlineExceeded,
		},
		{
			name: "Job failed with type Failed and reason DeadlineExceeded",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobFailed,
					Status: corev1.ConditionTrue,
					Reason: batchv1.JobReasonDeadlineExceeded,
				},
			},
			expectedStatus: false,
			expectedReason: batchv1.JobReasonDeadlineExceeded,
		},
		{
			name: "Job failed with type FailureTarget and reason BackoffLimitExceeded",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobFailureTarget,
					Status: corev1.ConditionTrue,
					Reason: batchv1.JobReasonBackoffLimitExceeded,
				},
			},
			expectedStatus: false,
			expectedReason: batchv1.JobReasonBackoffLimitExceeded,
		},
		{
			name: "Job failed with type Failed and reason BackoffLimitExceeded",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobFailed,
					Status: corev1.ConditionTrue,
					Reason: batchv1.JobReasonBackoffLimitExceeded,
				},
			},
			expectedStatus: false,
			expectedReason: batchv1.JobReasonBackoffLimitExceeded,
		},
		{
			name:           "Job has no conditions",
			jobConditions:  []batchv1.JobCondition{},
			expectedStatus: false,
			expectedReason: "",
		},
		{
			name: "Job has irrelevant conditions",
			jobConditions: []batchv1.JobCondition{
				{
					Type:   "IrrelevantCondition",
					Status: corev1.ConditionTrue,
					Reason: "IrrelevantReason",
				},
			},
			expectedStatus: false,
			expectedReason: "",
		},
	}

	g := NewWithT(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			job := &batchv1.Job{
				Status: batchv1.JobStatus{
					Conditions: test.jobConditions,
				},
			}
			status, reason := getJobCompletionStatusAndReason(job)
			g.Expect(status).To(Equal(test.expectedStatus))
			g.Expect(reason).To(Equal(test.expectedReason))
		})
	}
}

func TestPodFailureReasonAndLastTransitionTime(t *testing.T) {
	tests := []struct {
		name                   string
		podConditions          []corev1.PodCondition
		containerStatuses      []corev1.ContainerStatus
		expectedReason         podFailureReason
		expectedTransitionTime time.Time
	}{
		{
			name: "Pod has DisruptionTarget condition with reason PreemptionByScheduler",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.DisruptionTarget,
					Status: corev1.ConditionTrue,
					Reason: string(podReasonPreemptionByScheduler),
					LastTransitionTime: metav1.Time{
						Time: time.Now().Add(-time.Hour),
					},
				},
			},
			expectedReason:         podReasonPreemptionByScheduler,
			expectedTransitionTime: time.Now().Add(-time.Hour),
		},
		{
			name: "Pod has DisruptionTarget condition with reason DeletionByTaintManager",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.DisruptionTarget,
					Status: corev1.ConditionTrue,
					Reason: string(podReasonDeletionByTaintManager),
					LastTransitionTime: metav1.Time{
						Time: time.Now().Add(-2 * time.Hour),
					},
				},
			},
			expectedReason:         podReasonDeletionByTaintManager,
			expectedTransitionTime: time.Now().Add(-2 * time.Hour),
		},
		{
			name: "Pod has DisruptionTarget condition with reason EvictionByEvictionAPI",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.DisruptionTarget,
					Status: corev1.ConditionTrue,
					Reason: string(podReasonEvictionByEvictionAPI),
					LastTransitionTime: metav1.Time{
						Time: time.Now().Add(-3 * time.Hour),
					},
				},
			},
			expectedReason:         podReasonEvictionByEvictionAPI,
			expectedTransitionTime: time.Now().Add(-3 * time.Hour),
		},
		{
			name: "Pod has DisruptionTarget condition with reason TerminationByKubelet",
			podConditions: []corev1.PodCondition{
				{
					Type:   corev1.DisruptionTarget,
					Status: corev1.ConditionTrue,
					Reason: string(podReasonTerminationByKubelet),
					LastTransitionTime: metav1.Time{
						Time: time.Now().Add(-4 * time.Hour),
					},
				},
			},
			expectedReason:         podReasonTerminationByKubelet,
			expectedTransitionTime: time.Now().Add(-4 * time.Hour),
		},
		{
			name: "Pod has no DisruptionTarget condition but terminated container with process failure",
			containerStatuses: []corev1.ContainerStatus{
				{
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason: string(podReasonProcessFailure),
							FinishedAt: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
					},
				},
			},
			expectedReason:         podReasonProcessFailure,
			expectedTransitionTime: time.Now().Add(-30 * time.Minute),
		},
		{
			name:                   "Pod has no relevant conditions or terminated containers",
			podConditions:          []corev1.PodCondition{},
			containerStatuses:      []corev1.ContainerStatus{},
			expectedReason:         podReasonUnknown,
			expectedTransitionTime: time.Time{},
		},
		{
			name: "Pod has irrelevant conditions and no terminated containers",
			podConditions: []corev1.PodCondition{
				{
					Type:   "IrrelevantCondition",
					Status: corev1.ConditionTrue,
					Reason: "IrrelevantReason",
					LastTransitionTime: metav1.Time{
						Time: time.Now().Add(-time.Hour),
					},
				},
			},
			containerStatuses:      []corev1.ContainerStatus{},
			expectedReason:         podReasonUnknown,
			expectedTransitionTime: time.Time{},
		},
	}

	g := NewWithT(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			pod := &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions:        test.podConditions,
					ContainerStatuses: test.containerStatuses,
				},
			}
			reason, lastTransitionTime := getPodFailureReasonAndLastTransitionTime(pod)
			g.Expect(reason).To(Equal(test.expectedReason))
			g.Expect(lastTransitionTime).To(BeTemporally("~", test.expectedTransitionTime, time.Second))
		})
	}
}

func TestPodForJob(t *testing.T) {
	tests := []struct {
		name                      string
		job                       *batchv1.Job
		pods                      []corev1.Pod
		expectedPodNamespacedName *types.NamespacedName
		expectedError             bool
		expectedErrMsg            string
	}{
		{
			name: "Single pod matches the job selector",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"job-name": "test-job"},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "default",
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
						Labels:    map[string]string{"job-name": "test-job"},
					},
				},
			},
			expectedPodNamespacedName: &types.NamespacedName{
				Name:      "test-pod",
				Namespace: "default",
			},
			expectedError: false,
		},
		{
			name: "No pods match the job selector",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"job-name": "test-job"},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "default",
				},
			},
			pods:                      []corev1.Pod{},
			expectedPodNamespacedName: nil,
			expectedError:             true,
			expectedErrMsg:            "Pod \"test-job\" not found",
		},
		{
			name: "Multiple pods match the job selector",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"job-name": "test-job"},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "default",
				},
			},
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-1",
						Namespace: "default",
						Labels:    map[string]string{"job-name": "test-job"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod-2",
						Namespace: "default",
						Labels:    map[string]string{"job-name": "test-job"},
					},
				},
			},
			expectedPodNamespacedName: &types.NamespacedName{
				Name:      "test-pod-1",
				Namespace: "default",
			},
			expectedError: false,
		},
		{
			name: "Error converting job selector to label selector",
			job: &batchv1.Job{
				Spec: batchv1.JobSpec{
					Selector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "job-name",
								Operator: "InvalidOperator",
							},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "default",
				},
			},
			pods:                      []corev1.Pod{},
			expectedPodNamespacedName: nil,
			expectedError:             true,
			expectedErrMsg:            "InvalidOperator",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			objects := []client.Object{test.job}
			for _, pod := range test.pods {
				objects = append(objects, &pod)
			}

			// Create a fake client with the job and pods
			fakeClient := fake.NewClientBuilder().
				WithObjects(objects...).
				Build()

			pod, err := getPodForJob(context.TODO(), fakeClient, log.Log, test.job)

			if test.expectedError {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(test.expectedErrMsg))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(pod).NotTo(BeNil())
				g.Expect(pod.Name).To(Equal(test.expectedPodNamespacedName.Name))
				g.Expect(pod.Namespace).To(Equal(test.expectedPodNamespacedName.Namespace))
			}
		})
	}
}

func createObjectsForJobStatusChangedPredicate(g *WithT, name string, isJobObj, isStatusChanged bool) (obj client.Object, oldObj client.Object) {
	// if the object is not a job object, create a config map (random type chosen, could have been anything else as well).
	if !isJobObj {
		obj = createConfigMap(g, name)
		oldObj = createConfigMap(g, name)
		return
	}
	now := time.Now()
	// create job objects
	oldObj = &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1,
			StartTime: &metav1.Time{
				Time: now,
			},
		},
	}
	if isStatusChanged {
		obj = &batchv1.Job{
			Status: batchv1.JobStatus{
				Succeeded: 1,
				StartTime: &metav1.Time{
					Time: now,
				},
				CompletionTime: &metav1.Time{
					Time: time.Now(),
				},
			},
		}
	} else {
		obj = oldObj
	}
	return
}

func createObjectsForSnapshotLeasePredicate(g *WithT, name string, isLeaseObj, isNewObject, isHolderIdentitySet, isHolderIdentityChanged bool) (obj client.Object, oldObj client.Object) {
	// if the object is not a lease object, create a config map (random type chosen, could have been anything else as well).
	if !isLeaseObj {
		obj = createConfigMap(g, name)
		oldObj = createConfigMap(g, name)
		return
	}

	// create lease objects
	var holderIdentity, newHolderIdentity *string
	// if it's a new object indicating a create event, create a new lease object and return.
	if isNewObject {
		if isHolderIdentitySet {
			holderIdentity = ptr.To(strconv.Itoa(generateRandomInt(g)))
		}
		obj = createLease(name, holderIdentity)
		return
	}

	// create old and new lease objects.
	holderIdentity = ptr.To(strconv.Itoa(generateRandomInt(g)))
	oldObj = createLease(name, holderIdentity)
	if isHolderIdentityChanged {
		newHolderIdentity = ptr.To(strconv.Itoa(generateRandomInt(g)))
	} else {
		newHolderIdentity = holderIdentity
	}
	obj = createLease(name, newHolderIdentity)

	return
}

func createLease(name string, holderIdentity *string) *coordinationv1.Lease {
	return &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity: holderIdentity,
		},
	}
}

func createConfigMap(g *WithT, name string) *corev1.ConfigMap {
	randInt := generateRandomInt(g)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			"k": strconv.Itoa(randInt),
		},
	}
}

func generateRandomInt(g *WithT) int {
	randInt, err := rand.Int(rand.Reader, big.NewInt(1000))
	g.Expect(err).NotTo(HaveOccurred())
	return int(randInt.Int64())
}
