// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package condition

import (
	"context"
	"fmt"
	"time"

	druidv1alpha1 "github.com/gardener/etcd-druid/api/v1alpha1"
	"github.com/gardener/etcd-druid/pkg/utils"

	"github.com/go-logr/logr"
	coordinationv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fullSnapshotBackupReadyCheck struct {
	cl client.Client
}

type deltaSnapshotBackupReadyCheck struct {
	cl client.Client
}

const (
	// SnapshotUploadedOnSchedule is a constant that means that the etcd backup has been uploaded on schedule
	SnapshotUploadedOnSchedule string = "SnapshotUploadedOnSchedule"
	// SnapshotMissedSchedule is a constant that means that the etcd backup has missed the schedule
	SnapshotMissedSchedule string = "SnapshotMissedSchedule"
	// Unknown is a constant that means that the etcd backup status is currently not known
	Unknown string = "Unknown"
	// NotChecked is a constant that means that the etcd backup status has not been updated or rechecked
	NotChecked string = "NotChecked"
)

func (f *fullSnapshotBackupReadyCheck) Check(ctx context.Context, logger logr.Logger, etcd druidv1alpha1.Etcd) Result {
	// Default case
	result := &result{
		conType: druidv1alpha1.ConditionTypeFullSnapshotBackupReady,
		status:  druidv1alpha1.ConditionUnknown,
		reason:  Unknown,
		message: "Cannot determine full snapshot upload status",
	}
	// Special case of etcd not being configured to take snapshots
	// Do not add the BackupReady condition if backup is not configured
	if !isBackupConfigured(&etcd) {
		return nil
	}
	var (
		err, fullSnapErr     error
		fullSnapshotInterval = 1 * time.Hour
		fullSnapLease        = &coordinationv1.Lease{}
	)
	fullSnapErr = f.cl.Get(ctx, types.NamespacedName{Name: getFullSnapLeaseName(&etcd), Namespace: etcd.Namespace}, fullSnapLease)
	if fullSnapErr != nil {
		return result
	}
	fullLeaseRenewTime := fullSnapLease.Spec.RenewTime
	fullLeaseCreateTime := &fullSnapLease.ObjectMeta.CreationTimestamp

	// TODO: make etcd.Spec.Backup.FullSnapshotSchedule non-optional, since it is mandatory to
	// set the full snapshot schedule, or introduce defaulting webhook to add default value for this field
	if etcd.Spec.Backup.FullSnapshotSchedule != nil {
		if fullSnapshotInterval, err = utils.ComputeScheduleInterval(*etcd.Spec.Backup.FullSnapshotSchedule); err != nil {
			logger.Error(err, "unable to compute full snapshot duration from full snapshot schedule", "fullSnapshotSchedule", *etcd.Spec.Backup.FullSnapshotSchedule)
			return result
		}
	}

	// There are only two cases in this
	if fullLeaseRenewTime != nil {
		if time.Since(fullLeaseRenewTime.Time) < fullSnapshotInterval {
			result.status = druidv1alpha1.ConditionTrue
			result.reason = SnapshotUploadedOnSchedule
			result.message = fmt.Sprintf("Full snapshot uploaded successfully %v ago", time.Since(fullLeaseRenewTime.Time))
			return result
		} else {
			result.status = druidv1alpha1.ConditionFalse
			result.reason = SnapshotMissedSchedule
			result.message = fmt.Sprintf("Full snapshot missed schedule, last full snapshot was taken %v ago", time.Since(fullLeaseRenewTime.Time))
			return result
		}
	}
	return result
}

func (d *deltaSnapshotBackupReadyCheck) Check(ctx context.Context, logger logr.Logger, etcd druidv1alpha1.Etcd) Result {
	// Default case
	result := &result{
		conType: druidv1alpha1.ConditionTypeDeltaSnapshotBackupReady,
		status:  druidv1alpha1.ConditionUnknown,
		reason:  Unknown,
		message: "Cannot determine delta snapshot upload status",
	}
	// Special case of etcd not being configured to take snapshots
	// Do not add the BackupReady condition if backup is not configured
	if !isBackupConfigured(&etcd) {
		return nil
	}
	var (
		deltaSnapErr   error
		deltaSnapLease = &coordinationv1.Lease{}
	)
	deltaSnapErr = d.cl.Get(ctx, types.NamespacedName{Name: getDeltaSnapLeaseName(&etcd), Namespace: etcd.Namespace}, deltaSnapLease)
	if deltaSnapErr != nil {
		return result
	}
	deltaLeaseRenewTime := deltaSnapLease.Spec.RenewTime
	deltaSnapshotPeriod := etcd.Spec.Backup.DeltaSnapshotPeriod.Duration
	if deltaLeaseRenewTime == nil {
		if time.Since(deltaSnapLease.ObjectMeta.CreationTimestamp.Time) < deltaSnapshotPeriod {
			result.status = druidv1alpha1.ConditionTrue
			result.reason = SnapshotUploadedOnSchedule
			result.message = fmt.Sprintf("Delta snapshotting not started yet")
			return result
		} else if time.Since(deltaSnapLease.ObjectMeta.CreationTimestamp.Time) < 3*deltaSnapshotPeriod {
			result.status = druidv1alpha1.ConditionUnknown
			result.reason = Unknown
			result.message = fmt.Sprintf("Periodic delta snapshot not started yet")
			return result
		}
	} else {
		if time.Since(deltaLeaseRenewTime.Time) < deltaSnapshotPeriod {
			result.status = druidv1alpha1.ConditionTrue
			result.reason = SnapshotUploadedOnSchedule
			result.message = fmt.Sprintf("Delta snapshot uploaded successfully %v ago", time.Since(deltaLeaseRenewTime.Time))
			return result
		} else if time.Since(deltaLeaseRenewTime.Time) < 5*deltaSnapshotPeriod {

		}
	}

	// Cases where delta snapshot lease is not updated for a long time
	// If delta snapshot lease is present and leases aren't updated, it is safe to assume that backup is not healthy

	if etcd.Status.Conditions != nil {
		var prevBackupReadyStatus druidv1alpha1.Condition
		for _, prevBackupReadyStatus = range etcd.Status.Conditions {
			if prevBackupReadyStatus.Type == druidv1alpha1.ConditionTypeDeltaSnapshotBackupReady {
				break
			}
		}

		// Transition to "False" state only if present state is "Unknown" or "False"
		if deltaLeaseRenewTime != nil && (prevBackupReadyStatus.Status == druidv1alpha1.ConditionUnknown || prevBackupReadyStatus.Status == druidv1alpha1.ConditionFalse) {
			if time.Since(deltaLeaseRenewTime.Time) > 3*deltaSnapshotPeriod {
				result.status = druidv1alpha1.ConditionFalse
				result.reason = SnapshotMissedSchedule
				result.message = fmt.Sprintf("Delta snapshot missed schedule, last delta snapshot was taken %v ago", time.Since(deltaLeaseRenewTime.Time))
				return result
			}
		}
	}
	// Transition to "Unknown" state is we cannot prove a "True" state
	return result

}

func isBackupConfigured(etcd *druidv1alpha1.Etcd) bool {
	if etcd.Spec.Backup.Store == nil || etcd.Spec.Backup.Store.Provider == nil || len(*etcd.Spec.Backup.Store.Provider) == 0 {
		return false
	}
	return true
}

func getDeltaSnapLeaseName(etcd *druidv1alpha1.Etcd) string {
	return fmt.Sprintf("%s-delta-snap", etcd.Name)
}

func getFullSnapLeaseName(etcd *druidv1alpha1.Etcd) string {
	return fmt.Sprintf("%s-full-snap", etcd.Name)
}

// BackupReadyCheck returns a check for the "BackupReady" condition.
func BackupReadyCheck(cl client.Client) Checker {
	return &backupReadyCheck{
		cl: cl,
	}
}

func FullSnapshotBackupReadyCheck(cl client.Client) Checker {
	return &fullSnapshotBackupReadyCheck{
		cl: cl,
	}
}

func DeltaSnapshotBackupReadyCheck(cl client.Client) Checker {
	return &deltaSnapshotBackupReadyCheck{
		cl: cl,
	}
}
