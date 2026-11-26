---
name: review-loop
description: Review a change with two fresh independent reviewers per round, fix every P0 finding and the P1–P2 findings both reviewers report, and repeat in batches of at most 3 rounds, asking the user before each additional batch. Use when asked to review changes, or when another skill calls review-loop.
argument-hint: "[target: uncommitted | <commit range>] [slice: <slice of spec.md>]"
---

# Review loop

Inputs:

- target: the uncommitted change set (staged, unstaged, untracked; in jj, the working copy against its parent) or a commit range.
- slice: optional. The slice of `spec.md` under review, when the caller implements one slice at a time.

Pass nothing else to the reviewers. What the change must do, its scope, goals, non-goals and guardrails, is in `spec.md` or `rca.md` at the project root, and each reviewer works it out from there.

## Round limit

Run at most 3 rounds initially. One round is two reviewers' passes, matching their findings, and fixing the accepted ones. Stop earlier once a round has no accepted findings.

If another round is needed after round 3, pause and report the rounds completed and the accepted findings still open. Ask: "Continue for up to 3 more review rounds?" Wait for explicit user approval. Each approval authorizes at most 3 additional rounds; ask again at the next boundary (6, 9, and so on). Approval of a fix, a design change or the overall task is not approval for extra rounds.

Carry the round count across resumptions and caller hand-offs for the same review; invoking this skill again must not bypass the limit. A new target starts a new count. While paused, do not declare the review complete or continue caller steps that need a completed review. If the user declines, report the incomplete review and its open findings. If the caller is unattended and cannot ask, stop and include the continuation question in its final report instead.

## Each round

### 1. Two independent reviews

Launch two fresh subagents in parallel with the same prompt. Never reuse a reviewer across rounds and never show one reviewer the other's output.

```text
Review <target> in <repository root>. Work read-only; run probes in a temporary directory outside the repository.
<slice given: "Only slice <slice> of spec.md is under review. Tests still listed in its pending skip-list block belong to later slices and are not gaps.">

1. Scope. Before reviewing, read `spec.md` or `rca.md` at the repository root. If neither exists, use the commit messages and the issue or PR they reference. State:
   - the problem the change solves;
   - its goals, non-goals and guardrails;
   - the slice under review, if any;
   - what the change does.
   Judge the change only against this scope. Check both directions: what the scope requires that the change does not do, and behaviour the change adds that the scope does not ask for.

2. Review against, in priority order, naming the ones you used:
   1. The guardrails in `spec.md` or `rca.md`. Any violation is P0–P2 regardless of its impact.
   2. Project guidelines: any `*-review-guidelines` skill for this repository, and the repository's AGENTS.md or CLAUDE.md.
   3. The `review-oracle` skill, including its language reference for the files changed.

3. Findings. Use the priority definitions of the project's review guidelines, or `review-oracle`'s where the project defines none. For each finding give:
   - [P0 | P1 | P2 | P3] <title>
   - Location: <file:line>
   - Issue: what is wrong.
   - Why: the evidence, which depends on the kind of finding:
     - Wrong behaviour (crash, wrong result, wrong diagnostic, invalid input accepted): an end-to-end test you wrote and ran through the project's user-facing entry point (for Nutcracker, a `.bal` program run with `go run ./cli/cmd run`, or a corpus test). Give its source, the command, the actual output, and the expected output with the specification clause it follows from.
     - Guideline or convention violation: quote the rule and name its source (AGENTS.md, a `*-review-guidelines` skill, `review-oracle`, a spec guardrail), and quote the violating code.
     - Departure from `spec.md` or `rca.md`: quote the item.
   - Correction: the specific change.
   Report only findings with this evidence. No findings is a valid result.

Return exactly:

SCOPE
<the scope statement from step 1>

FINDINGS
<findings, P0 first> | none

No praise, summaries or optional suggestions.
```

### 2. Match the findings

Pair the two reviewers' findings. Two findings match when they report the same defect: the same root cause or the same rule broken at the same place, however they are worded or tagged. Do not re-verify a finding; your job here is only to decide whether both reviewers report it. When unsure whether two findings are the same defect, treat them as unmatched.

- Agreed: reported by both. Its priority is the higher of the two tags.
- Unmatched: reported by only one reviewer.

Accepted findings are every P0 finding, agreed or unmatched, and every agreed P1–P2 finding. A P0 needs no second reviewer: its end-to-end test already demonstrates the crash or internal error.

### 3. Act on them

1. Fix every accepted finding, without re-validating it. The exception is a correction that would change the agreed design (a goal, non-goal, guardrail, API or mechanism in `spec.md` or `rca.md`, or new behaviour or abstraction they do not contain): ask the user, or return it to the calling skill for a decision, and wait before applying it.
2. Run proportionate checks after the fixes.
3. Do not fix unmatched P1–P3 findings or agreed P3 findings. Record them for the report, with each reviewer's text.
4. If any accepted finding was fixed, start another round, subject to the round limit. Otherwise the review is complete.

A caller reviewing a change it does not own (for example a PR review session) runs one round and presents the result instead of fixing.

## Report

Return, per round: both reviewers' scope statements, the accepted findings and what was done for each, and the other unmatched findings with the reviewer that raised each. Then list, across all rounds, the unmatched P1–P3 findings and agreed P3 findings still open, for the user to decide on. If accepted findings remain open, do not describe the review as complete.
