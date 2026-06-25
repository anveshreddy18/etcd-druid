# etcd-druid + etcd-backup-restore: Exhaustive Capability Catalog

> Source: code-walk of `internal/controller`, `internal/component`, `internal/health`, `internal/store`, `api/core/v1alpha1`, `docs/proposals/*`, `docs/usage/*`, plus the `etcd-backup-restore` repo (`cmd/`, `pkg/snapshot/*`, `pkg/snapstore/*`, `pkg/server/httpAPI.go`, `pkg/member/*`, `pkg/leaderelection`, `docs/proposals/*`, `docs/usage/*`).
> Scope: every capability that is shipped (not just proposed) plus the live alpha feature gate.
> Differentiation key:  🔥 = unique / hard to replicate, ✨ = strong differentiator, ✓ = table-stakes for a serious etcd operator.

---

## 1. Cluster lifecycle (provisioning, scaling, upgrades, deletion)

### 1.1 Declarative single-API etcd provisioning  ✓
- **What:** A single `Etcd` custom resource (CRD `druid.gardener.cloud/v1alpha1.Etcd`) provisions everything an etcd cluster needs — StatefulSet, ConfigMap, peer/client Services, ServiceAccount + Role + RoleBinding, member Leases, snapshot Leases, PodDisruptionBudget. Replicas, storage class, capacity, quota, resources, TLS, backup config are all declarative.
- **How:** `internal/controller/etcd/reconciler.go` runs PreSync → Sync → Cleanup over the component operators registered in `internal/component/registry.go` (statefulset, configmap, clientservice, peerservice, serviceaccount, role, rolebinding, memberlease, snapshotlease, poddistruptionbudget). Each component implements `Operator{GetExistingResourceNames, PreSync, Sync, TriggerDelete}`. The `Etcd` resource has a `/scale` subresource so `kubectl scale` works.
- **Pain relieved:** Bringing up a TLS-enabled, backup-configured, multi-node etcd from scratch normally takes a long checklist of manifests; here it's one CR.

### 1.2 Single-node ⇄ multi-node scale-out (1 → N)  ✨
- **What:** Patch `spec.replicas` from 1 to 3 (or 5) and druid converts a single-node cluster into a quorate multi-node one, on the fly, including enabling peer TLS if it wasn't already on.
- **How:** Documented in DEP-03 (`docs/proposals/03-scaling-up-an-etcd-cluster.md`). Druid sets the `gardener.cloud/scaled-to-multi-node` annotation on the StatefulSet; the backup-restore sidecar in each new pod detects it, adds itself to the cluster as a **learner** via `pkg/member/member_control.go` (`AddMemberAsLearner`), waits for sync, then calls `PromoteMember`. Peer-URL TLS state is tracked via the `member.etcd.gardener.cloud/tls-enabled` annotation on member Leases — druid patches the ConfigMap and StatefulSet to enable TLS before scale-out begins, rolling the existing single pod first.
- **Pain relieved:** Going from single-node (cheap, low-latency, fine for dev/test) to HA without re-creating the cluster or losing data.

### 1.3 Hibernation (scale to 0) and resume  ✓
- **What:** Scale `spec.replicas` to 0 to "hibernate" — pods stop, PVCs and backups are retained, no compute cost. Scale back up to resume.
- **How:** Update validation only blocks scale-down to non-zero (`self==0 ? true : self < oldSelf ? false : true` CEL). Druid keeps PVCs because of the StatefulSet `volumeClaimTemplate` model.
- **Pain relieved:** Saving cost during off-hours / for paused dev clusters without re-bootstrapping.

### 1.4 Vertical scaling via `/scale` subresource + VPA integration  ✓
- **What:** Pod resource requests/limits can be patched on the `Etcd` resource or driven by an external VPA. PDB protects the cluster during rollouts.
- **How:** `/scale` subresource is wired on the Etcd CRD. PDB component (`internal/component/poddistruptionbudget`) creates a `minAvailable=quorum` PDB for multi-node, `minAvailable=0` for single-node.

