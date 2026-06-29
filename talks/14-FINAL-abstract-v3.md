# The hardest part of running etcd is not running etcd

**Event:** Bangalore Kubernetes Meetup
**Format:** ~30 min talk + Q&A
**Submission status:** Draft V3, pending team review

Imagine running etcd across thousands of production Kubernetes clusters that hold the data your business runs on. The database itself is excellent software; that part rarely surprises you. What surprises you is everything that has to happen around it for the rest of the team to sleep. The snapshot saved to S3 that nobody has ever tried to restore. The defragmentation job that locks the leader for five seconds on a Wednesday. The certificate that rolls past midnight and takes the API server with it. The day a corrupted volume comes back from a node reboot and your runbook has thirty-one steps in it.

What if you didn't have to live with any of that? Over the last eight years of operating etcd at Gardener's scale, we built and now maintain two open-source projects that answer exactly this gap. etcd-druid is the Kubernetes operator. etcd-backup-restore is the sidecar that ships with it. Together they turn the long tail of etcd operations into ordinary Kubernetes reconciliation. One example, to give you a feel for the shape of the answer: every time the operator runs a compaction job, it quietly restores the snapshot chain into an embedded etcd, so you find out a backup is broken before you ever need it, not during the outage.

Used in production by SAP through Gardener, alongside Akamai Cloud, StackIT, and Scaleway, etcd-druid and etcd-backup-restore are now Linux Foundation projects under NeoNephos, with an active maintainer community. Come for the story of what running etcd at scale actually costs. Leave with a clear picture of what self-healing etcd looks like when it is real, and the open-source tools that get you there.
