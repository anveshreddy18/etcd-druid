# Title: It's 3 AM. Your etcd Quorum Is Gone. Now What?

**Speakers:** [etcd-druid maintainers — placeholder]
**Format:** Bangalore K8s Meetup talk (~30 min)

## Abstract

It's 3 AM. PagerDuty is screaming. Your etcd cluster has hit `mvcc: database space exceeded` — orphaned secrets filled the 2 GB quota, the API is read-only, and you can't `kubectl delete` the bloat because kubectl needs the API that just died. The runbook is 47 steps. Step 12 says "pick the member with the latest revision." You don't know how. Your last restore drill was... when, exactly?

Almost nobody chose to run etcd. It came bundled with Kubernetes — and the moment it misbehaves, you're debugging Raft, bbolt page tables, and fsync latency in production, blind. NOSPACE freezes. Certificate-expiry blackouts at midnight. Defrag on the leader spiking write latency to 5 seconds. PVs returning corrupt after a node reboot. Permanent quorum loss that Platform9's KB literally answers with "open a support ticket."

What if you never had to do any of this again?

In this talk we introduce **etcd-druid** and **etcd-backup-restore** — the open-source Kubernetes operator and sidecar pair that encode 8+ years of etcd operational scars as automation. Declarative `Etcd` CRs replace YAML checklists. Scheduled full+delta snapshots across **eight** backup providers — with immutable (WORM) backups and dual-site cross-cloud sync that no other open-source etcd operator ships. Validator-driven auto-restore on every pod start. Leader-aware, member-by-member defrag. Quorum-aware rolling updates. Live 1→N scale-out with auto peer-TLS upgrade. And an `EtcdOpsTask` framework that turns day-2 operations into one-line CRs.

This isn't theory. It runs in production today behind **SAP's Gardener, Akamai Connected Cloud, StackIT, and Scaleway** — managing **tens of thousands of etcd clusters with zero recorded data loss**. Now a Linux Foundation project under the **NeoNephos** sub-foundation, with vendor-neutral governance and an active maintainer community.

We'll show you what self-healing etcd actually looks like. Bring your scars.

## Why you should attend
- You've been paged for etcd at least once — and you'd rather not be paged for it again.
- You want to see a *live* quorum-loss recovery that finishes before the coffee brews.
- You want the operational playbook proven at 10,000-cluster scale, in open source you can fork tonight.

## Reviewer notes
- **Hook strategy:** Pattern A (war story) from the KubeCon corpus — open with sensory detail (3 AM, PagerDuty, the read-only API), then enumerate three more pains in rapid fire so every operator finds themselves in at least one. The "47 steps" and "your last restore drill was... when, exactly?" are designed to land on real shame, not hypothetical risk.
- **Curiosity gap:** Deliberately *no* live-demo description, no architecture diagram, no list of "what you'll learn." The phrase "what self-healing etcd actually looks like" is the bait; the demo is the payoff only attendees see. The "bring your scars" closer invites participation, not passive viewing.
- **Credibility placement:** Scale numbers and named users are deferred to paragraph 4 — after the audience has felt the pain and seen the capability list. This avoids the "brag dump" anti-pattern and uses Pattern F (named users) only once the reader is already invested. NeoNephos is the last credibility beat because vendor-neutrality is the *answer* to the unspoken "is this just a SAP tool?" objection.
- **Differentiation tactic:** "Eight providers," "immutable backups no other open-source etcd operator ships," "dual-site cross-cloud sync" — concrete, falsifiable, unique-in-OSS claims rather than generic "feature-rich" language. CFP reviewers smell vague claims.
- **Why this angle works in Bangalore:** Bangalore's K8s meetup audience is heavily platform/SRE engineers at Indian fintech, SaaS, and managed-K8s vendors — they have all been on the wrong end of etcd at least once, and they read CNCF blogs. The war-story opener short-circuits the "is this relevant to me?" filter. The LF NeoNephos angle matters because Indian enterprises are increasingly wary of single-vendor open source after several high-profile relicensing events.
- **Tonal restraint:** No "revolutionary," no "game-changer," no exclamation marks. The drama is in the specifics (`mvcc: database space exceeded`, 47 steps, 5-second latency), not the adjectives — which is how the accepted KubeCon abstracts read.
