Title: From Fragile to Self-Healing: Operationalizing etcd at Scale

Abstract:
Running etcd across thousands of production Kubernetes clusters reveals critical failures: data corruption, quorum loss, member failures, and storage bloat, all capable of disrupting clusters. Managing these manually is complex, error-prone, and difficult to scale.

In this talk, we will be sharing lessons from 8+ years of operating production-grade etcd at scale with close to zero manual intervention and how we built etcd-backup-restore to reliably automate tasks like scheduled backups, corruption detection, quorum loss recovery, member replacement, and scheduled defragmentation, with zero recorded data loss across thousands of clusters.

Used in production by SAP Gardener, Akamai Cloud, StackIT, and Scaleway, etcd-backup-restore is now a Linux Foundation project under Linux Foundation EU via the NeoNephos Foundation, with an active maintainer community. We'll show what it takes to run etcd reliably at scale, and make it truly self-healing.