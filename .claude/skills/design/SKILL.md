---
name: design
description: Refine a feature idea, agree its API and design with the user and write spec.md with failing acceptance tests, ready for an implementation agent. Use when asked to refine an idea or plan a feature. Does not implement it. Bugs go through `fix`.
argument-hint: [issue number, or a description of the idea]
---

# Design

Produce `spec.md` for an implementation agent with no conversation history, together with the acceptance tests it must make pass. Do not implement or commit. Probes and spikes in a temporary workspace are allowed; a spike is throwaway code written only to answer one question, such as whether a design fits the existing types and callers, and is deleted afterwards. `research.md` and `spec.md` are scratch files and are never committed.

Paths to `references/` and `scripts/` are relative to this skill's directory, not the repository; run the scripts by their full path from the repository root.

Write every document in [the document format](references/spec-format.md) and run `scripts/lint_spec.py <document>` before showing it to the user; fix every problem it reports.

Whenever you show the user a structure, follow [show-me](../show-me/SKILL.md): the smallest view that makes the point, with the same evidence as prose, and only Mermaid diagrams that render.

## Evidence

Every claim about the current system is cited with an excerpt, backed by a probe you ran, or labelled an assumption. Citations show where; probes show what happens. When reading leaves competing explanations, run a probe whose outcome distinguishes them rather than reasoning about which is right.

Do not design around an unknown repository fact. Investigate it first. If you cannot validate it, tell the user the assumption, the evidence, why it remains unverified and the proposed validation, and wait for their decision. Never present an assumption as a fact.

## 1. Frame

Restate the goal, the problem it solves and the smallest change you expect to meet it. If the evidence already rules out the requested change, show the evidence and stop.

## 2. Expected behaviour

Read the project's instructions. Check the available skill descriptions for a matching behaviour policy and read it. Follow its rules for determining expected behaviour. If none applies, use [the generic fallback](references/expected-behavior.md).

## 3. Research

Follow [research](references/research.md). Find out whether the requested behaviour, or a capability to reuse, already exists before proposing new mechanisms. Present how the system works today and let the user correct it before resolving any decision.

## 4. Resolve decisions with the user

Facts are your job: find them with research, never ask the user for something you can look up or probe. Decisions are the user's.

Work through the decision tree in rounds. Each round, ask every decision whose prerequisites are settled, numbered, each with its options, tradeoffs and your recommended answer, worded so that "yes" accepts the recommendation. When options differ in structure, show each as a shape diff against today's code (call tree, API signature or file tree). After each round, research what the answers expose and ask the next round. Revisit dependent decisions if an earlier answer changes. Continue until no open decision remains; there is no limit on rounds. An unanswered question does not authorize your recommendation.

Resolve the API first: the functions and types the change adds, changes or removes, with signatures and contracts, written against the code the research documented. Then design the control flow, data flow and failure handling that implement it. When it is uncertain whether a design fits the existing types and callers, settle it with a spike that only needs to build, not with argument.

Never change an agreed goal, constraint, guardrail or API without approval. When one cannot be satisfied, show the conflict and evidence, recommend options, and wait. Return to this step whenever writing the tests or the spec exposes another decision.

## 5. Acceptance tests

Derive the test scenarios from the requirements, the affected code paths and the expected-behaviour sources, including success, failure, boundary and interaction cases. If a project test oracle exists (a `<project>-test-oracle` skill), use its rules for deriving cases. If a property cannot be tested, say why and specify another way to verify it.

Write each new test now, following the project's testing conventions. Each must fail today because the requested behaviour is missing, not because the test is wrong. An existing test may cover a scenario only if its assertions establish the expected outcome; read them first.

Have a fresh subagent derive the expected results independently: give it the new test programs with their expected-result assertions removed, the behaviour policy and its sources, and ask it to derive each program's observable result with its reasoning. Compare its results with your tests. Show every disagreement to the user with both derivations; do not settle it yourself.

Run each new test and record how it fails as the scenario's `Fails today`. A test that passes does not test the change. A test that fails differently from your prediction means your picture of the code is wrong: research why before continuing.

Group the scenarios into slices. A slice is a vertical, end-to-end increment that makes at least one new scenario pass and can be verified without any later slice. Prefer several thin slices. If the work cannot be sliced vertically, say why.

### Nutcracker

For the Nutcracker repository (Go module `github.com/ballerina-nutcracker/ballerina`):

- New tests are corpus tests under their final names (`-v`, `-e`, `-p`), listed in the spec's pending block in `test_util/skip.go` (see [the format](references/spec-format.md)) so every commit stays green. Scenarios the user decided to defer beyond this change keep or get `-fv`/`-fe`/`-fp` names instead.
- Check the new tests with `scripts/lint_spec.py spec.md --list-new-tests | scripts/nutcracker_check_failing.py`, run from the checkout. It runs them end to end with the pending entries removed and reports any that pass or fail differently from `Fails today`.
- Recovery fixtures cannot be skipped yet (ballerina-nutcracker/ballerina#1578): write them into the spec for their slice to add, as `(new, slice N)`.

## 6. Write the spec

Write `spec.md`. Resolve implementation decisions in it: do not leave missing mechanisms or ambiguous behaviour for the implementer to invent. Copy into Repository context every fact from `research.md` the implementer needs; the implementer does not read `research.md`. Exclude conversation summaries, previous iterations and progress commentary. Include rationale only when it explains a decision or constraint the implementer needs, such as why an obvious alternative was rejected.

The spec is not sent for review. Check it yourself against the user's requirements and decisions, then lint it.

## 7. Present for approval

Present the spec: link to it, summarise the API and design as shape diffs, list the scenarios with how each fails today, and list the slices. Report any assumptions still open.

When the user approves, run `scripts/lint_spec.py spec.md --stamp` to record the approved tests. A later change to the spec or its tests needs the user's approval and a new stamp. Do not start implementation.
