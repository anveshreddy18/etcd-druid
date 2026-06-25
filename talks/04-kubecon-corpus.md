# KubeCon Abstract Corpus — what gets etcd / stateful / operator talks accepted

Goal: study 4 years of KubeCon abstracts on etcd, stateful workloads, operators,
backup/restore, and "control plane at scale". Extract the rhetorical patterns
that selection committees say yes to.

Each entry below captures: Title, Speakers + affiliation, Edition, Track,
abstract content (verbatim where retrievable, otherwise paraphrased from the
session card), URL, and a short "why accepted" analysis.

---

## 1. etcd-specific talks

### 1.1 "Lessons Learned From Etcd the Data Inconsistency Issues"
- Speakers: Marek Siarkowicz (Google), Benjamin Wang (VMware, etcd maintainer)
- KubeCon NA 2022, Detroit — Wed Oct 26, 3:25pm EDT, Room 410 A
- URL: https://kccncna2022.sched.com/event/182N9
- Abstract (verbatim):
  > "Earlier the year there was an event that shook the cloud native ecosystem.
  > The latest release of etcd had a critical data inconsistency issue. Etcd,
  > the critical component that powers many cloud native solutions including
  > Kubernetes, could corrupt your data. The issue was so bad, that it required
  > every single administrator to take an action or risk their system becoming
  > unrecoverable. This presentation will discuss what led to the data
  > inconsistency issues, how they were discovered, what was needed to fix them
  > and what lessons we learned that could benefit the whole community."
- Why accepted:
  - Drama opener: "an event that shook the cloud native ecosystem".
  - Maximum stakes — "every single administrator … or risk their system
    becoming unrecoverable". Personal risk = audience can't tune out.
  - Authority: etcd maintainers themselves doing the post-mortem.
  - Promises 4 deliverables: cause, discovery, fix, lessons. Reads as a
    structured incident review, not a marketing pitch.

### 1.2 "Back To Basics: How To Measure Etcd Performance And Not To Die Trying"
- Speaker: David Perez Rodriguez (Gorilla Logic)
- KubeCon NA 2022, Detroit — Lightning talk, Tue Oct 25, 6:10pm
- URL: https://kccncna2022.sched.com/event/184sm
- Abstract (paraphrased from the card):
  > Real-world scenario where an on-prem IoT system processing terabytes
  > suffered "kube-api errors, missed heartbeats, database operators started
  > rolling restarting deployments" because etcd was starved. Covers latency vs
  > throughput, how on-prem hardware degrades vs cloud, and a benchmark
  > playbook to validate any new cluster before it goes live.
- Why accepted:
  - War-story opener with a specific failure cascade — every operator has
    seen "kube-api errors and rolling restarts" in their sleep.
  - Promises a runnable playbook (benchmarks), not just opinions.
  - Title is a meme ("And Not To Die Trying"). Memorable in the schedule grid.
  - Niche fit: on-prem + lightning slot = no committee competition.

### 1.3 "Forging a Stronger Bond Between Etcd and Kubernetes"
- Speakers: Marek Siarkowicz, Wenjia Zhang (Google); James Blair (Red Hat)
- KubeCon NA 2023, Chicago — Wed Nov 8, 3:25pm
- URL: https://kccncna2023.sched.com/event/1R2na
- Why accepted (abstract not retrievable):
  - "Heart-of-the-stack" framing — etcd reframed as the half of Kubernetes
    most users don't think about.
  - Multi-vendor maintainer line-up (Google + Red Hat) signals project
    health, attracts a cross-cutting audience.

### 1.4 "Journey Through Time: Understanding Etcd Revisions and Resource Versions in Kubernetes"
- Speaker: Priyanka Saggu (SUSE)
- KubeCon NA 2023, Chicago — Tue Nov 7, 11:15am
- URL: https://kccncna2023.sched.com/event/1R2lQ
- Why accepted:
  - Educational + mysterious title ("Journey Through Time") on a topic every
    Kubernetes engineer has hit (`resourceVersion` confusion is famous).
  - Solves a concrete debugging headache the audience self-identifies with.