### 1.5 Quorum-aware pod updates via OnDelete strategy (DEP-07)  ✨
- **What:** Instead of letting the StatefulSet controller roll pods highest-ordinal-first (which can take down a healthy member while an already-unhealthy one waits), druid takes the StatefulSet to `OnDelete` strategy and chooses the deletion order based on member health and role.
- **How:** DEP-07 (`docs/proposals/07-quorum-aware-pod-updates.md`). Dedicated controller deletes unhealthy/non-participating pods first; updates followers before the leader to avoid leader elections. Health and role come from member Leases.
- **Pain relieved:** Transient quorum loss during routine image bumps or config changes — a real pain at scale.

### 1.6 In-place etcd version upgrade (alpha)  ✨
- **What:** `UpgradeEtcdVersion` feature gate triggers an automated in-place upgrade to etcd 3.5.27, with a guaranteed pre-upgrade full snapshot.
- **How:** Feature gate documented in `docs/deployment/feature-gates.md`. Alpha, default off, since druid v0.36.
- **Pain relieved:** Most teams treat etcd version bumps as scary one-off projects.

### 1.7 Resource protection webhook  ✨
- **What:** A validating webhook (`Etcd Components Webhook`) blocks anyone from manually editing or deleting the StatefulSet / ConfigMap / Services / Role / RoleBinding / PDB / Leases that back an `Etcd` resource. Only druid itself, the etcd member's own ServiceAccount (for Lease renewal), and explicitly exempted ServiceAccounts (e.g. VPA) can mutate them.
- **How:** `internal/webhook/etcdcomponents`. Per-resource decision uses managed-by/part-of labels and the reconciler's ServiceAccount. Can be disabled per-cluster with `druid.gardener.cloud/disable-etcd-component-protection`.
- **Pain relieved:** The "I just edited the StatefulSet to debug something and now my etcd is gone" class of incident.

### 1.8 Spec auto-reconcile vs gated reconcile  ✓
- **What:** `--enable-etcd-spec-auto-reconcile` controls whether spec changes apply immediately or wait for the `gardener.cloud/operation: reconcile` annotation. Spec reconcile can also be suspended via `druid.gardener.cloud/suspend-etcd-spec-reconcile`.
- **How:** `internal/controller/etcd/reconciler.go` checks these gates in PreSync.
- **Pain relieved:** Letting operators batch risky changes into a maintenance window.

### 1.9 Auto-deletion (finalizer-protected)  ✓
- **What:** Deleting the `Etcd` resource cleans up all managed sub-resources via finalizer-driven deletion.
- **How:** `internal/controller/etcd/reconcile_delete.go`, each component's `TriggerDelete`. PVCs intentionally retained unless explicitly removed.

---

## 2. Backup & restore

### 2.1 Scheduled full + delta snapshots  ✓
- **What:** Cron-scheduled full snapshots; threshold-based (memory or interval) delta snapshots between full snapshots.
- **How:** `etcd-backup-restore`'s `snapshot` / `server` commands. `pkg/snapshot/snapshotter/snapshotter.go`. Driven by `Etcd.spec.backup.fullSnapshotSchedule`, `deltaSnapshotPeriod`, `deltaSnapshotMemoryLimit`. Only the leading sidecar takes snapshots (see 2.10).
- **Pain relieved:** No bespoke cronjobs or operators around `etcdctl snapshot save`.

### 2.2 Multi-cloud backup providers  ✨
- **What:** AWS S3, Azure Blob, GCS, OpenStack Swift, Alicloud OSS, Dell EMC ECS, OpenShift OCS, S3-compatible (e.g. STACKIT), local filesystem.
- **How:** `pkg/snapstore/*` in etcd-backup-restore (one file per provider). Druid plumbs provider through `internal/store/store.go`. Credentials come via Secret refs in `Etcd.spec.backup.store.secretRef`.
- **Pain relieved:** Many operators only support one or two clouds.

### 2.3 Immutable (WORM) backups (DEP-06)  🔥
- **What:** When backups land in a bucket with WORM retention configured (GCS / Azure Blob / S3 / Alicloud OSS), etcd-backup-restore respects the immutability period and won't try to delete or mutate locked objects — and skips immutable snapshots from garbage collection so it doesn't fail.
- **How:** DEP-06 (`docs/proposals/06-immutable-etcd-backups.md`); `docs/usage/enabling_immutable_snapshots.md`. GC logic in `pkg/snapshot/snapshotter/garbagecollector.go` consults each snapshot's immutability expiry. Swift is unsupported (no provider immutability).
- **Pain relieved:** Ransomware / insider-deletion of backups. Compliance regimes (SOC2, financial) that mandate immutable backups.

