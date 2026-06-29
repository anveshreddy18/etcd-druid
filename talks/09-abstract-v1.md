# The hardest part of running etcd is not running etcd

Upstream etcd is excellent software. What is missing is everything around it. The scheduled snapshot that nobody ever tried to restore from. The defragmentation job that locks the database for five seconds when it lands on the leader. The certificate rotation that turns into a yearly planned outage. The permanent quorum loss that ends with "open a support ticket."

etcd-druid and etcd-backup-restore are an open-source answer to that gap. A single CRD provisions a TLS-secured, backup-configured cluster; member-by-member defrag avoids stop-the-world spikes; on every pod start the data directory is validated and auto-restored if corrupt. The interesting part: every compaction job restores the snapshot chain into an embedded etcd, so you learn that a backup is broken before disaster, not during it. SAP Gardener has run this stack for over eight years across the largest known multi-tenant etcd fleet in open source, alongside Akamai, StackIT and Scaleway. Both projects are now under Linux Foundation NeoNephos governance.

You will leave with a concrete picture of what production etcd needs beyond the binary, the failure modes that justify each piece, and a path to adopting them in your own clusters.
