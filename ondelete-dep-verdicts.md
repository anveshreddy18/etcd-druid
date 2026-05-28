# PR #879 — unmarshall review: clustered analysis & verdict slots

37 unmarshall-touched threads grouped into 11 themes. Order = priority for addressing.

**Notation:** ❓ = needs your verdict. Write your take under "Verdict" for each item.

---

## CLUSTER A — 🔴 FOUNDATIONAL: "UpdateStrategy doesn't belong in the Etcd API"

Single decision that gates ~8 threads. Unmarshall's repeated thesis: StatefulSet is an internal implementation detail; the user only deals with the `Etcd` resource; exposing `updateStrategy` leaks that detail.

| # | Thread | Unmarshall's point | Current state |
|---|---|---|---|
| A1 | **L42** (5 comments) | "Which user decision do you hint at? The user has no choice in the matter today w.r.t which update strategy is chosen by etcd-druid internally." Anvesh: "I guess we need a discussion on that. Wdyt?" | **Discussion pending** |
| A2 | **L47** (Goal: introduce spec field) | "Why should this be done? This exposes internal implementation detail... If tomorrow we move away from STS these strategies might no longer make sense. In which situations would you choose RollingUpdate vs OnDelete?" | No reply yet |
| A3 | **L48** (Goal: propagate STS template changes) | "Not sure how is this a goal? StatefulSet is an internal implementation detail. The user only deals with `Etcd` resource." | No reply yet |
| A4 | **L58** (Proposal section: Update Strategy as Etcd Spec Field) | "As mentioned earlier - please provide reasons to do that." | No reply yet |
| A5 | **L95** (Per-cluster choice sentence) | "There is no transition for an end-user. etcd-druid hides the implementation detail and only provides a high level capability to update an etcd-cluster..." | No reply yet |
| A6 | **L97** (Feature-gate vs spec-field comparison) | "Why is a feature gate option relevant for comparison? Not clear." | No reply yet |
| A7 | **L116** (Coordination paragraph) | "This is completely unnecessary. It stems from the same observation that exposing this field in the API is pointless from my perspective." | No reply yet |
| A8 | **L231** (Transitioning Between Strategies section) | "This is not even required. This entire section IMHO is not required." | No reply yet |

**Decision to make:** Do we keep `spec.updateStrategy` as a public API field, or remove it and make the choice internal?

- **If REMOVED** → A1-A8 resolve, and Cluster B (separate controller) gets simpler too. Proposal sections to delete/rewrite: Goals #2, "Update Strategy as an Etcd Spec Field", "Transitioning Between Strategies", "Coordination with the Etcd reconciler".
- **If KEPT** → Need stronger justification (concrete user scenarios where the choice matters) and replies clarifying internal/external boundary.