### 2.4 Dual-site backup sync  🔥
- **What:** Periodically replicates the entire snapshot stream (full + delta) from a primary bucket to a secondary bucket — possibly on a different provider or region — for cross-region DR or multi-cloud redundancy.
- **How:** `pkg/snapshot/copier/` in etcd-backup-restore, driven by `--secondary-backup-sync-enabled` and `--secondary-backup-sync-period` flags. Runs as a background goroutine in the leading sidecar.  See `docs/usage/backup_sync_dual_site.md`.
- **Pain relieved:** "We back up to S3 — but what if AWS us-east-1 is gone?" Most teams have to bolt on a Lambda or rclone cron.

### 2.5 Snapshot compaction job (DEP-02)  🔥
- **What:** When the delta-event count crosses a threshold, druid spawns a separate `Job` that restores the latest full + deltas into an embedded etcd, compacts and defragments it, and writes a new compacted full snapshot back to the store. Restoration from compacted snapshots takes seconds vs minutes.
- **How:** Compaction controller in `internal/controller/compaction/reconciler.go`, only registered when `--enable-backup-compaction=true`. Triggered when `etcd-events-threshold` (default 1M) is exceeded. Runs the `etcdbrctl compact` subcommand (`cmd/compact.go`). Active deadline default 3h. Per-job metrics exposed.
- **Pain relieved:** Restore time blowing up because of huge delta-snapshot chains. Also acts as **continuous backup integrity validation** — if compaction fails, your backups are broken; you find out before disaster.

### 2.6 Backup-aware readiness probe  ✨
- **What:** Backup-restore exposes `/healthz`; if backups are stalling, the sidecar can be wired so the pod becomes unready, which gates client traffic and bounds the worst-case data loss.
- **How:** `pkg/server/httpAPI.go: serveHealthz`. Used by Gardener in production.

### 2.7 Compression policy (gzip/lzw/zlib) for snapshots  ✓
- **How:** `Etcd.spec.backup.compression`. Pluggable via `pkg/compressor/`.

### 2.8 Two GC policies (Exponential, LimitBased) for old snapshots  ✓
- **What:** Exponential keeps last hour fully, then hourly for 24h, then daily for a week, then weekly for a month. LimitBased keeps N most recent full snapshots. Both always retain the latest full + its deltas.
- **How:** `pkg/snapshot/snapshotter/garbagecollector.go`. Configurable via `Etcd.spec.backup.garbageCollectionPolicy` / `maxBackupsLimitBasedGC` / `garbageCollectionPeriod` / `deltaSnapshotRetentionPeriod`.

### 2.9 On-demand snapshot via EtcdOpsTask  ✨
- **What:** Create an `EtcdOpsTask` CR with `config.onDemandSnapshot.type=full|delta` (optional `isFinal`) and druid drives a one-off snapshot. Task moves Pending → InProgress → Succeeded/Failed/Rejected.
- **How:** `internal/controller/etcdopstask/handler/ondemandsnapshot/`. Hits the leading sidecar's `/snapshot/full` or `/snapshot/delta` HTTP endpoint. `isFinal=true` is used in pre-decommission flows (DR cutover).
- **Pain relieved:** Most teams write ad-hoc shell scripts that exec into the pod and call `etcdctl snapshot save`.

### 2.10 Leader-only snapshotting & defragmentation  ✓
- **What:** In a multi-node cluster, exactly one sidecar takes backups and triggers defrag, avoiding bucket-write storms and inconsistent backups.
- **How:** `pkg/leaderelection/leaderelection.go`. The sidecar polls its colocated etcd's `MaintenanceClient.Status` and considers itself "Leader" when its etcd is the raft leader. State machine: Follower → Leader → Unknown.

