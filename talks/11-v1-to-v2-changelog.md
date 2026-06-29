# V1 → V2 changelog

## Consensus fixes applied

1. **Cut the warm-up first sentence "Upstream etcd is excellent software."**
   - Critics: track-curator (FIX #1, explicit), skeptical-engineer (cliché list, "borderline cliché … keep, but tighten").
   - Action: Moved this phrase out of the lead position. The four-pain catalogue now opens; "Upstream etcd is excellent software" survives only as the connective at the end of paragraph 1, where it acts as the pivot from pain to thesis rather than as a warm-up.

2. **Promote and harden specific failure modes in the opener (real error strings, real issue references).**
   - Critics: skeptical-engineer (3 AM smell test — "one real error string would lift this from 8 to 9", cite the `#11555` hot-reload issue), track-curator (Pattern A: more lived-in detail).
   - Action: Replaced the generic "certificate rotation … yearly planned outage" with "the CA bundle etcd refuses to hot-reload, turning cert rotation into a yearly planned outage" (encodes #11555 without the issue number cluttering the prose). Replaced the generic "permanent quorum loss" sentence with "the `mvcc: database space exceeded` that flips kube-apiserver read-only at 3 AM" (concrete error string from the problem-solution map).

3. **Replace "etcd-druid and etcd-backup-restore are an open-source answer to that gap."**
   - Critics: community-impact (FIX #2: project-pitch pivot too early), skeptical-engineer (Fix #1: corporate-blog phrasing, name the competition).
   - Action: New paragraph-2 opener: "This talk walks through what an etcd operator actually has to do, using etcd-druid and etcd-backup-restore as the worked example." Reframes the projects as illustrations, not the thesis.

4. **Add a number + named competitor.**
   - Critics: track-curator (FIX #2: Pattern B numbers), skeptical-engineer (Fix 2: "tens of thousands" lives in the problem-solution map, use it; Fix 1: name CoreOS/Zalando).
   - Action: "Tens of thousands of clusters" now in paragraph 2. "Since 2018" replaces "over eight years" (specificity > duration, per skeptical-engineer). The CoreOS etcd-operator archival is named at the end of paragraph 1: "where the CoreOS etcd-operator stopped answering when it was archived in 2020."

5. **Cut the Linux Foundation NeoNephos line.**
   - Critics: community-impact (FIX #1: pure positioning, zero audience benefit), track-curator (community-benefit ratio commentary: "could be cut … to push the ratio above 70%").
   - Action: Removed entirely. Customer names (Akamai, StackIT, Scaleway) retained because track-curator scored them as Pattern F evidence, not promotion.

6. **Replace the generic "you will leave with a concrete picture …" closer.**
   - Critics: all three. Community-impact (FIX #3: "concrete picture" is itself abstract), track-curator (FIX #3: name takeaways), skeptical-engineer (Fix 3: kill the boilerplate, add a "what we don't solve" line).
   - Action: New closer takes the skeptical-engineer rewrite as the spine — "which failure modes the operator pattern eliminates, which ones it merely contains, and which ones … still live in your runbook" — and names two specific unsolved items (permanent quorum loss, 8 GB ceiling). This single sentence carries the failure-mode honesty the V1 was missing.

7. **Promote the compaction-as-restore-validation mechanism.**
   - Critics: community-impact (KEEP + promote earlier), track-curator (KEEP, "promote it harder").
   - Action: It now sits as the climax of paragraph 2's four-mechanism enumeration, immediately before the credibility line. "The interesting part:" workshop filler removed per skeptical-engineer's cliché note.

## Disagreements resolved (audience-benefit tiebreaker)

### Disagreement A — Keep customer logos?
- community-impact wanted the entire SAP/Akamai/StackIT/Scaleway+NeoNephos block cut as "credibility-as-brag, zero audience benefit."
- track-curator scored the same block as a strength: Pattern B (numbers) + Pattern F (named users).
- skeptical-engineer agreed with track-curator: "alongside Akamai, StackIT and Scaleway — good (named users, Pattern F), not cliché. Keep."
- **Tiebreaker — non-adopting reader test:** A reader who will never install druid still benefits from knowing this isn't a single-vendor toy. Pattern F is the difference between "trust us" and "trust these other teams who also bet on it." Vote 2-1 + tiebreaker both point the same way.
- **Resolution:** Keep the named users. Drop NeoNephos (no equivalent audience benefit — governance status doesn't change what the reader does Monday morning).

### Disagreement B — Mention the archived CoreOS etcd-operator?
- skeptical-engineer asked for it explicitly: "the unspoken Zalando elephant."
- community-impact and track-curator did not request it.
- **Tiebreaker — non-adopting reader test:** A reader who tried etcd-operator three years ago and gave up gets immediate context for why a new operator exists. A reader who never tried it learns the historical shape of the gap. Both populations benefit.
- **Resolution:** Include, but compress to one clause ("where the CoreOS etcd-operator stopped answering when it was archived in 2020") instead of skeptical-engineer's longer rewrite. The longer version would have pushed the abstract over the word budget without proportional benefit.

### Disagreement C — Bulleted enumerated takeaways vs prose closer?
- track-curator wanted a numbered list "(1) why every compaction job should double as a restore drill, (2) …" (Pattern E).
- skeptical-engineer wanted a prose closer naming what's NOT solved.
- community-impact wanted a count or a checklist ("seven things a production etcd setup needs").
- **Tiebreaker — non-adopting reader test:** A bulleted list invites pattern-matching; a "what we don't solve" line invites trust calibration. The corpus §2.1 uses a bulleted list — but the V1 already lists four pains in paragraph 1, so adding another list at the end would be repetitive. The honest-scope move is what V1 most lacks. Skeptical-engineer's closer wins on marginal-information grounds.
- **Resolution:** Prose closer that names two specific unsolved items (permanent quorum loss, 8 GB ceiling). This satisfies community-impact's "specificity" requirement and skeptical-engineer's "name what you don't solve" without re-listing what V1 already lists.

### Disagreement D — Open with the four-pain catalogue, or with the "If you have ever wondered …" rewrite?
- community-impact proposed a new "If you have ever wondered whether the etcd snapshot in your S3 bucket would actually restore …" opener.
- track-curator + skeptical-engineer both kept the four-pain rhythm as the strongest paragraph in the V1 ("KEEP" list of both).
- **Tiebreaker — non-adopting reader test:** The four-pain catalogue gives four self-identification triggers in one paragraph; the "wondered whether the snapshot would restore" opener gives one. More hooks = higher chance a given reader finds themselves.
- **Resolution:** Keep the four-pain catalogue, with the upgrades listed in Consensus fix #2.

## What's preserved (union of "KEEP" lists across all three reviewers)

- Title: "The hardest part of running etcd is not running etcd" (track-curator: do not touch).
- The four-pain catalogue opening rhythm (community-impact + track-curator + skeptical-engineer).
- The compaction-as-restore-validation sentence as the centerpiece mechanism (all three).
- Named users (Akamai, StackIT, Scaleway) as Pattern F evidence (track-curator + skeptical-engineer).

## Verification results

| Check | Result |
|---|---|
| Em-dashes (`—`) in body | **0** (verified by `grep -c "—"`) |
| Word count (body only, excluding title) | **199** (target 160–200) |
| You / your / yours references | **3** |
| We / our / us references | **0** |
| Audience-benefit ratio | 3 : 0 (≫ 1.5× target; first-person plural eliminated entirely) |
| Cliché scan (`seamlessly`, `robust`, `industry-leading`, `cutting-edge`, `battle-tested`, `self-healing`, `world-class`, `next-generation`, `transform`, `revolutionize`, `leverage`, `unparalleled`, `comprehensive`, `holistic`) | **none present** |
| Failure-mode honesty location | Paragraph 3, final sentence: "which ones it merely contains, and which ones, like permanent quorum loss and the 8 GB ceiling, still live in your runbook." Names two specific unsolved categories. |
| Corpus voice match (§1.1 / §3.2 / §2.1) | Present-tense descriptive register; sentence rhythm mirrors §2.1 (problem framing → mechanism enumeration → calibrated takeaway); adjective density low; no bulleted block inside the abstract body. |