**Verdict:**
> We have concluded that we want to go with a feature gate instead of introducing the updatestrategy in the Etcd spec field, the reason being that currently it leaks the implementation details of how we realize and manage the cluster and the updatestrategy is very statefulset specific, tomorrow we might remove the statefulset entirely then this field won't make sense and we won't have the choice to remove this from the API then, then it'll become all messier. So I agree with his points about not making it a spec field and I proposed the feature-gate way instead and he said that's the right thing to do -- To give you more information on how the feature gate can be added here: go through this https://github.com/anveshreddy18/etcd-druid/blob/a6d601ac93b1b565a40497d7ed995a332d4fd210/docs/proposals/06-sts-ondelete-strategy.md file which is a previous version of the DEP wherein I was using feature gate, take it as an inspiration about how to put forth the feature gate but also check the etcd-druid codebase on how a feature gate is created and managed here for this usecase and what happens when this feature gate is enabled or disabled and when the druid version is bumped & reverted from older version to this implementation version and vice versa respectively. You already have a good amount of understanding of how the strategy switch works as I've outlined in the proposal but the current one in the DEP ties it with etcd spec update but since we will no longer be having that spec field, it's all with this feature gate now, so we need to explain in detail the various ways of switch between strategies that can happen ( the meaningful ones ) and how our DEP deals with that. And without saying, ofcourse we need to remove the mentioning of the spec field whereever possible. I can understanding that this disturbs the DEP structure a bit, but we need to make sure that we're not making just patches in place of previous ones and we should think and figure out where is the best place to write about different kind of stuff so that the readability is great. Now when you go through the code of the previous version of file that I shared with you here that talks about feature gate, I don't want to pollute your context of other things present in that file because those could be outdated/wrong, so for reading that file you might want to spin up an separate agent to read that. Also read about feature gates semantics from this k8s documentation specifically https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/#feature-stages as it explains in clear what the semantics are in the k8s area. With this I believe points A1, A2, A4, A5 should be addressed stragight away, and for A3 we should remove the line "Ensure that updates triggered by changes to the StatefulSet pod template (image versions, configuration, resource requests) are propagated to pods under druid's control." from our goals section, for A6 - well we don't need the feature gate comparision as we're going to use feature gate, at the same time we don't need a comparision of feature gate with the spec field as well ( because spec field doesn't make sense) but we should still justify why we're going with the feature gate, what it solves for us - which is to let us disable the feature and use RollingUpdate for instance if we see any issues with the newly introduce OnDelete mechanism. Also whenever we enable/disable feature gate - we should make sure that the right updateStrategy reflects in the sts always, this is a must. This is also applicable in the case of version upgrade/downgrade, we might want to mention these in the relevant place so that we acknowledge these and don't forget them. And for A7 - he's right, we need to change it to suit our new ways. And for A8 - no we should still have this type of information as he also agreed which explains how things behave when switching between strategies and when does that happen ( as I've explained already so far ), so we need to have better points, maybe a better place or something ( don't try to force it, if it's naturally well placed already we will just keep it in that place )... but you get my point right? Let's go.

---

## CLUSTER B — 🔴 FOUNDATIONAL: "Why a separate controller?"

Related to A but somewhat independent: even if the field is kept, unmarshall thinks the OnDelete update logic belongs in the existing Etcd reconciler.

| # | Thread | Unmarshall's point |
|---|---|---|
| B1 | **L107** (Bullet: "Pod-by-pod waits do not block the reconciler") | "In the etcd reconciler we no longer block, we always requeue. So blocking or tie up of threads is not really a concern. If many clusters are getting updated simultaneously then you will just transfer the same issue to the new reconciler. Also IMHO having 2 reconcilers which are going to reconcile a single `Etcd` resource is not a k8s recommended way as it can lead to race conditions... There is the other reconciler that will have to anyways do that work. It is not only STS there are other components like ConfigMap that might also need changes, would you then duplicate the entire set of components for another reconciler?" |
| B2 | **L108** (Bullet: "Separation of concerns") | "etcd reconciler ensures that you move towards a desired state. Even today you can have errors when calling `Sync` on all managed components. For these errors we requeue. So where is the difference? Updates are intrinsic part of the same reconciler because it defines a new desired state that you need to reconcile to." |
| B3 | **L118** (Strategy switch coordinated by sequencing) | "So for rollingupdates, you will continue to use the existing etcd reconciler and only for on-delete you will delegate to a dedicated controller. This responsibility boundary is not well defined and as stated above is not preferred. I do not understand the need for doing this." |
| B4 | **L146** (Shreyas/Anvesh/Unmarshall) | Unmarshall to Shreyas: "How are you blocking a reconciler thread? A quick requeue is not a block - it is the opposite of blocking. Can you please explain?" |

**Decision to make:** Single reconciler vs separate OnDelete controller. (Note: depends on Cluster A outcome — if A field is removed, the OnDelete logic is the only reason a separate controller exists at all.)