### 2.11 Cross-bucket backup copy (EtcdCopyBackupsTask)  ✨
- **What:** Standalone `EtcdCopyBackupsTask` CR copies backups from a source store to a target store — with optional age filter, count cap, and "wait for final snapshot" handshake.
- **How:** `internal/controller/etcdcopybackupstask/reconciler.go` provisions a Pod that runs `etcdbrctl copy` (`cmd/copy.go`, `pkg/snapshot/copier/`).
- **Pain relieved:** Migrating an etcd cluster's history between clouds; on-demand DR drills.

---

## 3. Failure recovery

### 3.1 Automatic data-directory validation and restore (single-node)  ✨
- **What:** On every pod start, the backup-restore sidecar validates the etcd data directory's structure, content (corruption check), and revision (compared against latest snapshot). If invalid, it auto-restores from the latest full + delta snapshots before letting etcd start.
- **How:** `pkg/initializer/validator/` driven from `etcdbrctl initialize` (or the embedded init in `server` mode). `etcd-wrapper` calls the sidecar's `/initialization/start` and polls `/initialization/status`; etcd only starts once status = Successful. Logic per `docs/proposals/validation.md` and `restoration.md`.
- **Pain relieved:** "PV came back corrupt after node failure" — operator-free recovery.

### 3.2 Multi-node automatic recovery from transient quorum loss  ✓
- **What:** Etcd-native: as long as quorum can be re-established, the cluster heals itself once pods/network return. Druid just runs the StatefulSet to keep pods rescheduled.

### 3.3 Manual recovery from permanent quorum loss (today) + planned automation  ✨ (roadmap)
- **What:** A documented step-by-step recovery (`docs/usage/recovering-etcd-clusters.md`) using `druid.gardener.cloud/suspend-etcd-spec-reconcile`. Plans (DEP-05) call for a `QuorumRecovery` EtcdOpsTask type.
- **How:** Today: suspend → scale to 0 → delete PVCs → scale up; backup-restore restores from snapshot. Future: a state-machine `EtcdOpsTask` handler that automates this.
- **Pain relieved:** Permanent quorum loss is one of the worst on-calls in the business — runbooks help, automation will be transformative.

### 3.4 Member replacement / re-join  ✓
- **What:** A pod whose data is gone (or whose member-ID has been removed from the cluster) is removed from the etcd member list and re-added as a learner, then promoted once caught up.
- **How:** `pkg/member/member_control.go`: `RemoveMember`, `AddMemberAsLearner`, `PromoteMember`. `IsMemberInCluster` / `WasMemberInCluster` detect the case. Retried via `AddLearnerWithRetry`.

### 3.5 ClusterID mismatch detection  ✨
- **What:** A separate condition (`ClusterIDMismatch`) flags the case where a member somehow joined a different cluster (e.g. accidentally restored from a foreign backup) — visible as an `Etcd` printer column.
- **How:** `internal/health/condition/check_cluster_id_mismatch.go`.
- **Pain relieved:** A subtle, hard-to-diagnose class of restore failures.

### 3.6 Revision-based stale-data protection  ✨
- **What:** Backup-restore refuses to start an etcd whose local revision is behind the latest snapshot's revision — preventing a "stale" member from overwriting good backups with old data.
- **How:** Revision check in validation flow (`docs/proposals/validation.md`).

---

## 4. EtcdOpsTask — out-of-band operator tasks (DEP-05)

Unified CRD (`druid.gardener.cloud/v1alpha1.EtcdOpsTask`, shortname `eot`) for ad-hoc operations. State machine: `Pending → InProgress → Succeeded | Failed | Rejected`, with three sub-phases (Admit → Execute → Cleanup), TTL-based GC, per-run `runID`, `LastOperation`, `LastErrors` with error codes. Spec is immutable (CEL: `self == oldSelf`).

### 4.1 On-demand full or delta snapshot (shipped)  ✨
- See 2.9. Includes `isFinal` flag for cutover scenarios.

### 4.2 On-demand maintenance — compaction + defragmentation (designed)  ✨ (roadmap)
- DEP-05 + DEP-02 envisage triggering compaction or rolling defrag as an EtcdOpsTask.

### 4.3 Quorum recovery (designed)  🔥 (roadmap)
- A single CR that drives the full permanent-quorum-loss recovery workflow.

### 4.4 Backup copy as task (orthogonal: `EtcdCopyBackupsTask`)
- See 2.11.

