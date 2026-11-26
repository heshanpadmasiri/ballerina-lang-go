---
name: implement
description: Implement an approved spec.md one vertical slice per session, with fresh-context reviews, and report each slice to the user. Use when the user asks to implement a spec.
disable-model-invocation: true
---

# Implement

Precondition: `spec.md` exists, the user approved it, and it has an `## Approved tests` section (written by `design`). The spec is authoritative.

Work one slice per session. The session implements its slice, has it reviewed, reports it to the user and handles the user's comments. The next slice starts in a new session, from files only: `spec.md`, `progress.md` (scratch, never committed: one report per finished slice), the pending tests and the git log.

`lint_spec.py` and `nutcracker_check_failing.py` are in `../design/scripts/`, relative to this skill's directory, not the repository; run them by their full path from the repository root.

## Start of a session

1. Read `spec.md` and `progress.md`. The slice to implement is the first one without a report in `progress.md`. Its base is the previous slice's tip recorded in `progress.md`, or HEAD for the first slice.
2. Run `lint_spec.py spec.md --guard --slice-done <last finished slice>` (omit `--slice-done` before the first slice). It checks that the approved tests are unchanged and that no test has been added to the pending block. Stop and report to the user if it fails.

## Commits

A commit is one coherent, independently reviewable change with its tests and documentation, and it builds and passes on its own. Commits are not checkpoints: do not commit to save progress, and do not commit a change that is not yet passing.

- Commit only through the `commit` skill, which runs the project's checks first. Never bypass a hook (`--no-verify` or any equivalent); a failing hook is a defect to fix.
- Before the slice's review, run the project's build, lint and tests on every commit of the slice (`git rebase --exec '<checks>' <slice base>`, or the jj equivalent). A failing commit is fixed up with `fix-up-target` before anything else. A commit that mixes concerns is split.
- A fresh subagent then reads the slice's commits and reports any commit that is not one reviewable change or whose message does not describe its diff. Fix before review.
- Handle every fix after a slice's commits exist through `../fix-up-target/SKILL.md`, including targeting, placement, descendant rebasing, verification and squashing.
- This workflow authorizes squashing agent-review fixups before the user sees the slice, so the user reviews atomic commits. Fixups for the user's comments stay separate until the user approves the slice, so the user sees only those changes.
- After squashing, each slice commit must still pass the project's checks, and `lint_spec.py spec.md --guard --slice-done <this slice>` must pass.

## Per slice

1. Implement the slice. Make its scenarios pass. Never edit an approved test; a test that seems wrong is a deviation.
2. Remove the slice's tests from the pending block. For Nutcracker, generate their goldens following the repository's `manage-corpus-tests` skill and add any recovery fixtures the spec assigns to this slice.
3. Run `lint_spec.py spec.md --guard --slice-done <this slice>`.
4. Check the commits as described under Commits.
5. Run `lint_spec.py spec.md --list-new-tests` piped to `nutcracker_check_failing.py` (Nutcracker) or the project's equivalent, for the tests still pending. Report any that now pass; do not unskip them.
6. Run `review-loop` on the slice's commits, passing the slice. Follow its 3-round limit: pause for user approval before each additional batch of up to 3 rounds; do not proceed as though a paused review passed. Commit its fixes as fixups, then squash them.
7. Report the slice to the user (below) and wait.
8. Handle the user's comments under Findings as fixup commits, without squashing. Rerun the affected checks and `review-loop` on the slice, and report again with the fixup range. Repeat until the user approves.
9. Squash the remaining fixups. Append the slice's report, with its final base and tip, to `progress.md`. If slices remain, tell the user to start the next slice in a new session (`/implement`) and stop. After the last slice, continue below.

### Slice report

```markdown
## Slice <n>/<total>: <title>
Review: `<slice base>..<tip>`
- <short sha> <subject>, one line per commit
Fixups for your comments: `<tip you last reviewed>..<tip>` (only when reporting again)
Implemented: <the behaviour now supported, precisely, in the spec's terms>
Shape: <a shape diff of the call tree or file tree this slice changed (see ../show-me/SKILL.md)>
Tests now passing: <scenarios and tests>

Things to know
- Spec deviations: <each approved deviation, and its decision> | none
- Added beyond the spec: <what and why> | none
- Review findings left for you (unmatched between reviewers, or P3): <...> | none
- Pending tests that already pass: <...> | none
```

## After the last slice

1. Coverage: run the tests with coverage on the whole change. For each part not covered, except panics, asserts and internal errors, add a test within the approved behaviour. A new test must fail with the change reverted; say how you checked. Behaviour the spec does not define is a deviation, not a test to write.
2. Run `review-loop` on the whole change, with no slice, for issues that span slices, following the same 3-round approval gates. Commit its fixes as fixups and squash them.
3. If the spec says behaviour or performance is unchanged, run the project's benchmarks before and after and include the numbers.
4. Check that the pending block is gone and `lint_spec.py spec.md --guard --slice-done <last slice>` passes.

## Findings

Apply these rules to coverage and user findings. `review-loop` findings follow its own rule: fix the P0–P2 findings both reviewers report, subject to rule 2 below.

1. Validate each finding against the code, tests and spec before changing anything. Run a reproducer where possible. Reject a finding only with evidence that disproves it.
2. Before applying a correction, check it against the agreed design and the user's decisions. If it changes a goal, non-goal, guardrail, API, mechanism or other agreed decision, or adds behaviour, a mechanism or an abstraction the spec does not have, raise a deviation and wait for the user's decision. A reviewer does not authorize design changes.
3. Fix confirmed findings within the agreement and run the relevant checks. Investigate unresolved P0/P1 or explicitly blocking findings; do not treat them as rejected or complete the review while they remain unresolved.

Use the P0–P3 priority definitions in `review-oracle`. Guardrail violations, false claims and defects that would make the implementation incorrect or incomplete block completion regardless of priority; do not inflate their priority merely to express that gate.

## Deviations

Never change the design on your own. Ask when the spec cannot be followed, when a proposed correction would change it, or when following it forces the code to work around the design, even if the tests pass:

- a type needs an escape hatch the spec does not have: a cast, an untyped value, a nil-able field, or a parameter added only to pass state through;
- the same workaround appears in two or more places;
- a deviation has the same shape as an earlier one in this change;
- a function or type in the spec's API needs a different signature;
- an approved test seems wrong, or a slice cannot be done without changing a later slice's tests.

An edge case handled within the design is not a deviation.

```markdown
## Deviation request
- Spec item: <quote>
- Blocker: <what prevents it, or the workaround it forces: `file:line`, or the failing test>
- Options:
  - A (recommended): <change to the spec>; effort, risk
  - B: ...
```

Ask before applying the change, and pause work that depends on the answer. Record the user's decision in `spec.md`. If it changes the approved tests or slices, run `lint_spec.py spec.md --stamp` after the user approves the changed tests. Then rerun the relevant checks and reviews. Only the user's decision authorizes a spec change.

## Report

After the last slice, when no P0/P1 or explicitly blocking findings remain, report the commits, the coverage, test and benchmark results, the approved design changes, and any remaining P2/P3 findings.
