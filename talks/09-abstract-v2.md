# The hardest part of running etcd is not running etcd

The scheduled snapshot nobody ever tried to restore from. The defragmentation job that locks the leader for five seconds on a busy Wednesday. The CA bundle etcd refuses to hot-reload, turning cert rotation into a yearly planned outage. The `mvcc: database space exceeded` that flips kube-apiserver read-only at 3 AM. Upstream etcd is excellent software; everything around it is where production teams bleed, and where the CoreOS etcd-operator stopped answering when it was archived in 2020.

This talk walks through what an etcd operator actually has to do, using etcd-druid and etcd-backup-restore as the worked example. One CR provisions a TLS-secured, backup-configured cluster. Defrag runs follower-first, leader-last, driven by a leader-elected sidecar. Every pod start runs a validator that auto-restores from snapshot if the data directory is corrupt. Every compaction job restores the snapshot chain into an embedded etcd, so you find a broken backup before disaster. SAP Gardener has run this stack since 2018 across tens of thousands of clusters, alongside Akamai, StackIT, and Scaleway.

You will leave knowing which failure modes the operator pattern eliminates, which ones it merely contains, and which ones, like permanent quorum loss and the 8 GB ceiling, still live in your runbook.