### 4.5 Pluggable handler registry  ✨
- **How:** `internal/controller/etcdopstask/handler/registry.go` exposes `TaskHandlerRegistry.Register(taskType, factory)`. New task types are added without changing the controller — only a new handler. Each handler implements Admit / Execute / Cleanup.
- **Pain relieved:** This is the platform on which any future day-2 automation gets bolted on.

---

## 5. Observability

### 5.1 Rich Etcd CR printer columns and conditions  ✓
- **Printer columns:** `Ready`, `Quorate`, `AllMembersReady`, `BackupReady`, `ClusterIDMismatch`, `Cluster Size`, `Current Replicas`, `Ready Replicas`. A single `kubectl get etcd` tells you cluster state.
- **Conditions:** `Ready`, `AllMembersReady`, `AllMembersUpdated`, `BackupReady`, `DataVolumesReady`, `ClusterIDMismatch`. Implementations in `internal/health/condition/check_*.go`.

### 5.2 Per-member status via Leases (today) → EtcdMember CR (DEP-04)  ✨
- **Today:** Each member has a Lease object that the backup-restore sidecar renews; the Lease annotations carry member-ID, role (Leader/Member), TLS state. Druid uses these to know per-member liveness.
- **Roadmap:** DEP-04 (`04-etcd-member-custom-resource.md`) introduces an `EtcdMember` CR with much richer per-member state — designed for future per-member day-2 actions.

### 5.3 Prometheus metrics — druid  ✓
- **What:** `etcddruid_compaction_jobs_total{succeeded}`, `etcddruid_compaction_jobs_current{etcd_namespace}`, `etcddruid_compaction_job_duration_seconds`, `etcddruid_compaction_num_delta_events`. Standard controller-runtime workqueue/runtime metrics. (`internal/controller/compaction/metrics.go`, `internal/metrics/metrics.go`.)

### 5.4 Prometheus metrics — backup-restore  ✓
- **What:** `etcdbr_snapshot_duration_seconds`, `etcdbr_snapshot_gc_total`, `etcdbr_snapshot_latest_revision`, `etcdbr_snapshot_latest_timestamp`, `etcdbr_snapshot_required`, `etcdbr_defragmentation_duration_seconds`, `etcdbr_validation_duration_seconds`, `etcdbr_restoration_duration_seconds`, `etcdbr_snapstore_latest_deltas_total`, plus passthrough `etcd_*` metrics from the embedded etcd. (`docs/operations/metrics.md`.)

### 5.5 `lastOperation` / `lastErrors` on every CR  ✨
- **What:** `Etcd.status.lastOperation` + structured error codes give a clear, machine-readable trail of the latest reconciliation, with a `runID` to correlate logs. Same on `EtcdOpsTask`.

### 5.6 pprof endpoints on backup-restore  ✓
- **What:** `/debug/pprof/*` exposed by the sidecar HTTP server for live profiling.

---

## 6. Multi-tenancy / scale

### 6.1 Cluster-shared operator running thousands of etcd clusters  🔥
- **What:** One druid deployment manages every etcd cluster in the (Kubernetes) cluster it runs in. SAP's Gardener runs **tens of thousands of shoot control-plane etcd clusters** off this stack across production seed clusters.
- **How:** Workqueue with `--etcd-workers`, `--compaction-workers`, `--etcd-copy-backups-task-workers`, `--secret-workers` etc. Per-resource exponential-backoff rate limiter on EtcdOpsTask. Status sync is decoupled from spec reconcile.
- **Pain relieved:** You don't deploy an operator per cluster. This is the largest documented production deployment of any etcd operator.

### 6.2 HA druid via leader election  ✓
- **How:** `--enable-leader-election` + Lease-based lock.

### 6.3 Per-Etcd ServiceAccount + Role + RoleBinding  ✓
- **What:** Each etcd's backup-restore sidecar runs as a dedicated SA scoped to just the Leases / StatefulSet / pods of its own cluster. No cross-cluster blast radius.

### 6.4 Caching tuning for scale  ✓
- **How:** `--disable-lease-cache` flag, `--etcd-status-sync-period` (default 15s), `--etcd-member-notready-threshold` (5m), `--etcd-member-unknown-threshold` (1m).

---

## 7. Security

