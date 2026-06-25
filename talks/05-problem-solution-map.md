# Problem → Solution Map: The Spine of the Talk

> Maps the Top 10 pain points from `02-pain-points.md` onto the capability inventory in `03-capabilities.md`.
> Ordering is **narrative**, not frequency: we open with the pain everyone in the room has felt, then climb into the sophisticated stuff, then close with the "wow, you do that?" capabilities.
> Honesty matters — open gaps are listed at the bottom rather than hidden.

---

## Differentiation Matrix (the one slide that does a lot of work)

| Pain | Manual ops (kubeadm, scripts) | etcd-operator (CoreOS/Zalando, deprecated) | Cloud-managed etcd (none mainstream; closest = managed K8s control planes) | **etcd-druid + etcd-backup-restore** |
|---|:--:|:--:|:--:|:--:|
| Declarative provisioning of a TLS-secured, backup-configured cluster | ❌ checklist of YAML | ⚠️ basic CRD, no backup integration | ✅ hidden from user | ✅ one `Etcd` CR |
| Scheduled full + delta snapshots, multi-cloud | ❌ DIY cron + scripts | ❌ S3-only, no delta | ✅ hidden | ✅ AWS / Azure / GCS / Swift / OSS / ECS / OCS / local |
| Immutable (WORM) backups | ❌ | ❌ | ⚠️ if provider supports | ✅ DEP-06, GCS/Azure/S3/OSS |
| Cross-region / cross-cloud backup sync | ❌ rclone glue | ❌ | ⚠️ region-pinned | ✅ dual-site backup sync |
| Automatic restore on PVC corruption / unclean shutdown | ❌ paging on-call | ⚠️ partial | ✅ hidden | ✅ validator + auto-restore on every pod start |
| Compaction *and* defrag — safely, cluster-aware | ❌ manual, dangerous on leader | ❌ | ✅ hidden | ✅ leader-only, member-by-member, separate compaction Job |
| Quota / NOSPACE prevention | ❌ "you should have monitored" | ❌ | ✅ hidden | ✅ snapshot compaction + defrag prevents runaway DB growth |
| Quorum loss recovery | ❌ Platform9-style ticket | ❌ | ✅ hidden | ✅ runbook today, `EtcdOpsTask` automation in DEP-05 |
| Single-node ⇄ multi-node live scale-out | ❌ rebuild from scratch | ❌ | ❌ fixed topology | ✅ live, with auto peer-TLS upgrade |
| Quorum-aware rolling updates | ⚠️ StatefulSet ordinal-based, dumb | ❌ | ✅ hidden | ✅ OnDelete strategy (DEP-07) |
| Cert lifecycle (peer + client mTLS) | ⚠️ kubeadm yearly outage | ❌ | ✅ hidden | ✅ secret-driven, restart orchestration |
| Resource-tamper protection ("I edited the StatefulSet…") | ❌ | ❌ | ✅ hidden | ✅ validating webhook |
| Running 10,000s of etcd clusters from one operator | ❌ | ⚠️ untested at scale | ✅ vendor-specific | ✅ Gardener production scale |
| Vendor-neutral governance | n/a | ❌ archived 2020 | ❌ vendor-locked | ✅ LF / NeoNephos |

Legend: ✅ first-class, ⚠️ partial / manual, ❌ not supported.

The deprecated CoreOS / Zalando etcd-operator column is intentional — many in the audience will have tried it years ago and given up. Saying "that one is dead, here's what replaced it" anchors druid as the surviving open-source answer.

---

## The Top 10 Problem → Solution Mapping (narrative order)

### 1. "etcd is a black box I inherited with Kubernetes" (Pain #1, #32)

