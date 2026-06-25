# The Hardest Part of Running etcd Isn't Running etcd

**Speakers:** [Placeholder — to be filled by the team]
**Event:** Bangalore Kubernetes Meetup
**Format:** ~30 min talk + Q&A
**Track:** Operations / SRE / Cloud Native

## Abstract

Almost nobody chose to run etcd. It came bundled with Kubernetes, and most teams operate it with a mental model that ends at the kubeadm reference architecture. That works — right up until `mvcc: database space exceeded` shows up at 3 AM, kubectl stops responding, and you discover you cannot delete the orphaned secrets that filled the 2 GB quota *because the API needed to delete them is the one that's down*.

Here is the counter-intuitive part: the hard part of running etcd is not running etcd. Upstream etcd is excellent software. The gap is everything around it — defragmentation that won't stop the world when it lands on the leader, scheduled snapshots that someone actually tested with a restore, quorum-aware rolling updates that survive a routine image bump, certificate rotation that isn't a yearly planned outage, and recovery from permanent quorum loss without — to quote Platform9's KB literally — "opening a support ticket." None of this ships with etcd. Most teams pay an invisible engineering tax rediscovering each of these the hard way.

We have been operating production etcd inside SAP's Gardener for 8+ years, across the largest known multi-tenant etcd fleet in open source, with no project-tracked data-loss incidents. Along the way we built — and now maintain in the open — **etcd-druid** (the Kubernetes operator) and **etcd-backup-restore** (the sidecar). They encode every painful lesson the docs won't tell you. One detail to whet your appetite: every compaction job continuously *restores* backups into an embedded etcd, so you find out a snapshot is broken **before** the outage, not during it. Both projects now live under the Linux Foundation's NeoNephos sub-foundation, with production users beyond SAP — Akamai, StackIT, Scaleway.

In 30 minutes we will walk the operational scars every etcd operator eventually earns — defrag stalls, expired peer certs, NOSPACE, PV corruption, permanent quorum loss — and show, with live failure scenarios on a kind cluster, what self-healing looks like when the cluster is actually broken. We will also name what we don't solve yet. Bring your scars.

## Why you should attend

- You inherited etcd with Kubernetes, and you suspect you are one bad Friday away from learning it the hard way.
- You want a real, open-source answer to backups, defrag, cert rotation, and quorum loss — with the failure scenarios reproduced live, not slideware.
- You will walk out with a runnable mental model, a one-slide checklist of specific etcd failure modes, and a project you can drop into your stack on Monday.

## Speaker bios placeholder

[team to fill]

## Submission notes (for the team — explain final design decisions)

- **Which variant won and why:** The counter-intuitive variant (82/100, highest score). It is the freshest angle in the etcd-talk landscape (Pattern D from the corpus, rare for this topic), gets audience identification on the first line ("nobody chose to run etcd"), and the "the hard part is not running etcd" thesis is the kind of line that travels in a CFP committee's heads. War-story scored lower (74) because its 3 AM trope is over-used; mind-share-direct (76) slipped into vendor-deck cadence in its closer.
- **Which phrases/structures we grafted from the other variants:**
  - From war-story: the `mvcc: database space exceeded` error string in the opener, the Platform9 "open a support ticket" line (promoted from a buried jab into the counter-intuitive thesis paragraph — the critique flagged this as "the single most distinctive sentence" in that variant), the "Bring your scars" closer.
  - From mind-share-direct: the "compaction continuously *restores* backups into an embedded etcd" teaser (the strongest curiosity-gap line of all three variants), and the "kind cluster" demo specificity.
  - Structural move from war-story: enumerating multiple scars in a rapid-fire list in the closing paragraph so every operator finds themselves in at least one.
- **Top critique fixes applied (from the counter-intuitive critique):**
  1. **Cut paragraph 3's feature dump by ~40%.** Removed the comma-separated catalogue (leader-only defrag, WORM, dual-site sync, validation, scale-out, OpsTask). Replaced with a single concrete teaser (compaction-as-validation) — names the *outcome*, withholds the mechanism.
  2. **Killed the "Top 10" framing.** Replaced with named scars ("defrag stalls, expired peer certs, NOSPACE, PV corruption, permanent quorum loss").
  3. **Replaced vague "tens of thousands" with a verifiable framing.** "Largest known multi-tenant etcd fleet in open source" is defensible without a fabricated cluster count; "no project-tracked data-loss incidents" is honest rather than the marketing-flavoured "zero recorded data loss."
  4. **Tightened "Almost nobody in this room"** → "Almost nobody chose to run etcd" (drops 4 words, works for both reader and listener — per the critique's line edit).
  5. **Trimmed the LF NeoNephos sentence** to a single clause naming two non-SAP production users (per critique #8).
  6. **Replaced "self-healing at scale"** with "self-healing when the cluster is actually broken" (kills the "at scale" cliché flagged in #5).
  7. **Replaced "not slideware"** in Why-You-Care bullet 2 with "with the failure scenarios reproduced live, not slideware" — keeps the punch without the purely defensive framing.
  8. **Added an honesty beat** ("We will also name what we don't solve yet") — reviewers reward this; the problem-solution-map explicitly recommends an open-gaps section.
- **Verifiable facts the team must confirm before submission:**
  - "8+ years" of Gardener production operation — corroborate with first commit / first production rollout date.
  - "Largest known multi-tenant etcd fleet in open source" — confirm there is no comparable public claim from CNCF / Red Hat / Oracle.
  - "No project-tracked data-loss incidents" — confirm what the project actually tracks; if there is a public Gardener incident log, cite the timeframe.
  - Akamai, StackIT, Scaleway as production users — confirm each is on-record (public talk, blog, or doc page).
  - LF NeoNephos sub-foundation status — confirm current branding and that "Linux Foundation's NeoNephos sub-foundation" is the right phrasing.
  - Platform9 "open a support ticket" KB — keep a screenshot/URL on hand in case a reviewer challenges.
- **Suggested talk outline (5-7 bullets, 30 min):**
  1. (3 min) **The opener.** Live demo: `mvcc: database space exceeded`, kubectl read-only, the recursive trap. Audience nods.
  2. (4 min) **The thesis.** "The hard part of running etcd isn't running etcd" — what etcd does well, where the gap is, why the deprecated CoreOS/Zalando operator left a vacuum.
  3. (8 min) **Three operational scars, deep.** NOSPACE → compaction + defrag flow; PV corruption → validator + auto-restore on every pod start; permanent quorum loss → today's runbook + the `EtcdOpsTask` roadmap.
  4. (6 min) **The live failure scenario.** Kill a member on a kind cluster, watch druid recover. Single-node → multi-node live scale-out with auto peer-TLS upgrade.
  5. (4 min) **The "wow, you do that?" round.** Immutable (WORM) backups, dual-site cross-cloud sync, the validating webhook that refuses to let you `kubectl edit` the StatefulSet.
  6. (3 min) **Honest gaps.** The 8 GB ceiling, in-cluster multi-tenancy, in-place upgrades still alpha — what we don't solve yet, and why.
  7. (2 min) **Where to find us.** Linux Foundation NeoNephos, GitHub, the recovering-etcd-clusters doc. Q&A.