### 7.1 TLS everywhere  ✓
- **What:** Client-server TLS (etcd↔kube-apiserver), peer-peer TLS (etcd↔etcd), wrapper↔etcd, wrapper↔backup-restore, backup-restore↔etcd, backup-restore↔wrapper. Documented in `docs/usage/securing-etcd-clusters.md`.
- **How:** `spec.etcd.clientUrlTls`, `spec.etcd.peerUrlTls`, `spec.backup.tls`. Druid auto-propagates TLS state through ConfigMap + StatefulSet + Lease annotation.

### 7.2 Automatic peer-TLS enablement during scale-out  ✨
- See 1.2. Druid notices peer TLS is off when scaling 1→N and turns it on first.

### 7.3 Secret reconciler with finalizers  ✓
- **What:** `internal/controller/secret/` puts a finalizer on any Secret referenced by an `Etcd` so it cannot be deleted while in use; protects backup credentials and TLS certs.

### 7.4 CRD validation via OpenAPI + CEL  ✓
- **What:** Strong schema validation: cron-expression regex, duration regex, enum constraints, immutability of `storageClass` / `storageCapacity` / `volumeClaimTemplate`, "garbageCollectionPeriod > deltaSnapshotPeriod" CEL rule, "storageCapacity > 3×quota when backups enabled" CEL rule.
- **How:** Documented in `docs/usage/validating-etcd-clusters.md`.

### 7.5 Resource-protection webhook (anti-tamper)  ✨
- See 1.7.

---

## 8. Upgrades & migration

### 8.1 Druid version compatibility matrix  ✓
- `docs/deployment/version-compatibility-matrix.md` pins compatible etcd / backup-restore / wrapper versions.

### 8.2 OperatorConfiguration (versioned config)  ✓
- **What:** New recommended deployment uses a versioned config CR (`druidconfigv1alpha1.OperatorConfiguration`) via `--config`. Old per-flag CLI is deprecated but supported.

### 8.3 In-place etcd version upgrade (feature gate)  ✨
- See 1.6.

### 8.4 Snapshot-format compatibility / directory layout v1→v2  ✓
- Backup directory layout was migrated to flat v2 for compaction (DEP-02); restoration reads both for backward compatibility.

---

## 9. Self-healing (automatic, no human)

| Failure | What happens automatically |
|--------|----------------------------|
| Pod crash | StatefulSet reschedules; sidecar validates data dir; restores from snapshot if corrupt; etcd resumes. |
| PVC corruption | Validator detects → restorer pulls latest full + deltas → etcd starts clean. (single-node) |
| Member data loss in multi-node | Sidecar detects member is no longer in cluster, removes stale member-ID, re-adds as learner, promotes when synced. |
| Backup pod loses leadership | New leader elected; new sidecar starts taking snapshots; old one steps down. |
| Snapshot bucket transient error | Snapshotter retries; `etcdbr_snapshot_required` metric raises; readiness flips off until recovery. |
| Etcd member becomes stale | Revision check blocks startup, forces restore from snapshot. |
| Delta chain too long | Compaction job auto-spawned at event threshold; produces fresh compacted full snapshot. |
| Pod template change | (Future / DEP-07) Quorum-aware OnDelete picks safe ordering. |
| Transient quorum loss | Etcd native; druid keeps pods alive. |
| Permanent quorum loss | Operator-driven today (runbook); EtcdOpsTask automation planned. |
| Manual mutation of managed K8s resource | Webhook rejects it. |
| Spec change during incident | Operator can `suspend-etcd-spec-reconcile` to freeze druid. |

---

## 10. CLI / UX

### 10.1 `druidctl` (planned)  ✨
- DEP-05 explicitly calls out a CLI to drive EtcdOpsTasks. Not in master yet, but designed.

### 10.2 `etcdbrctl` subcommands  ✓
- `snapshot`, `restore`, `initialize`, `compact`, `copy`, `server` (`etcd-backup-restore/cmd/`). All cobra-based, work standalone outside Kubernetes.

### 10.3 Backup-restore HTTP API  ✓
- Endpoints: `/initialization/start`, `/initialization/status`, `/snapshot/full`, `/snapshot/delta`, `/snapshot/latest`, `/config`, `/healthz`, `/metrics`, `/debug/pprof/*`. (`pkg/server/httpAPI.go`.) Leader delegates non-leader-served calls automatically (`delegateReqToLeader`).