**Verdict:**
> While in overall he agreed in the call to having a separate controller do the operations as I've proposed today but he wanted me to be clear with the reasons on why I think that's the best, he thinks the current points lack weight and he asked me not to compare the reconciler threads, no.of requeues stuff because apparently that is almost the same with either choosing to do the work the Statefulset component of the etcd reconciler or the separate ondelete controller. And he recommended that my main points are these two That with choosing a separate controller we don't touch the existing mechanisms of etcd controller, as it keeps doing what it was doing before i.e propagating the required information to the k8s resources instead of acting on behalf of the resources also, so we will keep that behaviour intact so there won't be any code changes to that, so the ondelete controller can evolve independently and also that's the way it's happening today with the sts controller as well because etcd controller's main job is to propages the changes to sts and it's the responsibility of the sts controller to take that in and manage the pods, so by having a separate ondelete controller, we still respect that for separation of concerns. ( Don't write it like I mentioned, you need to think through and come up with crisp points that cannot be challenged once more ). For B, B2 - doing what I said will address his points. For B3 - ignore that point. For B4 - Ignore this as well as it's question to Shreyas and anyway it won't be standing still after we make these changes. Let's go and make the changes!

---

## CLUSTER C — 🟢 EDITORIAL WINS (quick): suggestion edits & terminology

| # | Thread | Status |
|---|---|---|
| C1 | **L16** Summary rewrite | ✅ Already applied (Anvesh's local edit, uncommitted) |
| C2 | **L25a** "can also be a learner, right?" (re: Follower definition) | Anvesh replied: "learner falls into non-participating category". Awaiting unmarshall to acknowledge/resolve. |
| C3 | **L25b** Participating/Non-Participating wording | ✅ Already accepted in working copy |
| C4 | **L32** Motivation paragraph rewrite | ✅ Already applied. Unmarshall added a 2nd comment about CaptainIRS's two motivations (reduce transient quorum loss + reduce leadership transfers) — this might need a one-line addition or already implicitly covered. |

**Verdict on C4 (the "two motivations" framing):**
> So the C1, C3 changes are previously added but now they're in the text.diff - you can add them back. For C2 - we ignore since that is answered. For C4 - the paragraph rewrite is done but unmarshall added a comment explaining the motivations, now apart from those two he mentioned, there's one more - that with OnDelete, the PVC resizing story becomes more safer, more like the OnDelete strategy makes PVC resizing safer (see [Interaction with PVC Resizing](#interaction-with-pvc-resizing)), but the PVC resize flow itself is a separate feature. So I want you to hint that OnDelete helps PVC resizing but it'll be covered in a different issues/proposal as it's already written somewhere. We need to just streamline it and make it consistent in the proposal. Also in the non goals we already say that "This proposal does not cover PVC resizing or volume replacement." -- we just need to be more explicit on it. Let's go.

---

## CLUSTER D — 🟡 INDEPENDENT: Goals & Non-Goals framing (3 threads)

| # | Thread | Unmarshall's point | Your earlier take? |
|---|---|---|---|
| D1 | **L52** ("poorly-ordered voluntary pod updates") | "These usage of words reflects that StatefulSet controller made a bad judgement call. But this is not true... It is the poor judgement on our part to depend upon a defined update strategy which can potentially cause transient quorum loss. For nonHA this is fine but for HA it can violate the SLA for availability during updates." | — |
| D2 | **L53** (PVC resizing as non-goal) | "To do volume resizing, you will have to use `OnDelete` update strategy. So this use case becomes a motivation which you should explicitly mention in the `Motivation` section. However detailing on how to achieve volume resizing is out of scope for this DEP and should be covered in a separate DEP." | — |
| D3 | **L54** (">3 replicas optimization" as non-goal) | "An update strategy is for an STS of any size >=1. How can this be a non-goal? When you write a DEP for changing the update strategy you envision it in entirety. You can say iteration-1 vs iteration-2 of implementation, but this cannot be a non-goal." | — |

**Verdict (per item):**
- D1: _accept reword / not correct / other_
- D2: _accept move to motivation / not correct / other_
- D3: _reframe as scope phasing / disagree / other_

---

## CLUSTER E — 🟡 INDEPENDENT: Health assessment clarifications (3 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| E1 | **L132** (readiness probe = quorum participation) | "It is possible that the etcd-member can serve traffic but there is no quorum, thus negating this argument." Provided a rewrite suggesting kubelet probes pod IP directly, linearizable Get against local etcd, etc. |
| E2 | **L140** ("controller leaves them untouched") | "What does that mean precisely? Also a container can remain stuck in `ContainerCreating` state (CSI driver fails to attach PV). Per your statement it will leave these pods as `untouched` — so the update is stuck as well." |
| E3 | **L142** (use container status now, /livez later) | "Above you mentioned distinguishing 'process dead' vs 'alive but quorum lost' allows better decisions. But here you say controller works without the probe. Do you recommend in this DEP to use /livez or not? A bit unclear." |

**Verdict (per item):**
- E1: _accept his suggested rewrite / partially / keep current_
- E2: _add explicit handling for ContainerCreating stuck case / treat as "dead" / clarify "untouched"_
- E3: _make recommendation explicit (current = container-status, future = /livez)_

---

## CLUSTER F — 🟡 INDEPENDENT: Leader handling edge cases (2 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| F1 | **L170** (read role from member lease) | Race condition: leadership change at T1, member lease reflects at T3, druid picks at T2 (T1<T2<T3) using stale lease → picks new leader thinking it's a follower → causes another leader election. |
| F2 | **L172** (leadership transfer on deletion) | "While this is true, you must state clearly that transfer of leadership as done from within the etcd code is `best-effort`. It does not wait for the leadership to be transferred successfully. This helps the reviewer understand what is the trade-off." |

**Verdict (per item):**
- F1: _acknowledge race + mitigate (re-check role just before delete?) / accept as known limitation / other_
- F2: _add best-effort caveat to the existing note_

---

## CLUSTER G — 🟡 INDEPENDENT: Stuck-pod / parallel updates / quorum-loss handling (4 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| G1 | **L146 (Vlatko-orig)** (long, multi-author with Ishan + Shreyas) | Last turn from unmarshall: "If you have a 5-member cluster all healthy, you have room of 2 without losing quorum. If the first pod blocks, should we stop or check 'do I have sufficient room to try another Pod' and proceed? Should we play extremely conservative and stop at the first failure or go to where quorum is still maintained?" Anvesh reply (May 27): "we would still want to be conservative." |
| G2 | **L146c** (Step 1 procedure) | "Let us consider quorum already lost. 2/3 unhealthy. Would it make sense to delete both and recreate both together in one shot instead of one at a time at all times?" |
| G3 | **L161** (Delete the selected pod and requeue) | "This was precisely my question above. Should we have something similar to `maxUnavailable` but re-interpreted in the context of quorum based workload updates? Of course you will not call this field with this name." |
| G4 | **L274** (Future scope: concurrent updates for >3) | "Why is this a future scope and why do we assume that the cluster size is always 3 for a HA etcd cluster?" |

**These four are one theme:** parallel/batch deletion semantics, especially when there's slack (>3 members) or quorum already lost. Anvesh leans conservative; unmarshall is pushing to think bigger and not punt to future scope.

**Verdict:**
> _stay conservative + acknowledge in proposal / explore quorum-aware "maxConcurrentDeletes" knob in v1 / fold the >3 work back into goal scope / other_

---

## CLUSTER H — 🟡 INDEPENDENT: Status / in-flight pod tracking (2 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| H1 | **L150** (Step 1: compare each pod's controller-revision-hash) | "Can you check if `StatefulSet.Status.UpdatedReplicas` can be used with `OnDelete`? If we can rely on that it's cheaper than a LIST call. Note: #73492 only talks about CurrentRevision; #136833 fixes it." |
| H2 | **L152** (CaptainIRS + Anvesh + unmarshall) | "What CaptainIRS highlighted is valid — you need to record the decision from the previous reconciliation. If a Pod is selected for update and in the next reconciliation you must have a deterministic way to check if there is an `InFlightPod` selected in the previous cycle. This should survive crash of etcd-druid operator, therefore preserved in `Etcd.Status`. Can you detail the offline discussions that mark this as a mere implementation detail?" |

**These together challenge the "stateless controller" framing:**

**Verdict:**
- H1: _investigate UpdatedReplicas usability under OnDelete with the upstream fix + document_
- H2: _add `InFlightPod` to Etcd.Status / argue against (with reasons) / other_

---

## CLUSTER I — 🟡 INDEPENDENT: Pod deletion / PDB clarifications (3 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| I1 | **L186** ("direct Delete") | "Clarify what `direct Delete` is" + provided a precise rephrasing (pod DELETE endpoint vs `pods/eviction` subresource, no PDB consultation). |
| I2 | **L196** (option 3: AlwaysAllow weakens protection) | "In Gardener for etcd cluster `AlwaysAllow` policy is used [link to gardener source]. So your argument that it weakens PDB protection is not entirely true. You must understand the reasons why PDB restriction was relaxed for etcd clusters in gardener so you can re-evaluate or add supporting documentation." |
| I3 | **L202** ("RollingUpdate already uses Delete") | "Always provide proof of your claim" + provided rewrite with concrete link to `stateful_pod_control.go`. |

**Verdict (per item):**
- I1: _apply his rephrasing_
- I2: _re-evaluate AlwaysAllow argument / cite Gardener's reason and update / counter-argue_
- I3: _apply his rephrasing with the link_

---

## CLUSTER J — 🟡 INDEPENDENT: Diagrams (2 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| J1 | **L34** (RollingUpdate state diagram) | "The diagram can actually be removed. (1) it is not very clear (2) you have already explained that briefly above so there is no real need to pictorially represent that." |
| J2 | **L178** (OnDelete state diagram) | "This diagram is not complete and I would remove it. Example: 'Any non-participating outdated pods?' in the yes branch leads to 'Is etcd process dead?' — of which pod? Since you have explained the flow clearly above I think this diagram is just dead weight." |

**Verdict:**
> _remove both / keep both / fix and keep / other_

---

## CLUSTER K — 🟡 INDEPENDENT: Metrics (2 threads)

| # | Thread | Unmarshall's point |
|---|---|---|
| K1 | **L268** (etcddruid_ondelete_update_duration_seconds) | "Why specific to ondelete? There needs to be a single metric to compute the total time to update an etcd cluster. If you want extra info about strategy used, have it as a label/tag." |
| K2 | **L269** (etcddruid_ondelete_reconcile_cycles_total) | "Can you explain how and where will you be using this metric when you have already proposed to capture the total time to update an etcd cluster?" Anvesh replied (May 27) with rationale about quantifying retry cadence. |

**Verdict:**
- K1: _rename to generic with `strategy` label / keep as ondelete-specific / drop_
- K2: _keep with explanation already given / drop / merge into K1_

---

## Summary by priority

1. **Decide Cluster A** — unblocks A1-A8 + simplifies B.
2. **Decide Cluster B** — possibly trivial after A.
3. **Apply Cluster C** — fast wins, already mostly done.
4. **Cluster D** — Goals/Non-Goals reframing.
5. **Cluster E** — Health assessment clarifications.
6. **Cluster F** — Leader race + best-effort caveat.
7. **Cluster G** — Stuck pod / parallel updates (open design question).
8. **Cluster H** — Status / in-flight pod tracking (mini-design question).
9. **Cluster I** — Pod deletion / PDB editorial.
10. **Cluster J** — Diagrams remove/keep.
11. **Cluster K** — Metrics generic vs ondelete-specific.