### 1.5 "Insights and Gotchas from the Zero-Downtime Migration of 10000+ Cloud Hosted Etcd Key-Value Stores"
- Speaker: Prabhakar Palanivel (Oracle Corporation)
- KubeCon NA 2023, Chicago — Thu Nov 9, 11:00am
- URL: https://kccncna2023.sched.com/event/1R2qJ
- Why accepted:
  - Number in title — "10000+". Instant credibility.
  - "Zero-downtime" + "Insights and Gotchas" = the audience smells real scars.
  - Vendor-neutral framing (Oracle cloud, but title doesn't push product).

### 1.6 "On the Hunt for Etcd Data Inconsistencies"
- Speaker: Marek Siarkowicz (Google)
- KubeCon EU 2023, Amsterdam — Wed Apr 19, 15:25 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted (abstract not retrievable):
  - Sequel framing to the 2022 incident talk — "Hunt" = active investigation,
    not a victim retrospective. Repeat speaker, brand familiarity.

### 1.7 "Tales from on-Call: Fun with Operating Etcd at Scale"
- Speakers: Geeta Gharpure & Chao Chen (Amazon)
- KubeCon EU 2023, Amsterdam — Wed Apr 19, 16:30 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted (abstract not retrievable):
  - "Tales from on-call" — every SRE in the room recognizes themselves.
  - "Fun" is ironic = audience knows it was not fun. Promises battle stories.
  - "At Scale" + Amazon — implicit AWS-internal etcd fleet, no need to brag.

### 1.8 "Setting up Etcd with Kubernetes to Host Clusters with Thousands of Nodes"
- Speakers: Marcel Zięba (Isovalent), Laurent Bernaille (Datadog)
- KubeCon EU 2023, Amsterdam — Thu Apr 20, 11:00 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted (abstract not retrievable):
  - Concrete scale number ("Thousands of Nodes") in the title.
  - Two-co-speaker pattern from companies with credible scale stories
    (Datadog has been famous for its k8s scaling blog posts).

### 1.9 "Etcd 3.6 and Beyond"
- Speakers: Wenjia Zhang, Marek Siarkowicz, Siyuan Zhang (Google); Benjamin Wang (Red Hat)
- KubeCon EU 2024, Paris — Thu Mar 21, 17:25 CET. Maintainer Track.
- URL: https://kccnceu2024.sched.com/event/1Yhgg
- Abstract (paraphrased): SIG-Etcd maintainers walk through 3.6 priorities —
  bbolt corruption fixes + logging, async storage writes in raft, V2 API
  deprecation, lease enhancements, minor version downgrade, livez/readyz
  probes, and a performance regression test harness.
- Why accepted:
  - Roadmap talk = guaranteed slot for maintainer track.
  - Every line item maps to a real operator pain (downgrade, probes, perf
    regressions). Implicit promise: "your migration pain is on the agenda".

### 1.10 "Scaling and Safeguarding the Heart of Kubernetes: Deep Dive Into etcd"
- Speakers: Wenjia Zhang (Google), Marek Siarkowicz (Google), James Blair (Red Hat), Ivan Valdes Castillo, Wei Fu (Microsoft)
- KubeCon NA 2024, Salt Lake City — Thu Nov 14, 5:25pm. Maintainer Track.
- URL: https://kccncna2024.sched.com/
- Why accepted (abstract not retrievable):
  - "Heart of Kubernetes" — emotional metaphor, easy to remember.
  - 5 speakers from 4 orgs — signals etcd is a healthy multi-vendor project.

### 1.11 "etcd V3.6.0 and etcd-operator V0.1.0"
- Speakers: Benjamin Wang (VMware by Broadcom), Ivan Valdes Castillo (Inmar Intelligence), Siyuan Zhang (Google), Arka Saha (VMware by Broadcom), Ciprian Hacman (Microsoft)
- KubeCon EU 2025, London — Thu Apr 3, 16:00 BST. Maintainer Track.
- URL: https://kccnceu2025.sched.com/
- Why accepted:
  - Two co-shipping releases announced together — turns a routine project
    update into a moment ("0.1.0 of etcd-operator").

### 1.12 "Project Demo: Less Toil, More Features: How Autonomous Testing Freed etcd Maintainers"
- Speaker: Marek Siarkowicz (etcd Tech Lead)
- KubeCon NA 2025, Atlanta — Tue Nov 11, 12:00pm. Project Pavilion.
- URL: https://kccncna2025.sched.com/
- Why accepted:
  - Direct SRE-jargon hook: "Less Toil, More Features". The Google SRE book
    is a shared vocabulary — instant credibility.
  - Counter-intuitive: maintainer talks usually describe more work; this one
    promises less. Curiosity gap = "how?"
  - Concrete deliverable (autonomous testing), not an abstract roadmap.

### 1.13 "Kubernetes and etcd: Common Pitfalls and How To Avoid Them"
- Speakers: Arka Saha & Nabarun Pal (Broadcom)
- KubeCon NA 2025, Atlanta — Tue Nov 11, 4:15pm
- URL: https://kccncna2025.sched.com/
- Why accepted:
  - Classic listicle frame — "Common Pitfalls and How to Avoid Them" is
    a known-good schedule-grid hook (audience self-tests against the list).
  - Vendor-neutral, operator-focused, broad applicability.

### 1.14 "Sponsored Demo: Running etcd in Production: Best Practices & Rescue Recipes"
- KubeCon NA 2025, Atlanta — Wed Nov 12, 3:30pm. Demo Theater (sponsored).
- URL: https://kccncna2025.sched.com/
- Why accepted (sponsored slot, but the framing still matters):
  - "Rescue Recipes" is the hook — implies cookbook-style outputs. The word
    "rescue" presumes the audience has already had a near-miss.

### 1.15 "What's New with etcd, The Backbone of Kubernetes: Project Updates and New Features"
- Speaker: Arka Saha (VMware by Broadcom)
- KubeCon India 2024, Delhi — Thu Dec 12, 11:30am IST
- URL: https://kccncind2024.sched.com/
- Why accepted:
  - "Backbone of Kubernetes" — same heart/backbone metaphor recurring.
  - Maintainer update slot, low-risk acceptance for regional KubeCons.

---

## 2. Operator-focused talks

### 2.1 "Essential Patterns For Designing And Implementing Your Operator"
- Speakers: Michael Hrivnak, Austin Macdonald (Red Hat)
- KubeCon NA 2022, Detroit — Wed Oct 26, 2:30pm. Room 430 AB.
- URL: https://kccncna2022.sched.com/event/182Dc
- Abstract (verbatim):
  > "It's easy to get started developing operators with kubebuilder and
  > operator-sdk to manage your workloads and infrastructure – but what
  > challenges will you face as your operator matures? This presentation will
  > share the most essential lessons learned across years of experience helping
  > teams and organizations design and implement real-world operators for a
  > wide variety of use cases. Coding topics will focus on Go-based operators.
  > You will learn about: - API anti-patterns: Common API design choices that
  > lead to future regret, and how to overcome them in the wild. - Bridging the
  > gap between slow (and complex and buggy) imperative infrastructure
  > management and the declarative Kubernetes API. - Taking control of the
  > client's cache to maximize its usefulness and avoid memory bloat. -
  > Interacting with multiple clusters efficiently from a single operator
  > instance. - Minimizing load imposed on the API server. Attendees will be
  > ready to face key challenges as they enhance their operators with new
  > features and evolving APIs."
- Why accepted:
  - "Easy to get started, but what happens when it matures?" — classic
    bait/payoff structure that admits the audience's current state.
  - Bulleted, scannable list of 5 concrete takeaways — committees love this.
  - Specific anti-patterns named ("future regret", "memory bloat", "load on
    the API server"). Each one is a war wound.

### 2.2 "Beyond Kubebuilder - Generating Entire Kubernetes Controller Implementations"
- Speakers: Amine Hilaly, Jay Pipes (AWS)
- KubeCon NA 2022, Detroit — Wed Oct 26, 3:25pm
- URL: https://kccncna2022.sched.com/event/182Hd
- Abstract (verbatim fragment):
  > "What if you had to build dozens of controllers managing thousands of
  > resources? You'd need a factory to produce full controller implementations
  > from API model schemas."
- Why accepted:
  - "What if you had to …" — provocative question opener, sets stakes.
  - Scale number in the abstract: "dozens of controllers, thousands of
    resources". Authority via AWS Controllers for Kubernetes (ACK).
  - Reframes a familiar tool (kubebuilder) as inadequate — claim worth
    listening to from the team that ships ACK.

### 2.3 "API Evolution With CRDs: Best Practices For Authoring & Fuzz Testing APIs"
- Speakers: James Munnelly, Andrea Tosatto (Apple)
- KubeCon NA 2022, Detroit — Wed Oct 26, 11:00am
- URL: https://kccncna2022.sched.com/event/182HR
- Abstract (verbatim opener):
  > "CustomResourceDefinitions are prolific in Kubernetes. With so many new
  > projects being prototyped, developed and released into the ecosystem, it's
  > essential to ensure you're designing your APIs in a scalable, well tested
  > way."
- Why accepted:
  - Apple speakers + cert-manager pedigree (Munnelly) = unspoken authority.
  - "Fuzz testing" hooks the security/correctness crowd in addition to the
    operator-author crowd.

### 2.4 "Operating CERN SaaS at Scale with Operators"
- Speakers: Michael Hrivnak, Varsha Prasad Narsing (Red Hat); Rajula Vineet Reddy, Francisco Borges Aurindo Barros (CERN)
- KubeCon EU 2023, Amsterdam — Wed Apr 19, 15:25 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted:
  - "CERN" is a marquee user — the moment audiences see CERN in a talk
    title, the answer is yes. Physics labs are unimpeachable proof of scale.

### 2.5 "How to Develop a Robust Operator for Day-2 (Lesson Learned on KubeVirt/HCO)"
- Speaker: Simone Tiraboschi (Red Hat)
- KubeCon EU 2023, Amsterdam — Thu Apr 20, 14:30 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted:
  - "Day-2" is the magic phrase — every operator author is trying to escape
    "happy path Day-1" reviews.
  - Anchored on a real, well-known operator (KubeVirt/HCO) — concrete lessons.

### 2.6 "Story of Our Transition to a Custom Kubernetes Operator for an API Gateway"
- Speaker: Vincent Behar (Ubisoft)
- KubeCon EU 2023, Amsterdam — Thu Apr 20, 11:00 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted:
  - "Story of …" framing = personal narrative, easy to engage.
  - Ubisoft is a recognizable customer; gaming workloads imply scale.

### 2.7 "We Tested and Compared 6 Database Operators. The Results are In!"
- Speakers: Jérôme Petazzoni (Tiny Shell Script LLC), Alexandre Buisine (Enix)
- KubeCon EU 2024, Paris — Thu Mar 21, 11:00 CET
- URL: https://kccnceu2024.sched.com/event/1YeO5
- Abstract (paraphrased): Practical evaluation of CNPG, StackGres, MOCO,
  Zalando, Percona, and one more — "running a production database is more
  complex" than basic deployment; talk covers replication, failover, backups,
  PITR with a "candid yet detailed review" from real customer deployments.
- Why accepted:
  - Number in title ("6") + cliff-hanger payoff ("The Results are In!") =
    rare schedule-grid clickbait that actually delivers.
  - Independent consultants — credibility for "no horse in the race".
  - Petazzoni's Docker-era brand recognition.

---

## 3. Stateful workloads / DB / DR

### 3.1 "Stateful Apps On Kubernetes - Bring Them On!"
- Speakers: Diana Patton, Gunna Marripudi (NetApp); Scott Surovich (HSBC); Scott Miller (DreamWorks); Lisa-Marie Namphy (Cockroach Labs)
- KubeCon NA 2022, Detroit — Thu Oct 27, 3:25pm
- URL: https://kccncna2022.sched.com/event/182EX
- Abstract (paraphrased): Panel with DreamWorks, HSBC, NetApp on "build and
  operate at scale" with stateful apps; topics include state management,
  ecosystem tooling, storage utilization, hybrid/multi-cloud futures.
- Why accepted:
  - Marquee customer panel (HSBC + DreamWorks). Diverse domains = broad appeal.

### 3.2 "Multi-Cluster Stateful Set Migration: A Solution To Upgrade Pain"
- Speakers: Peter Schuurman (Google), Matt Schallert (Chronosphere)
- KubeCon NA 2022, Detroit — Wed Oct 26, 11:55am
- URL: https://kccncna2022.sched.com/event/182It
- Abstract (verbatim):
  > "As more stateful workloads like Redis, Kafka, or custom DBs are migrated
  > to Kubernetes, what operational paradigms need to change to support moving
  > state across clusters and maintaining availability during migration? How do
  > admins safely and reliably perform Day 2 operations and maintenance events
  > while protecting the data and state of the app? What visibility is needed?
  > Today, cluster administrators design complex workflows for data
  > replication, pod and persistent volume migration, and state management for
  > Day 2 ops. What if there was a way to seamlessly migrate StatefulSets
  > between node pools or across clusters to simplify problems related to
  > upgrades, workload migration, and stretching clusters? The speakers will
  > demonstrate the complex patterns developed at Chronosphere to safely
  > migrate stateful workloads to coordinate maintenance operations for
  > thousands of pods across multiple zones and regions. They will then discuss
  > a new enhancement to Kubernetes called StatefulSet Partition which is
  > integrated into a multi-cluster deployment like Chronosphere's and how this
  > can dramatically simplify their operations to focus instead on core
  > business logic."
- Why accepted:
  - Three rhetorical questions in row 1 — sets the problem space, invites the
    audience to nod.
  - "Thousands of pods across multiple zones and regions" — specific scale.
  - Promises Kubernetes upstream enhancement (StatefulSet Partition) =
    forward-looking, not just retrospective.

### 3.3 "Data On Kubernetes, Deploying And Running PostgreSQL And Patterns For Databases In a Kubernetes Cluster"
- Speakers: Chris Milsted (Ondat), Gabriele Bartolini (EDB)
- KubeCon NA 2022, Detroit — Thu Oct 27, 4:30pm
- URL: https://kccncna2022.sched.com/event/182GB
- Abstract (paraphrased): "Databases becoming first class citizens in our
  Kubernetes clusters". CloudNativePG + CSI plugin combo to match
  non-Kubernetes performance/resilience; minimum RTO architecture, replication's
  recovery impact, object storage for PITR against cyber threats, live demo.
- Why accepted:
  - Concrete RTO/PITR vocabulary — DBA audience knows what's being promised.
  - "Live demo" promised — drives attendance.

### 3.4 "Effective Disaster Recovery: The Day We Deleted Production"
- Speakers: Rick Spencer & Wojciech Kocjan (InfluxData)
- KubeCon EU 2022, Valencia
- URL: https://kccnceu2022.sched.com/
- Why accepted:
  - Ultimate war-story title. "The Day We Deleted Production" is impossible
    to scroll past in a schedule grid.

### 3.5 "Availability and Storage Autoscaling of Stateful Workloads on Kubernetes"
- Speaker: Leila Vayghan (Shopify)
- KubeCon EU 2023, Amsterdam — Wed Apr 19, 15:25 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted: Shopify as the implicit scale story; "autoscaling" + "stateful"
  is the unsolved-puzzle hook.

### 3.6 "Scaling Databases at Activision"
- Speakers: Greg Smith, Vladimir Kovacik (Activision/Blizzard)
- KubeCon EU 2023, Amsterdam — Thu Apr 20, 14:30 CEST
- URL: https://kccnceu2023.sched.com/
- Why accepted: Brand-name customer (Activision/Blizzard) + a topic that
  insiders know is hard. The committee assumes the talk has data behind it.

### 3.7 "Scaling Heights: Mastering Postgres Database Vertical Scalability with Kubernetes Storage Magic"
- Speakers: Gabriele Bartolini (EDB), Gari Singh (Google)
- KubeCon EU 2024, Paris — Thu Mar 21, 17:25 CET
- URL: https://kccnceu2024.sched.com/event/1YeM4
- Abstract (paraphrased): Vertical scalability is "an underestimated
  powerhouse, promising a more efficient and high-performing database system".
  Tablespaces, horizontal partitioning, CloudNativePG demos, benchmarks, Day-0
  to Day-2 simplification.
- Why accepted:
  - Counter-intuitive thesis: vertical > horizontal, in a community that
    fetishizes horizontal scale. Curiosity gap.

### 3.8 "One VTOrc To Rule Them All — High Availability In a Distributed Database System"
- Speakers: Deepthi Sigireddi, Manan Gupta (PlanetScale)
- KubeCon NA 2022, Detroit — Thu Oct 27, 11:00am
- URL: https://kccncna2022.sched.com/event/182J8
- Abstract (verbatim):
  > "Vitess is a scalable, highly available distributed database system built
  > around MySQL. … High availability is accomplished through a Vitess feature
  > known as cluster management. The next generation cluster management service
  > in Vitess is called VTOrc. Users can specify their durability rules as a
  > system configuration, which is respected while performing planned
  > failovers. VTOrc also performs failure detection with automatic failovers
  > while honoring the durability rules. VTOrc is already running successfully
  > in production in multiple deployments including at PlanetScale, and it will
  > be Generally Available in Vitess release 15 … followed by a demo of its
  > capabilities showing multiple failover scenarios."
- Why accepted:
  - Tolkien title meme. Demo of "multiple failover scenarios" promised —
    audience expects fire.
  - Anchored on a real GA release date and a real production deployment.

### 3.9 "Kubernetes Workload Resiliency in Action: Beyond Basics"
- Speakers: Nabarun Pal, Akhil Mohan (Broadcom)
- KubeCon India 2026 (Mumbai) — Fri Jun 19, 4:50pm IST. Operations + Performance.
- URL: https://kccncind2026.sched.com/event/2IW4o/
- Abstract (verbatim):
  > "Kubernetes provides powerful mechanisms to protect and isolate workloads,
  > but many teams deploy applications without leveraging these capabilities.
  > In the context of AI workloads, which are often resource-intensive and
  > long-running, maintaining resilience is more critical than ever.
  > This talk explores Kubernetes workload protection mechanisms and
  > demonstrates how to combine them strategically for maximum resilience.
  > We'll examine resource requests/limits, priority classes, resource quotas,
  > runtime classes, and strategies like pod disruption budgets and affinity
  > rules. Through practical examples and real-world scenarios, you'll learn
  > how to configure each mechanism, understand their interactions, and avoid
  > common pitfalls that lead to pod evictions, performance degradation, and
  > cascading failures.
  > Whether you're running mission-critical services or shared multi-tenant
  > clusters, this session will equip you with a resilience framework that
  > protects your workloads under pressure."
- Why accepted:
  - "Beyond Basics" — admits the audience already knows the basics, promises
    a step up. Anti-intro-talk hook.
  - Hooks AI workloads (2026 hype-cycle compliance) without being an AI talk.
  - Concrete shopping list of K8s primitives — committee sees a structured,
    teachable talk.
  - "Cascading failures" closes the loop with stakes.

### 3.10 "From Chaos to Control: Building Resilience with Effective Backup and Disaster Recovery in Kubernetes"
- Speaker: Yash Pimple (Independent)
- KubeCon India 2024, Delhi — Thu Dec 12, 12:20pm IST
- URL: https://kccncind2024.sched.com/
- Why accepted:
  - "From Chaos to Control" — narrative arc title; auditorium-friendly.
  - Independent speaker — regional KubeCons reward community contributors.

---

## 4. Data protection / backup-restore (committee-track)

### 4.1 "Kubernetes Data Protection WG Deep Dive"
- KubeCon NA 2022 / EU 2024 / EU 2025 — recurring slot, Xiangqian Yu (Google) /
  Dave Smith-Uchida (Veeam) variously.
- Why accepted: WG status talk; selection bar is low but the recurring slot
  means there is reliable audience demand for the topic.

### 4.2 "Kubernetes Backup Legitimized: CSI Changed Block Tracking Has Arrived"
- Speakers: Mark Lavi, Carl Braganza, Prasad Ghangal (Veeam); Xing Yang (VMware/Broadcom)
- KubeCon EU 2025, London — Thu Apr 3, 17:00 BST
- URL: https://kccnceu2025.sched.com/
- Why accepted:
  - "Legitimized" framing — concedes a long-running criticism (k8s backups
    were second-class) and announces the win.
  - "Has Arrived" = news. Time-sensitive => committee likes timeliness.

---

## 5. Patterns that work — distilled

Below are the rhetorical templates that recur across the 25+ accepted
abstracts above. For each, evidence is cited by section number.

### Pattern A — "The Day We Almost Lost Everything" (battle story)
The talk opens with a specific failure or near-miss. Stakes are personal:
"every administrator had to take action or risk their system becoming
unrecoverable" (§1.1). The audience can't tune out because they have lived
the same outage in a different uniform.
- Evidence: §1.1 etcd data inconsistency; §1.2 IoT kube-api meltdown; §3.4
  "The Day We Deleted Production"; §1.7 "Tales from on-call".
- For our talk: lead with a concrete etcd disaster (snapshot corruption,
  quorum loss, restore-from-backup at 3am).

### Pattern B — "Numbers in the title" (scale credibility)
A specific quantity in the title or first line buys the committee's
attention before the dot. Vagueness loses; "10000+ stores", "thousands of
nodes", "dozens of controllers managing thousands of resources", "6
database operators" — all in the corpus.
- Evidence: §1.5 (10000+); §1.8 (thousands of nodes); §2.2 (dozens/thousands);
  §2.7 (6 operators); §1.11 (3.6.0 + 0.1.0 dual-release).
- For our talk: lead the abstract with "N etcd clusters", "M GB snapshots",
  "P-second RTO", or the size of the Gardener fleet.

### Pattern C — "Day-2 reckoning" (escape the demo)
The committee is tired of Day-1 happy-path demos. Talks that explicitly
position themselves as Day-2 / "what happens when it matures" get a head
start.
- Evidence: §2.1 ("easy to get started … but what challenges will you face
  as your operator matures?"); §2.5 ("Day-2 … Lesson Learned on KubeVirt/HCO");
  §3.2 ("Day 2 operations and maintenance"); §3.10 (Chaos → Control).
- For our talk: frame etcd-druid as the Day-2 operator that picks up where
  the StatefulSet-and-PVC tutorials stop.

### Pattern D — "Counter-intuitive thesis" (curiosity gap)
A claim that contradicts the audience's default assumption.
- Evidence: §3.7 "vertical > horizontal" Postgres scaling; §1.12 "Less Toil,
  More Features" (more autonomy = less work for maintainers); §2.2 "Beyond
  Kubebuilder" (the canonical scaffold isn't enough).
- For our talk: a candidate is "Most etcd outages are not etcd bugs" or
  "The hardest part of running etcd isn't running etcd".

### Pattern E — "Listicle with stakes" (the bulleted promise)
The abstract explicitly enumerates 3–5 takeaways. Committees love
scannability; the audience self-tests against the list.
- Evidence: §2.1 (5 named anti-patterns); §1.2 (latency/throughput/on-prem
  trio); §1.13 ("Common Pitfalls"); §1.9 (3.6 feature bulletlist).
- For our talk: include a numbered list of pitfalls (compaction storms,
  snapshot timing, leader flapping, AZ-fault tolerance, defrag windows).

### Pattern F — "Named users, real production" (authority by proxy)
Mentioning a marquee user moves the abstract from speculation to evidence.
- Evidence: §2.4 (CERN); §2.6 (Ubisoft); §3.1 (HSBC + DreamWorks); §3.5
  (Shopify); §3.6 (Activision/Blizzard); §1.7 (Amazon at scale).
- For our talk: name Gardener and the SAP Cloud customers if possible —
  thousands of control planes managed by etcd-druid is the moat.

### Pattern G — "Live demo / cookbook" (concrete artifact)
Abstracts that promise a runnable artifact — benchmark, runbook, recipe —
out-compete pure narrative talks for the operations track.
- Evidence: §1.2 (benchmark playbook); §1.14 ("Rescue Recipes"); §3.3 (live
  demo of PITR); §3.8 (multiple failover scenarios demoed).
- For our talk: promise a chaos scenario run live — kill a member, lose a
  zone, restore from object storage, in <30s of wall-clock.

---

## Top-of-mind closing observation

Across 4 years, **the accepted etcd talks fall into three buckets**:
1. **Maintainer roadmap** (§1.9, §1.10, §1.11, §1.12) — always accepted,
   driven by the project, not the speaker.
2. **Operator pain at scale** (§1.5, §1.7, §1.8, §1.13) — needs a number, a
   user, and a war wound. This is the bucket etcd-druid fits.
3. **Incident retrospective** (§1.1, §1.6) — needs a public, dramatic
   failure; harder to manufacture, easy to overdo.

An etcd-druid talk should aim for bucket 2 with Pattern A (battle story)
opening, Pattern B (numbers), Pattern C (Day-2 reckoning), Pattern F
(Gardener as named user), and Pattern G (live recovery demo) as the spine.