- **The pain:** Almost nobody in the room *chose* to run etcd. It arrived with Kubernetes. The mental model is the kubeadm reference architecture, and that's where it ends — until something breaks at 3 AM and Google delivers "mvcc: database space exceeded" with no explanation.
- **How druid/etcdbr solves it:** The whole *point* of an operator is to encode etcd expertise as automation. The `Etcd` CR is a single, opinionated API. Status conditions (`Ready`, `Quorate`, `AllMembersReady`, `BackupReady`, `ClusterIDMismatch`) and printer columns mean `kubectl get etcd` tells you the state of the world without `etcdctl`.
- **Differentiation:** Rolling your own = you become the etcd expert. Zalando's etcd-operator was archived in 2020. Cloud-managed K8s hides etcd entirely — fine, until you need to do something it doesn't expose. Druid lets you stay in control without having to memorise raft.
- **Credibility hook:** SAP runs **tens of thousands** of etcd clusters this way in Gardener production. It is the largest known multi-tenant etcd-operator deployment, period.
- **If you don't have this:** You will spend Friday nights piecing together `etcdctl endpoint status`, `etcd-dump-db`, and pprof traces under pressure, while kubectl is down.

### 2. NOSPACE — your cluster goes read-only and you can't even delete the bloat (Pain #7, #4, #6)

- **The pain:** Hit the 2 GB default quota and etcd flips read-only. kube-apiserver writes fail. You can't delete the orphaned secrets that filled the disk *because the API needed to delete them is down*. The kops post-mortem of "2.2 GB of logging-operator secrets" is the canonical story.
- **How druid/etcdbr solves it:** Three layers.
  1. CEL validation refuses clusters with `storageCapacity < 3 × quota` — sized to survive growth + compaction.
  2. etcd-backup-restore takes scheduled snapshots, drives **automatic defragmentation** member-by-member (leader-only, never stop-the-world on the leader).
  3. The **snapshot compaction Job** (DEP-02) periodically rebuilds the snapshot stream by restoring into an embedded etcd, compacting, defragging, and writing back — which keeps the live DB lean *and* validates backups by exercising the restore path.
- **Differentiation:** Manual ops = `etcdctl defrag` on a Friday and pray. Zalando operator = nothing. Cloud-managed = invisible to you, fine until you outgrow their quota.
- **Credibility hook:** Compaction has run in production at Gardener scale for years; failure-reason metrics (`preempted | evicted | deadlineExceeded | processFailure`) let you SLO it.
- **If you don't have this:** Your cluster freezes, your customers can't deploy, and your first remediation step is editing `quota-backend-bytes` on a live cluster while disarming NOSPACE alarms.

### 3. "We have backups… right?" (Pain #11, #21, #22)

