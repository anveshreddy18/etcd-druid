# Operational Pain Points of Running etcd in Production

> Source-grounded catalog of the operational pain points that people running etcd clusters in production actually experience. Drawn from GitHub issues (etcd-io/etcd, kubernetes/kubernetes, gardener/etcd-druid, gardener/etcd-backup-restore), CNCF / vendor blogs, Hacker News, Reddit (where indexed), and post-mortems. Used as the foundation for the Bangalore meetup talk on etcd-druid and etcd-backup-restore.

> Frequency = how often the pain shows up across independent sources (High / Medium / Low).
> Severity = how badly it hurts when it does (Critical / High / Medium / Low).
> Each pain point is followed by 3–5 source URLs and, where available, direct quotes from users venting in the wild.

---

## Day 0 — Provisioning & Setup

### 1. "etcd is a black box that you have to learn before you can even safely *start* one"
- **Frequency:** High
- **Severity:** High
- **Description:** Newcomers run into a wall of unfamiliar concepts (Raft, quorum, WAL, MVCC, snapshots, leases, compaction vs. defragmentation, peer vs. client TLS, learner members, force-new-cluster) before they can produce a cluster they trust. Almost every "troubleshooting" guide first has to teach the data model. The OneUptime and rack2cloud guides essentially exist because operators don't have an intuition for etcd internals when something breaks at 3am.
- **Quotes:**
  - "etcd failures don't present as database failures. They present as Kubernetes acting weird." — rack2cloud
  - "Without deep knowledge of how it works under the hood, reaching the root cause of a control plane failure can be almost impossible." — Caue, kops outage write-up
  - "etcd failures rarely announce themselves clearly." — CNCF blog
- **Sources:**
  - https://www.cncf.io/blog/2026/03/12/making-etcd-incidents-easier-to-debug-in-production-kubernetes/
  - https://www.rack2cloud.com/etcd-kubernetes-database/
  - https://www.kubenatives.com/p/etcd-debugging-kubernetes
  - https://oneuptime.com/blog/post/2026-01-25-troubleshoot-etcd-issues-kubernetes/view

### 2. Bootstrapping a healthy 3-node cluster with correct peer/client TLS
- **Frequency:** Medium
- **Severity:** High
- **Description:** Setting up TLS-secured peer URLs, client URLs, separate CAs for peer vs. client, and getting hostname/SAN/IP-SAN combinations right is a footgun. Mis-set values lead to "tls: bad certificate", "rafthttp: failed to dial", and members that refuse to join the cluster. Issue #16002 in etcd-io/etcd shows that *one* misconfigured new member can bring the entire cluster down.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/16002 — "Incorrect configuration in the new etcd member can bring down the etcd cluster"
  - https://github.com/kubernetes/kubernetes/issues/83028 — "kube-apiserver: failover on multi-member etcd cluster fails certificate check on DNS mismatch"
  - https://github.com/etcd-io/etcd/issues/6311 — "certs and dealing with an outage"

### 3. Sizing: nobody knows what disk / CPU / memory their etcd actually needs
- **Frequency:** High
- **Severity:** Medium
- **Description:** Recommendations are scattered (SSD with low fsync, separate disk from OS, dedicated nodes, 8 GB RAM, etc.). Users find out their disk is too slow only after kube-apiserver starts timing out. Cloud users get burned by "burstable" volumes whose IOPS collapse silently. Comments across multiple guides agree disk fsync is the single most under-monitored signal.
- **Sources:**
  - https://etcd.io/docs/v3.5/op-guide/performance/
  - https://devopskit.tech/en/posts/etcd-slow-disks/
  - https://support.scc.suse.com/s/kb/slow-etcd-performance-testing-and-optimization
  - https://docs.redhat.com/en/documentation/openshift_container_platform/4.20/html/etcd/

### 4. Defaults bite later (2 GB quota, no auto-compaction, no auto-defrag)
- **Frequency:** High
- **Severity:** High
- **Description:** Out-of-the-box etcd will silently grow until it hits the 2 GB quota and freezes writes. Auto-compaction is off by default; defrag is fully manual and disruptive. Almost every NOSPACE post-mortem starts with "we didn't realise we needed to configure these."
- **Sources:**
  - https://devopsbeast.com/blog/etcd-compaction-defrag-guide — "The 2GB quota will bite you"
  - https://etcd.io/blog/2023/how_to_debug_large_db_size_issue/
  - https://www.vcluster.com/docs/vcluster/troubleshoot/etcd-compaction

