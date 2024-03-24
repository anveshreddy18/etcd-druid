// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package condition_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	. "github.com/gardener/etcd-druid/pkg/health/condition"
	"github.com/go-logr/logr"
)

var _ = Describe("QuorumReachedCheck", func() {
	Describe("#Check", func() {
		var (
			readyMember, notReadyMember, unknownMember druidv1alpha1.EtcdMemberStatus

			logger = logr.Discard()
		)

		BeforeEach(func() {
			readyMember = druidv1alpha1.EtcdMemberStatus{
				Status: druidv1alpha1.EtcdMemberStatusReady,
			}
			notReadyMember = druidv1alpha1.EtcdMemberStatus{
				Status: druidv1alpha1.EtcdMemberStatusNotReady,
			}
			unknownMember = druidv1alpha1.EtcdMemberStatus{
				Status: druidv1alpha1.EtcdMemberStatusUnknown,
			}
		})

		Context("when members in status", func() {
			It("should return that the cluster has a quorum (all members ready)", func() {
				etcd := druidv1alpha1.Etcd{
					Status: druidv1alpha1.EtcdStatus{
						Members: []druidv1alpha1.EtcdMemberStatus{
							readyMember,
							readyMember,
							readyMember,
						},
					},
				}
				check := QuorumReachedCheck(nil)

				result := check.Check(context.TODO(), logger, etcd)

				Expect(result.ConditionType()).To(Equal(druidv1alpha1.ConditionTypeQuorumReached))
				Expect(result.Status()).To(Equal(druidv1alpha1.ConditionTrue))
			})

			It("should return that the cluster has a quorum (members are partly unknown)", func() {
				etcd := druidv1alpha1.Etcd{
					Status: druidv1alpha1.EtcdStatus{
						Members: []druidv1alpha1.EtcdMemberStatus{
							readyMember,
							unknownMember,
							unknownMember,
						},
					},
				}
				check := QuorumReachedCheck(nil)

				result := check.Check(context.TODO(), logger, etcd)

				Expect(result.ConditionType()).To(Equal(druidv1alpha1.ConditionTypeQuorumReached))
				Expect(result.Status()).To(Equal(druidv1alpha1.ConditionTrue))
			})

			It("should return that the cluster has a quorum (one member not ready)", func() {
				etcd := druidv1alpha1.Etcd{
					Status: druidv1alpha1.EtcdStatus{
						Members: []druidv1alpha1.EtcdMemberStatus{
							readyMember,
							notReadyMember,
							readyMember,
						},
					},
				}
				check := QuorumReachedCheck(nil)

				result := check.Check(context.TODO(), logger, etcd)

				Expect(result.ConditionType()).To(Equal(druidv1alpha1.ConditionTypeQuorumReached))
				Expect(result.Status()).To(Equal(druidv1alpha1.ConditionTrue))
			})

			It("should return that the cluster has lost its quorum", func() {
				etcd := druidv1alpha1.Etcd{
					Status: druidv1alpha1.EtcdStatus{
						Members: []druidv1alpha1.EtcdMemberStatus{
							readyMember,
							notReadyMember,
							notReadyMember,
						},
					},
				}
				check := QuorumReachedCheck(nil)

				result := check.Check(context.TODO(), logger, etcd)

				Expect(result.ConditionType()).To(Equal(druidv1alpha1.ConditionTypeQuorumReached))
				Expect(result.Status()).To(Equal(druidv1alpha1.ConditionFalse))
				Expect(result.Reason()).To(Equal("MajorityMembersUnready"))
			})
		})

		Context("when no members in status", func() {
			It("should return that quorum is unknown", func() {
				etcd := druidv1alpha1.Etcd{
					Status: druidv1alpha1.EtcdStatus{
						Members: []druidv1alpha1.EtcdMemberStatus{},
					},
				}
				check := QuorumReachedCheck(nil)

				result := check.Check(context.TODO(), logger, etcd)

				Expect(result.ConditionType()).To(Equal(druidv1alpha1.ConditionTypeQuorumReached))
				Expect(result.Status()).To(Equal(druidv1alpha1.ConditionUnknown))
				Expect(result.Reason()).To(Equal("NoMembersInStatus"))
			})
		})
	})
})