- **The pain:** Tutorials end at `etcdctl snapshot save` on a cron. Nobody verifies them. Nobody runs restore drills. The first time you discover the backup is broken is the moment you need it. Upstream etcd literally has an open issue (#21283) asking for built-in integrity testing — it doesn't exist.
- **How druid/etcdbr solves it:**
  - Scheduled **full + delta** snapshots with the leader-only sidecar (no bucket-write storms).
  - **Multi-cloud:** AWS S3, Azure Blob, GCS, OpenStack Swift, Alicloud OSS, Dell ECS, OpenShift OCS, S3-compatible (MinIO/STACKIT), local FS. One config, eight backends.
  - Compression (gzip/lzw/zlib), two GC policies (Exponential / LimitBased).
  - **Compaction = continuous restore validation.** Every compaction Job restores backups into an embedded etcd. If they're broken, you find out *before* disaster.
  - On-demand snapshots via `EtcdOpsTask` (`isFinal=true` for DR cutover).
- **Differentiation:** Velero backs up Kubernetes objects, not etcd's raft state. Zalando operator was S3-only with no delta or compaction. Cloud-managed hides it — fine until you need a backup *out* of their cloud.
- **Credibility hook:** Per-job duration / event-count / failure metrics are exposed; Gardener runs this across every shoot control plane.
- **If you don't have this:** You discover at minute 47 of an outage that your snapshot file is hashless, partial, or restore-incompatible across minor versions (etcd #18340, #20967, #20009).

### 4. Defrag is a stop-the-world hand grenade (Pain #5)

- **The pain:** `etcdctl defrag` locks the entire database. Run it on the leader of a busy cluster and you get cascading 500s, 5-second write latencies, and rejected requests. Upstream etcd has no built-in safe defrag orchestration.
- **How druid/etcdbr solves it:** Backup-restore's defrag flow is **member-by-member, follower-first, leader-last**, with raft-catch-up wait between members. Only the leading sidecar drives it. Lease-based leader election (`pkg/leaderelection`) ensures exactly one node is in charge.
- **Differentiation:** Manual = guaranteed latency spike. No other open-source operator orchestrates cluster-aware defrag this way.
- **Credibility hook:** `etcdbr_defragmentation_duration_seconds` metric; production-tested at scale.
- **If you don't have this:** You either skip defrag (and hit NOSPACE), or you defrag the leader at 2 AM and explain the 5-second p99 spike at standup.

### 5. Certs expire and everything dies at the same instant (Pain #8, #29)

- **The pain:** Peer + client TLS certs and their CAs have hard expiry. They roll past midnight, every etcd member stops talking, kubectl fails everywhere, and you can't fix it because you need the API to fix the API. Worse: etcd cannot hot-reload trusted-CA bundles (issue #11555, open since 2020), so zero-downtime CA rotation is impossible upstream.
- **How druid/etcdbr solves it:** TLS material is declarative on the `Etcd` CR (`spec.etcd.clientUrlTls`, `spec.etcd.peerUrlTls`, `spec.backup.tls`). Druid auto-propagates state through ConfigMap, StatefulSet, and the per-member Lease's `member.etcd.gardener.cloud/tls-enabled` annotation. Secret finalizers prevent referenced TLS Secrets from being deleted while in use. Restart orchestration is quorum-aware (DEP-07 OnDelete) so a roll triggered by cert rotation doesn't take quorum down.
- **Differentiation:** Manual rotation = the yearly kubeadm fire drill. Zalando: none. Cloud-managed: you trust them.
- **Credibility hook:** Gardener rotates certs across thousands of shoot etcd clusters on a regular cadence using this machinery.
- **If you don't have this:** You schedule a planned outage every 12 months, or you forget and have an unplanned one.

### 6. Pod / member operations without taking quorum down (Pain #9, #10)

- **The pain:** Add-then-remove vs remove-then-add ordering, learner promotion, peer-URL updates, persistent WAL state — get the sequence wrong and you drop quorum. Worse, the StatefulSet controller will happily delete a healthy pod while an already-unhealthy one waits, taking the cluster down during a routine image bump.
- **How druid/etcdbr solves it:**
  - **Member replacement:** `pkg/member/member_control.go` does `RemoveMember` → `AddMemberAsLearner` → `PromoteMember` with retry, plus stale-member-ID detection (`IsMemberInCluster` / `WasMemberInCluster`).
  - **Quorum-aware pod updates (DEP-07):** StatefulSet is set to `OnDelete`; a dedicated controller deletes unhealthy pods first and followers before the leader.
  - **Single-node → multi-node live scale-out (DEP-03):** Patch `replicas: 1 → 3`, druid enables peer TLS on the existing pod first, then new members join as learners and get promoted.
- **Differentiation:** Manual = the sequence is in a runbook nobody re-reads under pressure. Zalando: never automated learner promotion. Cloud-managed: topology is fixed.
- **Credibility hook:** DEP-03 and DEP-07 are written-down, peer-reviewed designs, not folklore.
- **If you don't have this:** Every cluster operation is a held breath.

### 7. The PV came back corrupt — now what? (Pain #13, #15)

- **The pain:** A coordinated power loss, an OOM at the wrong moment, an AZ outage — bbolt can't recover its page tracker, you get "freepages: failed to get all reachable pages" or "page expected to be X but self identifies as 0", and etcd will not start. Only fix: restore from backup. And even `etcdutl snapshot restore` has its own bugs at scale (#16052 OOM, #20967 ghost members).
- **How druid/etcdbr solves it:** Every pod start runs through the backup-restore **validator + restorer** (`pkg/initializer/validator/`). It checks data-directory structure, content (corruption), and revision. If invalid, it auto-restores from the latest full + deltas *before* etcd is allowed to start. `etcd-wrapper` polls `/initialization/status` and gates the etcd process on success. Revision-based stale-data protection refuses to start a member whose local revision is behind the latest snapshot — preventing a stale node from overwriting good state.
- **Differentiation:** Manual = you on-call. Zalando: nothing comparable. Cloud-managed: hidden.
- **Credibility hook:** This is the single capability that makes the single-node etcd story palatable for Gardener (where each shoot starts as a single-node etcd by default).
- **If you don't have this:** Every PV-level glitch becomes a page and a restore-from-snapshot ritual you may or may not have practised.

### 8. Permanent quorum loss — the worst day of the year (Pain #12)

- **The pain:** Lose more than (N−1)/2 members permanently and the cluster is unrecoverable from inside etcd. The recovery procedure — pick the member with the latest revision, `etcdutl snapshot restore`, `--force-new-cluster`, re-add members, re-issue certs — is long, manual, typo-prone. Platform9's KB literally says "open a support ticket." `--force-new-cluster` itself causes cluster-ID mismatches if misused (#8169).
- **How druid/etcdbr solves it:**
  - **Today:** Written, supported runbook at `docs/usage/recovering-etcd-clusters.md` using `druid.gardener.cloud/suspend-etcd-spec-reconcile` to freeze druid, scale to 0, drop PVCs, scale back up, and let the validator restore from the latest snapshot.
  - **Roadmap:** DEP-05 designs a `QuorumRecovery` `EtcdOpsTask` type that automates the whole sequence as a single CR.
  - **Defence in depth:** `ClusterIDMismatch` condition (`internal/health/condition/check_cluster_id_mismatch.go`) is a printer column, so an accidentally-restored-from-foreign-backup member is flagged immediately.
- **Differentiation:** Manual = pages of runbook, prayer. Zalando: none. Cloud-managed: you call your provider.
- **Credibility hook:** This sequence is exercised at Gardener scale; the controller for the automated version is being staged behind the `EtcdOpsTask` framework that already shipped on-demand snapshots.
- **If you don't have this:** This is the kind of outage that ends up in a CNCF blog post with your company's name in it.

### 9. Diagnostics — why is the cluster unhealthy? (Pain #17, #26, #27, #28)

- **The pain:** Operators see symptoms (slow kubectl, pods pending, controllers lagging) and have to guess: disk, network, compaction, leader churn, quota? Error messages are useless — "apply request took too long", "context deadline exceeded". And nobody monitors fsync p99, which is the single most important etcd signal. etcd's own issue #20217 ("Implement a comprehensive diagnosis tool") openly admits the upstream tooling is a scatter.
- **How druid/etcdbr solves it:**
  - **Conditions on the CR:** `Ready`, `AllMembersReady`, `AllMembersUpdated`, `BackupReady`, `DataVolumesReady`, `ClusterIDMismatch` — printer columns make `kubectl get etcd` enough to triage.
  - **`lastOperation` + structured error codes + per-reconcile `runID`** — correlate Kubernetes events with logs without guessing.
  - **Backup-restore metrics:** `etcdbr_snapshot_duration_seconds`, `etcdbr_snapshot_latest_revision`, `etcdbr_defragmentation_duration_seconds`, `etcdbr_validation_duration_seconds`, `etcdbr_restoration_duration_seconds`, `etcdbr_snapstore_latest_deltas_total`, plus passthrough `etcd_*` metrics including disk fsync.
  - **Per-member Leases (today) / `EtcdMember` CR (DEP-04 roadmap)** for per-member visibility.
- **Differentiation:** Manual = "did you check fsync?" by people who've been burned. Zalando: no equivalent surface. Cloud-managed: only what they choose to expose.
- **Credibility hook:** Every condition has a corresponding `check_*.go` file under `internal/health/condition/` — verifiable, not aspirational.
- **If you don't have this:** Every incident has a 30-minute "what is even happening" phase.

### 10. Wow — *that* too? (Capabilities the audience didn't know to ask for)

These are the closing crescendo. The first nine items prove druid handles the daily-pain stories; this one earns the talk slot.

- **Immutable (WORM) backups (DEP-06)** — ransomware and SOC2-style compliance answers. GCS object retention, Azure Blob immutability, S3 Object Lock, Alicloud OSS WORM are all natively respected; the GC code consults each snapshot's immutability expiry. *No other open-source etcd operator ships this.*
- **Dual-site backup sync** — built-in cross-region/cross-cloud replication of the snapshot stream. Drop `--secondary-backup-sync-enabled` and the leading sidecar keeps a second bucket warm. Most teams write a Lambda or rclone cron for this.
- **Resource-protection validating webhook (1.7)** — manually editing the StatefulSet to debug something? Rejected. Only druid, the etcd's own ServiceAccount, and explicitly-exempted SAs (e.g. VPA) can mutate the managed resources. Identity-based, not annotation-based. Surprisingly few operators do this.
- **`EtcdOpsTask` + pluggable handler registry (Section 4)** — the day-2 platform. On-demand snapshot ships today; quorum-recovery, on-demand compaction/defrag, version-upgrade are next, *without changing the controller*.
- **In-place etcd version upgrade** (alpha, feature-gated) — guaranteed pre-upgrade full snapshot, then in-place to etcd 3.5.27. Version bumps stop being scary one-off projects.
- **Vendor-neutral governance (11.1)** — donated under the **Linux Foundation NeoNephos** sub-foundation as part of Gardener. Not a vendor's tool. Important to anyone burned by single-vendor abandonware (cough, CoreOS etcd-operator).

---

## Open challenges (honesty section — CFP reviewers reward this)

The talk is stronger if we name what we *don't* solve yet. None of these are blockers; all of them are interesting roadmap.

- **Automated permanent-quorum-loss recovery** is a runbook today, not a CR. DEP-05 designs `QuorumRecovery` as an `EtcdOpsTask` type but it has not shipped. We say so.
- **Horizontal scale / sharding (Pain #19, the 8 GB ceiling)** — druid does not solve this because etcd itself doesn't. Honest answer: if you have >8 GB of state, you need `--etcd-servers-overrides` to split CRDs, and that is out of scope for any etcd operator. The witness-member proposal upstream (etcd #20696) may help in the future.
- **Tenancy controls (Pain #25)** — per-tenant rate limiting / quotas don't exist in etcd, so druid can't surface them. Druid's contribution is workload **isolation** (one etcd per tenant in Gardener's model), not in-cluster multi-tenancy.
- **`druidctl` CLI (10.1)** is designed, not yet shipped. For now, `kubectl` + `EtcdOpsTask` CRs do the job.
- **Diagnostics tool of etcd itself** — druid surfaces what etcd makes available; the deeper "diagnose why apply request took long" tooling (etcd #20217) is happening upstream, not in druid. We're a consumer of it, not a builder.
- **In-place etcd version upgrade is alpha.** Default off. Production users still do most upgrades via re-deploy + restore-on-init.
- **Sidecar memory under high watch-event burst** — DEP `high_watch_event_ingress_rate.md` hardened it, but very heavy watch workloads can still pressure backup-restore memory. Active work.

---

## Why this ordering works for the talk

1. We open with "etcd is a black box" — every head in the room nods.
2. We dig into NOSPACE — half the room has had this exact outage.
3. Backups, defrag, certs, member ops — increasingly senior pains.
4. PV corruption and quorum loss — the catastrophic, paging-at-3-AM stuff.
5. Diagnostics — the meta-pain.
6. Closing wow factor — immutable backups, dual-site sync, governance.
7. Honest gaps — buys the audience's trust on everything we *did* claim.