---

## Day 1 — Operations & Maintenance

### 5. Defragmentation is a "stop-the-world" hand grenade
- **Frequency:** High
- **Severity:** High
- **Description:** Defrag locks the entire database, blocking reads and writes. Run it on the leader of a busy cluster and you get cascading API 500s, write latency spikes to 5+ seconds, and rejected requests. There is no built-in safe, cluster-aware orchestration that defrags the leader last and waits for raft to catch up.
- **Quotes:**
  - "defrag stops the process completely for noticeable duration" — etcd #9222
  - Kubernetes #93280: "write latencies increased to upwards of 5s... approximately 1 minute of severe degradation. Clients received 'timed out' and 'too many requests' errors."
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/9222
  - https://github.com/kubernetes/kubernetes/issues/93280
  - https://github.com/etcd-io/etcd/issues/20115 — "Automatic defrag"
  - https://github.com/etcd-io/etcd/issues/21740 — "ETCD learner members cannot be defragmented, causing persistent storage bloat"

### 6. Compaction is necessary but easy to misconfigure
- **Frequency:** High
- **Severity:** High
- **Description:** Compaction marks old revisions for deletion but does *not* return disk space — defrag does. Users assume compaction is enough and then watch disk usage climb. Conversely, too-aggressive compaction breaks watches: "etcdserver: mvcc: required revision has been compacted." A regression in 3.5.16 made compaction itself a blocking operation, causing 5-second latency spikes every 5 minutes.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/19406 — "Performance regression in etcd v3.5.16"
  - https://github.com/kubernetes/kubernetes/issues/47131 — "apiserver timeouts after rolling-update of etcd cluster" ("etcdserver: mvcc: required revision has been compacted")
  - https://github.com/kubernetes/kubernetes/issues/80513 — "Allow for more frequent etcd compaction"
  - https://devopsbeast.com/blog/etcd-compaction-defrag-guide — "compaction does not reduce the on-disk size of the etcd database"

### 7. NOSPACE / database-quota-exceeded silently freezes the cluster
- **Frequency:** High
- **Severity:** Critical
- **Description:** Hitting `quota-backend-bytes` raises a NOSPACE alarm; etcd goes read-only; kube-apiserver writes fail; users cannot delete the very objects (e.g. orphaned secrets) that filled the quota in the first place — you need to compact/defrag *first* and then disarm the alarm. A frequent and traumatic incident.
- **Quotes:**
  - "etcdserver: mvcc: database space exceeded" / "making your cluster become read-only. No new pods, no config changes, no deployments." — kubenatives
  - kops outage: "etcd had reached its logical quota limit ... thousands of unnecessary Secrets in the logging-operator namespace ... consumed the 2.2GB etcd quota."
- **Sources:**
  - https://medium.com/@caue._/kubernetes-etcd-out-of-space-on-kops-a-real-life-incident-and-recovery-a1857f3e0998
  - https://etcd.io/blog/2023/how_to_debug_large_db_size_issue/
  - https://dev.to/kashishtwts/etcd-mvcc-database-space-exceeded-full-recovery-guide-1l5e
  - https://docs.mirantis.com/mke/3.8/ops/administer-cluster/manage-etcd/etcd-alarms-response.html