### 10.4 Helm chart + Kind-based local setup  ✓
- `chart/` in both repos; `make kind-up` / `make deploy` / `make test-e2e`.

---

## 11. Extensibility / ecosystem

### 11.1 Donated to LF/CNCF NeoNephos sub-foundation  ✨
- etcd-druid is part of the **Gardener** project, which moved to the Linux Foundation under the NeoNephos technical charter. This signals vendor-neutral governance — a non-trivial differentiator vs single-vendor operators.

### 11.2 Pluggable storage providers  ✨
- New providers can be added by implementing the `SnapStore` interface (`pkg/snapstore/snapstore.go`). Documented in `docs/development/new_cp_support.md`.

### 11.3 Pluggable EtcdOpsTask handlers  ✨
- See 4.5.

### 11.4 Three coordinated CRDs  ✓
- `Etcd`, `EtcdOpsTask`, `EtcdCopyBackupsTask` form a cohesive API surface.

### 11.5 Three coordinated repos  ✓
- `etcd-druid` (operator) + `etcd-backup-restore` (sidecar) + `etcd-wrapper` (lifecycle wrapper). Each can be released independently; matrix documented.

### 11.6 First-class component-operator pattern  ✨
- Every K8s resource is its own `Operator{GetExistingResourceNames, PreSync, Sync, TriggerDelete}` — a clean, testable internal API. Adding a new managed resource = one new component. `internal/component/registry.go`.

---

## 12. Other notable details

- **Active deadline** on compaction jobs (default 3h) prevents runaway jobs.
- **Failure-reason labels** on compaction metrics: `preempted | evicted | deadlineExceeded | processFailure | unknown | none` — lets you SLO compaction failures distinct from pod-eviction noise.
- **High-watch-event-rate hardening** in delta snapshotter: backup-restore uses parallel goroutines to drain the etcd watch channel to avoid sidecar OOM under burst load (DEP `high_watch_event_ingress_rate.md`).
- **Additional advertise peer URLs** for advanced topologies (`spec.etcd.additionalAdvertisePeerURLs`) with CEL-validated naming.
- **Quota / storage-capacity coupling validation** — refuses to create a cluster with too little PV vs etcd `--quota-backend-bytes`.
- **Compression-on-the-wire to object store** (gzip/lzw/zlib) — saves egress cost at scale.

---

## Top 10 most differentiating capabilities (lean on these in the talk)

| # | Capability | Why it's hard for others to match |
|---|-----------|----------------------------------|
| 1 | **Running 10,000s of etcd clusters in production at SAP/Gardener** (capability 6.1) | The only documented planet-scale, multi-tenant etcd-operator deployment. Pattern-validated, not theoretical. |
| 2 | **Immutable (WORM) backups across GCS / Azure / S3 / OSS** (2.3) | Real ransomware/compliance answer. No other open-source etcd operator ships this. |
| 3 | **Dual-site backup sync across providers** (2.4) | Built-in cross-cloud DR for etcd backups; everyone else writes glue. |
| 4 | **Snapshot compaction job that doubles as continuous backup validation** (2.5) | Restoring is the only true validation — druid does it every day at scale and exposes failures via metrics. |
| 5 | **Single-node → multi-node live scale-out with auto peer-TLS upgrade** (1.2 + 7.2) | Cost-efficient start, HA when you need it, no re-bootstrap. |
| 6 | **Unified EtcdOpsTask CRD + pluggable handler registry** (Section 4) | The day-2 operations platform — extends without changing the controller. |
| 7 | **Quorum-aware pod updates via OnDelete strategy** (1.5) | Eliminates a real class of avoidable transient downtime during routine upgrades. |
| 8 | **Resource-protection validating webhook** (1.7) | "Hands off the StatefulSet" guardrail tied to ServiceAccount identity — surprisingly few operators do this. |
| 9 | **Vendor-neutral LF NeoNephos governance** (11.1) | A real, defensible "this is not a vendor lock-in" story. |
| 10 | **Automatic data-directory validation + restore on every pod start** (3.1) | The corner that makes the single-node story palatable — and the model that scales to multi-node member recovery (3.4). |
