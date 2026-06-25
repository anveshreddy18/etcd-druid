# Title: Nobody Chose to Run etcd. We Run 10,000 of Them. Here's What We Learned.

**Speakers:** [team — placeholder for now]
**Format:** Bangalore K8s Meetup talk (~30 min)

## Abstract

Almost nobody in this room *chose* to run etcd. It came bundled with Kubernetes — and most teams operate it with a mental model that ends at the kubeadm reference architecture. That works until 3 AM, when `mvcc: database space exceeded` shows up, kubectl stops responding, and you discover you cannot delete the orphaned secrets that filled the 2 GB quota *because the API needed to delete them is the one that's down*.

Here is the counter-intuitive part: the hard part of running etcd is not running etcd. Upstream etcd is excellent software. The gap is everything around it — safe defragmentation that doesn't stop the world on the leader, scheduled snapshots that someone actually tested with a restore, quorum-aware rolling updates that don't bring the cluster down during a routine image bump, certificate rotation that doesn't become a yearly planned outage, and recovery from permanent quorum loss without a Platform9 support ticket. None of this ships with etcd. Most teams pay an invisible engineering tax rediscovering each of these the hard way.

We have been operating production etcd at SAP's Gardener for 8+ years across **tens of thousands** of control-plane clusters, with zero recorded data loss. Along the way we built — and now maintain in the open — **etcd-druid** (the operator) and **etcd-backup-restore** (the sidecar). They encode every painful lesson the docs won't tell you: leader-only safe defrag, multi-cloud backups with WORM immutability, dual-site backup sync across providers, automatic data-directory validation and restore on every pod start, single-node ⇄ multi-node live scale-out with auto peer-TLS upgrade, and a pluggable EtcdOpsTask platform for day-2 operations. Both projects are now under the **Linux Foundation's NeoNephos** sub-foundation — vendor-neutral, used in production by Akamai, StackIT, and Scaleway alongside SAP.

In 30 minutes we will walk the Top 10 operational pains every etcd operator eventually meets, and show — with live failure scenarios — what self-healing actually looks like at scale.

## Why you should attend
- You inherited etcd with Kubernetes, and you suspect you are one bad Friday away from learning it the hard way.
- You want to see a real, open-source answer to backups, defrag, cert rotation, and quorum loss — not slideware.
- You will leave with a runnable mental model for what production-grade etcd operations looks like, and the project to drop into your stack on Monday.

## Reviewer notes
- **Hook strategy:** Pattern A (battle story) + Pattern D (counter-intuitive thesis), per the corpus. Opens with audience identification ("nobody chose to run etcd") before pivoting to the counter-intuitive claim ("the hard part is not running etcd"). The NOSPACE anecdote is the specific named scar that earns the rest of the abstract.
- **Curiosity gap:** The abstract names *what* the gap is (defrag, snapshots, rolling updates, certs, quorum loss) but doesn't reveal *how* druid solves each one. The Top 10 list and live failure scenarios are deliberately held back for the room — the reader senses there is a concrete, demonstrable answer without getting the whole talk.
- **Credibility placement:** Numbers ("8+ years", "tens of thousands", "zero recorded data loss") arrive in paragraph three, *after* the audience has been put in the protagonist seat — so the credibility lands as reassurance, not a brag. Named users (SAP, Akamai, StackIT, Scaleway) and LF NeoNephos governance follow naturally, not in a "logos" block.
- **Why this angle works for Bangalore:** Bangalore K8s meetup attendees skew toward platform/SRE practitioners at companies running real production Kubernetes — many at hyperscalers, ISVs, and managed-K8s providers. They have lived through NOSPACE, certificate-expiry outages, and the 2 AM defrag. The counter-intuitive opener ("nobody chose this") gives them permission to admit etcd is opaque to them, which a "we'll teach you etcd internals" framing wouldn't. It also sidesteps a common India-meetup failure mode — talks that read as KubeCon recycled — by anchoring on a problem the audience self-identifies with from the first sentence.
- **Honesty as a trust device:** "Zero recorded data loss" is verifiable (Gardener's public posture); we deliberately avoid claiming druid solves the 8 GB ceiling or in-cluster multi-tenancy. The Reviewer-facing version of the talk will include the honest "open gaps" section from `05-problem-solution-map.md`.
- **Closing line strategy:** The final sentence promises a runnable artifact ("live failure scenarios … self-healing actually looks like at scale") — Pattern G from the corpus. It tells the CFP reviewer the talk has a demo, not just a narrative.
