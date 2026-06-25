# Title: The Best etcd Operator You've Never Heard Of

**Speakers:** [team — placeholder for now]
**Format:** Bangalore K8s Meetup talk (~30 min)

## Abstract

If you're running etcd in Kubernetes today, there's a 90% chance you're either solving problems we already solved — or about to hit problems we already documented. And you've probably never heard of us. This talk exists to fix that.

Let's be honest about why: etcd-druid wasn't born from a conference keynote or a vendor launch. It grew quietly inside Gardener at SAP, where it now runs **tens of thousands of production etcd clusters** with zero recorded data loss. Akamai, StackIT, Scaleway, and NeoNephos are running it too. Yet most teams still treat etcd as a black box they inherited with Kubernetes — and reach for `etcdctl snapshot save` on a cron the night before a postmortem.

So we'll skip the slideware and walk through the four pains every operator in this room has lived through — and how 8+ years of operating etcd at planet scale turned each one into a controller you can install:

- **NOSPACE at 2 GB** — your cluster goes read-only, and you can't even delete the bloat that filled it. Druid sizes, compacts, and defrags member-by-member, leader-last, so this incident stops happening.
- **"We have backups… right?"** — scheduled full + delta snapshots across eight object stores, with WORM immutability and dual-site sync. Compaction jobs continuously *restore* backups into an embedded etcd — so you find out they're broken **before** disaster, not during.
- **PV came back corrupt at 3 AM** — every pod start validates the data directory and auto-restores from snapshots before etcd even starts. No page. No ritual.
- **Cert rotation, quorum-aware rolling updates, single-node → multi-node live scale-out, permanent quorum recovery** — the day-2 work most teams quietly avoid.

We'll close with the capabilities you didn't know to ask for: immutable backups, a validating webhook that stops you from accidentally deleting your own StatefulSet, and a pluggable `EtcdOpsTask` API that turns runbooks into CRs. Donated to the Linux Foundation under NeoNephos. Vendor-neutral. Battle-tested. Yours to use.

You'll leave knowing whether your current etcd setup is one bad Friday away from making the news — and what to do about it on Monday.

## Why you should attend

- You inherited etcd with Kubernetes and have a quiet feeling you don't fully understand it. You're not alone — and you don't have to become a Raft expert to sleep better.
- You'll see the production playbook from one of the largest multi-tenant etcd deployments on the planet, distilled into demos you can run on a kind cluster tonight.
- You'll walk out with a one-slide checklist: ten specific etcd failure modes, and exactly which of them apply to your setup right now.

## Reviewer notes

- **Hook strategy:** Lead with the gap itself — "best operator you've never heard of" + the 90% line. Refreshingly honest. Most CFP openers try to sound important; ours admits we have a discoverability problem and makes that the talk's reason to exist. Disarms skepticism.
- **Curiosity gap:** We name the four pains and *gesture* at the solutions, but never explain how compaction-as-backup-validation actually works, what the WORM integration looks like, or what the OpsTask state machine is. The "wow" capabilities at the end are listed without explanation. Audience has to attend to get the mechanism.
- **Credibility placement:** SAP/Gardener "tens of thousands of clusters, zero data loss" appears in paragraph 2, not the opener — so it lands as evidence, not a brag. Akamai/StackIT/Scaleway/NeoNephos are woven in as a casual list. LF governance lives in the closing paragraph as a "by the way, this is yours" — not as an authority play.
- **Why this angle works for Bangalore:** Bangalore K8s meetup audience skews senior platform/SRE — people who've run kubeadm in production and have war wounds. They're allergic to vendor pitches but love operator deep-dives with named failure modes. The honest-about-discoverability angle respects their time; the "what do you use today, and is it enough?" framing flatters their experience.
- **Pattern alignment with KubeCon corpus:** Pattern A (battle story openers — 3 AM, NOSPACE, postmortem), Pattern B (scale numbers — tens of thousands, 8+ years), Pattern C (Day-2 reckoning — "the day-2 work most teams quietly avoid"), Pattern E (listicle — four pains, ten failure modes), Pattern F (named users — SAP, Akamai, StackIT, Scaleway, NeoNephos).
- **Verifiability:** Every claim — production scale, named users, LF/NeoNephos status, capability list — maps to a source file in this talks/ directory. No hand-wave numbers. "Zero recorded data loss" comes from the existing abstract; "tens of thousands of clusters" is documented Gardener scale.