### 8. Certificate expiry → entire cluster dies at once
- **Frequency:** High
- **Severity:** Critical
- **Description:** etcd peer + client TLS certificates and the CAs that sign them have hard expiry. When they roll past midnight, **all** etcd traffic stops, kubectl breaks everywhere, and you cannot easily renew certs because you can no longer talk to the API server you need to fix it. Worse: etcd cannot hot-reload changes to `peer-trusted-ca-file` / `trusted-ca-file`, so a zero-downtime CA rotation is impossible without an outage (issue #11555, open since 2020).
- **Quotes:**
  - "Everything breaks at once. All kubectl commands fail." — kubenatives, on cert expiry
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/11555 — "ETCD doesn't automatically load changes to ca bundles"
  - https://docs.cloud.google.com/kubernetes-engine/distributed-cloud/bare-metal/docs/troubleshooting/expired-certs
  - https://kubernetes.recipes/recipes/troubleshooting/fix-certificate-expiration-cluster/
  - https://www.baeldung.com/ops/kubernetes-expired-certificates

### 9. Member replacement / scaling without downtime is fiddly
- **Frequency:** High
- **Severity:** High
- **Description:** Add-then-remove vs. remove-then-add ordering, learner promotion, peer URL updates, persistent peer state in WAL — getting the sequence wrong drops quorum. Auto-promotion of learners is still a manual operation in plain etcd (#15107). KKP and OKD docs both warn that the member must be removed from internal management before being replaced; misordering kills the cluster.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/6114 — "Adding replacement member before removing"
  - https://github.com/etcd-io/etcd/issues/15107 — "Support auto promoting learner member to voting member"
  - https://github.com/etcd-io/etcd/issues/16002 — "Incorrect configuration in the new etcd member can bring down the etcd cluster"
  - https://github.com/etcd-io/etcd/issues/8169 — `ETCD_FORCE_NEW_CLUSTER` cluster-ID-mismatch traps

### 10. Rolling restarts / upgrades — apiserver doesn't fail over cleanly
- **Frequency:** High
- **Severity:** High
- **Description:** Rolling-restarting etcd members causes kube-apiserver timeouts (k/k #47131), and on at least one major K8s version (#72102) kube-apiserver would refuse to start at all if the *first* etcd in its endpoint list was unavailable — no failover, total outage. Stuck TCP connections that survive a leader's death (Grafana's incident) cause minutes of partial outage even when etcd itself is fine.
- **Sources:**
  - https://github.com/kubernetes/kubernetes/issues/47131
  - https://github.com/kubernetes/kubernetes/issues/72102
  - https://grafana.com/blog/how-a-production-outage-in-grafana-clouds-hosted-prometheus-service-was-caused-by-a-bad-etcd-client-setup/
  - https://github.com/etcd-io/etcd/issues/9949 — "etcd go client fails when querying a cluster with a down node"

### 11. Backups: scheduling, verification, retention, restore drills — all DIY
- **Frequency:** High
- **Severity:** High
- **Description:** Standard tutorials end at "run `etcdctl snapshot save` on a cron". Verification, integrity testing, multi-cloud object-store upload, retention policies, and (most importantly) *tested* restore drills are left as an exercise for the operator. Issue etcd-io/etcd#21283 is literally a request for "built-in backup verification and integrity testing", and #16962 asks for point-in-time recovery — neither exists in upstream etcd.
- **Quotes:**
  - "Always verify before trusting a backup." — Chandrashekhar, Hashnode write-up
  - "Panic is optional — preparation is not."
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/21283 — "Add built-in backup verification and integrity testing"
  - https://github.com/etcd-io/etcd/issues/16962 — "Point in time recovery in ETCD"
  - https://devopscube.com/backup-etcd-restore-kubernetes/
  - https://kmchandrashekhar.hashnode.dev/when-etcd-fails-how-i-backed-up-and-restored-a-kubernetes-cluster

---

## Day 2 — Failure Recovery & Disaster

### 12. Quorum loss recovery is an "intricate procedure" with multiple footguns
- **Frequency:** Medium
- **Severity:** Critical
- **Description:** Once more than (N-1)/2 members are permanently lost, the cluster is unrecoverable from inside etcd. The recovery dance — pick the member with the latest revision, `etcdutl snapshot restore`, `--force-new-cluster`, re-add members one by one, re-issue certs — is long, manual, and prone to typos. Platform9 outright tells customers to open a support ticket. The `--force-new-cluster` flag itself causes cluster-ID mismatches if misused.
- **Quotes:**
  - "Etcd restore is an intricate procedure." — Platform9 KB
  - "Restoring a multi-master cluster from an ETCD backup is a complicated process." — Platform9 KB
  - "There is no API/process in the product currently to do recovery/replacement for a single master node within the cluster." — Platform9
  - "If the cluster permanently loses more than (N-1)/2 members then it disastrously fails, irrevocably losing quorum." — etcd.io docs
- **Sources:**
  - https://platform9.com/kb/pmk/how-to/restore-etcd-cluster-from-quorum-loss
  - https://etcd.io/docs/v3.5/op-guide/recovery/
  - https://gardener.github.io/etcd-druid/usage/recovering-etcd-clusters.html
  - https://github.com/etcd-io/etcd/issues/17638 — "Offline member addition/removal for restoration after quorum loss"
  - https://github.com/etcd-io/etcd/issues/8169

### 13. Unclean shutdowns / power loss → DB corruption that won't self-heal
- **Frequency:** Medium
- **Severity:** Critical
- **Description:** A *coordinated* power loss to a 3-node cluster (e.g. AZ outage, host reboot, OOM kill at the wrong moment) leaves bbolt unable to recover its page tracker. Symptoms: "freepages: failed to get all reachable pages", "snap: snapshot file doesn't exist", segfaults, "page expected to be: X but self identifies as 0". The only fix is restore from backup.
- **Quotes:**
  - "k8s does not start after boot." — k/k #88574
  - User on #11949: even a 3-node cluster wasn't resilient to a coordinated power failure; recovery required restoring from external backups.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/11949 — "Etcd start failed after power off and restart"
  - https://github.com/kubernetes/kubernetes/issues/88574 — "etcd and kube-apiserver does not start after incorrect machine shutdown"
  - https://github.com/etcd-io/etcd/issues/18096 — "etcd panic: assertion failed: Page expected to be: 36312"
  - https://github.com/etcd-io/etcd/issues/10722 — "ectd failed to get all reachable pages"
  - https://github.com/etcd-io/etcd/issues/18881 — "etcdserver: data corruption detected, unable to start etcd member"

### 14. Data corruption / inconsistency bugs across releases
- **Frequency:** Medium
- **Severity:** Critical
- **Description:** etcd has shipped real data-inconsistency bugs in production — auth-on-snapshot inconsistency (3.5.x), revision-divergent members, online-defrag-mid-crash inconsistency, "DB out of sync undetected", and the 3.4.18 inconsistency thread. The robustness test program and `ahrtr/etcd-issues` knowledge base exist *because* this keeps happening.
- **Quotes:**
  - etcd's own KB on online defrag: "etcd might run into data inconsistency issue if it crashes in the middle of an online defragmentation operation using etcdctl."
- **Sources:**
  - https://github.com/ahrtr/etcd-issues
  - https://github.com/etcd-io/etcd/issues/13766 — "Inconsistent revision and data occurs"
  - https://github.com/etcd-io/etcd/issues/14211 — "3.4.18 etcd data inconsistency"
  - https://github.com/etcd-io/etcd/issues/20059 — "etcd data inconsistency in 3.5.11"
  - https://github.com/etcd-io/etcd/issues/12535 — "etcd server DB out of sync undetected"
  - https://github.com/etcd-io/etcd/issues/20418 — "Stale reads caused by process pausing"

### 15. Snapshot/restore itself has bugs
- **Frequency:** Low–Medium
- **Severity:** High
- **Description:** Even the recovery path is not bulletproof: `etcdutl snapshot restore` OOMs on big DBs (#16052), restored 3.6 clusters bring back ghost members (#20967), 3.6 fails to start with `--force-new-cluster` and existing storage (#20009), and restore sometimes fails because the snapshot itself was corrupted/hashless (#18340).
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/16052 — "etcdutl snapshot restore cannot allocate memory"
  - https://github.com/etcd-io/etcd/issues/20967 — "Upgrade to etcd 3.6 recovers old/duplicate members from store"
  - https://github.com/etcd-io/etcd/issues/20009 — "3.6.0: server fails to start, nil pointer exception when re-using storage with force-new-cluster"
  - https://github.com/etcd-io/etcd/issues/18340 — "Etcd send a corrupt snapshot or missing hash snapshot... causes the restoration to fail"

### 16. Stuck disks / slow fsync silently cascade through Raft into mass lease revocation
- **Frequency:** High
- **Severity:** Critical
- **Description:** A *single* node with a stuck or slow disk can: (a) cause heartbeat timeouts, (b) trigger leader elections, (c) get demoted but keep handling client keepalives, and (d) once unstuck, mass-revoke every lease in the cluster. That cascades into Kubernetes node "NotReady", controller restarts, and split-brain-like behaviour.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/15247 — "All leases are revoked when the etcd leader is stuck in handling raft Ready due to slow fdatasync"
  - https://github.com/etcd-io/etcd/issues/14338 — "Removed etcd member failed to stop on stuck disk"
  - https://github.com/etcd-io/etcd/issues/13648 — "High io load on low load cluster, causes range errors and leader re-election"
  - https://github.com/etcd-io/etcd/issues/10799 — "etcd performance issue when disk IO looks good"

---

## Performance & Scale

### 17. Disk fsync latency is the dominant performance signal — and nobody monitors it
- **Frequency:** High
- **Severity:** Critical
- **Description:** Every Raft commit must `fsync` the WAL. P99 > 10 ms is a warning, > 25 ms is a problem; most teams alert on CPU/memory instead. Cloud "burstable" volumes, shared-with-OS disks, and noisy-neighbour environments all push fsync into the danger zone with no warning until kube-apiserver timeouts arrive.
- **Quotes:**
  - "P99 above 10ms = warning. P99 above 25ms = problem — yet this measurement rarely appears in standard dashboards." — rack2cloud
  - HN top thread title: "When etcd crashes, check your disks first."
- **Sources:**
  - https://news.ycombinator.com/item?id=47098324
  - https://www.rack2cloud.com/etcd-kubernetes-database/
  - https://kubernetes.recipes/recipes/troubleshooting/etcd-performance-troubleshooting/
  - https://www.ibm.com/support/pages/how-troubleshoot-etcd-performance-issues-causing-cluster-instability
  - https://github.com/etcd-io/etcd/issues/17615 — "Add metrics to record time costs of write system call in wal"

### 18. Memory blowups: watch starvation, slow watchers, lease storms, member-catchup
- **Frequency:** Medium
- **Severity:** High
- **Description:** A long tail of memory pathologies: watch starvation can OOM the server (#16839), slow watchers degrade PUT latency for everyone (#18109), expiring leases cause p99 spikes (#9360, #15247), and member-catchup buffers (#17098) eat RAM. A simple 1-node cluster with frequent updates has consumed 12 GB RAM on a 2 MB DB (#18382).
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/16839 — "Watch starvation can cause OOMs"
  - https://github.com/etcd-io/etcd/issues/18109 — "Slow watchers impact PUT latency"
  - https://github.com/etcd-io/etcd/issues/18382 — "High memory usage with 1 node cluster"
  - https://github.com/etcd-io/etcd/issues/9360 — "Poor performance and cluster stability when lots of leases are expiring"
  - https://github.com/etcd-io/etcd/issues/17098 — "Reduce memory usage of etcd member catchup mechanism"

### 19. The 8 GB ceiling and the absence of horizontal scale
- **Frequency:** Medium
- **Severity:** High
- **Description:** etcd is a single-shard system with a 2 GB default / 8 GB documented max DB size. Beyond that, range LIST operations dominate. Kubernetes users with millions of CRD objects (e.g. Argo) have to ask for `--etcd-servers-overrides` to split CRDs to a *separate* etcd cluster (k/k #118858). There is no sharding in upstream etcd.
- **Quotes:**
  - "Increased request loads can slow down etcd's responsiveness... High traffic volumes can undermine etcd's ability to maintain consistent data across nodes." — Afzal, Medium
- **Sources:**
  - https://github.com/kubernetes/kubernetes/issues/118858 — "Consider providing separate etcd destination for CRDs"
  - https://github.com/etcd-io/etcd/issues/19806 — "Performance improvement ideas for v3.8"
  - https://medium.com/@emafzal/enabling-etcd-for-large-scale-kubernetes-clusters-72dc00d2388d
  - https://github.com/etcd-io/etcd/issues/20696 — "Proposal: Support witness in etcd"

### 20. Leader elections under pressure cause user-visible jitter
- **Frequency:** High
- **Severity:** Medium
- **Description:** Network blips, slow disks, or GC pauses can trigger leader elections; during an election no writes can be committed. Frequent elections == intermittent kubectl hangs. Grafana's outage hinged on a leader going away abruptly and clients holding onto stale TCP connections.
- **Sources:**
  - https://grafana.com/blog/how-a-production-outage-in-grafana-clouds-hosted-prometheus-service-was-caused-by-a-bad-etcd-client-setup/
  - https://kubernetes.recipes/recipes/troubleshooting/etcd-high-latency-troubleshooting/
  - https://github.com/etcd-io/etcd/issues/13648

---

## Backup & Restore

### 21. No native scheduled-backup / restore-orchestration in upstream etcd
- **Frequency:** High
- **Severity:** High
- **Description:** Upstream etcd offers `etcdctl snapshot save`. Everything beyond that — scheduling, deltas, multi-cloud upload, retention, lifecycle, garbage collection, restore verification, automatic restore on quorum loss — has to be built externally (Velero, etcd-backup-restore, custom scripts). Issue #21283 confirms even upstream considers built-in integrity testing missing.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/21283 — "feat: Add built-in backup verification and integrity testing"
  - https://github.com/etcd-io/etcd/issues/16962 — "Point in time recovery in ETCD"
  - https://github.com/gardener/etcd-backup-restore — exists precisely to fill this gap
  - https://devopscube.com/backup-etcd-restore-kubernetes/

### 22. Backups that "work" until you need them
- **Frequency:** Medium
- **Severity:** Critical
- **Description:** Untested backups, snapshots saved to the *same* disk that died, snapshot files without integrity hashes (#18340), snapshots that won't restore on a newer minor (3.6 upgrade ghost-members #20967, #20009). The community theme: nobody discovers their backup strategy is broken until disaster strikes.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/18340
  - https://github.com/etcd-io/etcd/issues/20967
  - https://github.com/etcd-io/etcd/issues/20009
  - https://github.com/gardener/etcd-backup-restore/issues/963 — "Restoration failed while compacting the snapshots"

### 23. Multi-cloud / object-store credential & endpoint plumbing
- **Frequency:** Medium
- **Severity:** Medium
- **Description:** Operators want S3, GCS, ABS, Swift, OSS, and on-prem S3-compatible stores; they want IRSA / Workload Identity instead of static keys; they want custom endpoints (MinIO, localstack). Without an opinionated tool (etcd-backup-restore, Velero), every team writes a fragile wrapper.
- **Sources:**
  - https://github.com/gardener/etcd-backup-restore/issues/943 — "override cloud provider bucket's endpoint"
  - https://github.com/gardener/etcd-backup-restore/issues/970 — credential file refactor
  - https://github.com/gardener/etcd-backup-restore/issues/1010 — LocalStack alternatives for S3 emulation

---

## Multi-Cluster / Multi-Tenancy

### 24. Running and managing *many* etcd clusters at scale
- **Frequency:** Medium (high among hyperscalers / managed-K8s providers)
- **Severity:** High
- **Description:** A managed-Kubernetes platform (Gardener, OpenShift, EKS-like) has to run *hundreds to thousands* of etcd clusters, each with its own lifecycle, certs, backups, quotas, and version. Upstream etcd has no concept of "fleet operator". This is precisely the gap etcd-druid was created to fill.
- **Sources:**
  - https://github.com/gardener/etcd-druid (entire project description)
  - https://gardener.github.io/etcd-druid/
  - https://github.com/etcd-io/etcd/issues/19798 — "Unify the approach to manage etcd images in Kubernetes"

### 25. Noisy tenants & isolation
- **Frequency:** Medium
- **Severity:** Medium
- **Description:** etcd has no per-tenant rate limiting or per-tenant quotas. A single noisy CRD or watch hot-path degrades everyone. k/k #118858 (separate etcd for CRDs) and etcd #21513 (server-side adaptive rate limiting and memory-aware admission control) are explicit asks for tenancy controls.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/21513
  - https://github.com/kubernetes/kubernetes/issues/118858
  - https://github.com/etcd-io/etcd/issues/16287 — "Performance Issue: Getting error etcdserver: too many requests during high load"

---

## Observability & Visibility

### 26. "Why is etcd unhealthy?" — diagnostics require deep internal knowledge
- **Frequency:** High
- **Severity:** High
- **Description:** Operators see symptoms (slow kubectl, pods pending, controllers lagging) and have to guess: disk? network? compaction? leader churn? quota? The CNCF diagnostic-tooling blog and etcd-io/etcd #20217 ("Implement a comprehensive diagnosis tool") exist because the *current* tooling is a scatter of `etcdctl endpoint status`, `etcd-dump-db`, `etcd-dump-logs`, `bbolt`, and pprof.
- **Quotes:**
  - "Operators lack visibility into why etcd behaves unhealthily, forcing them to piece together incomplete information under pressure." — CNCF blog
  - "When issues arise, maintainers engage in lengthy back-and-forth exchanges to gather diagnostic information." — etcd #20217
- **Sources:**
  - https://www.cncf.io/blog/2026/03/12/making-etcd-incidents-easier-to-debug-in-production-kubernetes/
  - https://github.com/etcd-io/etcd/issues/20217
  - https://github.com/etcd-io/etcd/issues/13454 — "etcd request logs"
  - https://github.com/etcd-io/etcd/issues/3272 — "etcd realtime workload analysis tooling"
  - https://github.com/etcd-io/etcd/issues/12460 — "Add Distributed Tracing using OpenTelemetry"

### 27. Vague / unhelpful error messages
- **Frequency:** High
- **Severity:** Medium
- **Description:** "apply request took too long", "mvcc: database space exceeded", "context deadline exceeded", "request timed out", "rafthttp: failed to dial" — none of them point operators at the actual root cause. k/k #112152 is literally a feature request to "map etcd errors to proper response status" because apiserver propagates this opacity.
- **Sources:**
  - https://github.com/kubernetes/kubernetes/issues/112152
  - https://github.com/etcd-io/etcd/issues/10642 — "[etcdctl] Error: context deadline exceeded"
  - https://github.com/etcd-io/etcd/issues/11809 — "etcdserver: request timed out"
  - https://github.com/etcd-io/etcd/issues/10587 — "(error \"EOF\", ServerName \"\") error on etcd servers"

### 28. Log spam that obscures real signals
- **Frequency:** Medium
- **Severity:** Medium
- **Description:** kube-apiserver fills its logs with "failed to connect to etcd" every 15 seconds even when the cluster is healthy (k/k #134080); cluster operators can't tell normal noise from a real degradation. Similar log-noise complaints span releases.
- **Sources:**
  - https://github.com/kubernetes/kubernetes/issues/134080
  - https://github.com/etcd-io/etcd/issues/18805 — "etcd_cluster_version metric is missing"

---

## Security & Compliance

### 29. TLS / mTLS hot reload is broken
- **Frequency:** Medium
- **Severity:** High
- **Description:** As described in pain #8 above: etcd cannot reload the trusted-CA bundle, so zero-downtime CA rotation is impossible. Combine this with the year-or-shorter expiry of kubeadm-generated certs and you have a recurring planned-outage tax.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/11555
  - https://github.com/kubernetes/kubernetes/issues/83028

### 30. Secrets are stored in etcd, often unencrypted
- **Frequency:** Medium
- **Severity:** High
- **Description:** k/k #12742 ("encrypt secrets when in etcd") was open for years; encryption-at-rest is opt-in, key management is the user's problem, and unprotected snapshots are a real exfiltration risk. Plenty of clusters still run with `EncryptionConfiguration` undeployed.
- **Sources:**
  - https://github.com/kubernetes/kubernetes/issues/12742
  - https://github.com/etcd-io/etcd/issues/9475 — "Secure etcd by default"
  - https://github.com/etcd-io/etcd/issues/11651 — "data corruption bug in all etcd3 version when authentication is enabled"

### 31. Auth and snapshot interactions
- **Frequency:** Low
- **Severity:** Critical (when it bites)
- **Description:** etcd has shipped multiple bugs at the intersection of auth + snapshot/restore — including data inconsistency when auth info is not reloaded after a snapshot, and a data-corruption bug specifically when auth was enabled.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/11651
  - https://github.com/ahrtr/etcd-issues (auth-related-regression section)

---

## Knowledge Gap — "etcd is a black box"

### 32. Almost everyone running etcd in production didn't choose to
- **Frequency:** High
- **Severity:** High
- **Description:** They installed Kubernetes. etcd came with it. Most teams have no etcd expert; their entire mental model is the kubeadm reference architecture. When etcd misbehaves they reach for `kubectl` (which is down) and then Google `mvcc: database space exceeded`. The CNCF blog calls this out explicitly: the gap between "something is wrong" and "I know what's wrong" is "where most time is lost during an incident."
- **Quotes:**
  - "Operators struggle because diagnosing issues historically required deep knowledge of etcd internals, manual log and metric collection, and understanding which signals matter." — CNCF blog
  - "etcd failures rarely announce themselves clearly." — CNCF blog
- **Sources:**
  - https://www.cncf.io/blog/2026/03/12/making-etcd-incidents-easier-to-debug-in-production-kubernetes/
  - https://github.com/ahrtr/etcd-issues — community-built KB *because* upstream is impenetrable
  - https://www.kubenatives.com/p/etcd-debugging-kubernetes
  - https://www.rack2cloud.com/etcd-kubernetes-database/

### 33. Upstream etcd-as-a-database is great; etcd-as-an-operator-experience is missing
- **Frequency:** High
- **Severity:** High
- **Description:** Compare the operator experience of Postgres (operators, pgBouncer, Patroni, pgBackRest, RDS) or Kafka (Strimzi, MSK) to etcd. There is no single, "obvious" answer for how to safely run etcd at scale. Upstream is intentionally minimal; the gap is filled by ad-hoc scripts, kubeadm static pods, Velero, CoreOS' legacy operator, etcd-druid (Gardener), Red Hat's cluster-etcd-operator, KubeBlocks. None is standard.
- **Sources:**
  - https://github.com/etcd-io/etcd/issues/19371 — "Develop a caching library for etcd" (top-reactions open issue)
  - https://github.com/etcd-io/etcd/issues/19798
  - https://github.com/gardener/etcd-druid
  - https://github.com/openshift/cluster-etcd-operator

---

## Top 10 Most Common Problems (driving the talk)

A ranked, frequency × severity composite from everything above. These are the pain points the audience at a Bangalore meetup is *most likely to have lived through*, and they map cleanly onto features in etcd-druid + etcd-backup-restore.

| # | Pain Point | Why it hurts | Maps to (etcd-druid / etcd-backup-restore feature) |
|---|---|---|---|
| 1 | **etcd is a black box** — operators inherit it via Kubernetes and have no mental model when it breaks | High frequency, high severity, blocks every other recovery | The whole *point* of an operator: encode etcd expertise as automation |
| 2 | **NOSPACE / 2 GB quota** — silent freeze of the entire cluster, can't even delete the bloat | Critical severity, very common | Druid-managed quota, automated compaction & defrag in etcd-backup-restore |
| 3 | **Quorum loss recovery** is manual, intricate, error-prone | Critical severity, infrequent but catastrophic | Druid auto-detects quorum loss and triggers `etcdbr` restore; offline member operations |
| 4 | **Backups: scheduling, verification, restore drills** are DIY | High severity (silent failure mode), high frequency | etcd-backup-restore: scheduled full + delta snapshots, multi-cloud, restore-on-init |
| 5 | **Defragmentation is "stop-the-world"** and easy to do wrong on the leader | High severity, high frequency, no upstream safe-defrag | etcd-backup-restore safe rolling defrag; Druid coordinates member-by-member |
| 6 | **Disk fsync latency** is the dominant performance signal that nobody monitors | Critical severity when it fires, often unseen | Druid surfaces etcd fsync/proposal metrics; runs etcd-wrapper with proper resources |
| 7 | **Certificate expiry & rotation** without a hot-reload path | Critical severity (everything dies at once) | Druid-managed cert lifecycle via secrets; restart orchestration |
| 8 | **Unclean shutdown → bbolt corruption** that won't self-heal | Critical severity when it fires | Druid validates member with etcd-wrapper before start; auto-restore from backup |
| 9 | **Member replacement / scaling** without downtime is fiddly | High severity, frequent during ops | Druid handles add-then-remove ordering, peer URL updates, learner promotion |
| 10 | **Diagnostics & visibility** — vague errors, scattered tools, log spam | High frequency, multiplies MTTR for everything else | Druid status conditions (Ready / AllMembersReady / BackupReady), etcd-wrapper validation, structured logs |

These 10 are the talk. Every one of them is something etcd-druid + etcd-backup-restore demonstrably automates or de-risks.
