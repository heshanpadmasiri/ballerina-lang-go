---
name: nutcracker-behavior-policy
description: Determine expected behaviour for Nutcracker, the Ballerina Go compiler/interpreter (ballerina-lang-go repository or Go module github.com/ballerina-nutcracker/ballerina). Use when investigating bugs, designing features, writing specs or reviewing expected results in this project.
---

# Nutcracker expected behaviour

Inputs are Ballerina programs. Read the relevant [Ballerina language specification](https://ballerina.io/spec/lang/master/) sections in full and derive expected results from them. Cite the clauses and explain how they apply to each case. Follow repository instructions about the specification version.

The latest `doc/lang/subset<N>.md` and its restrictions describe what Nutcracker supports today; they do not redefine language semantics. Use them to:

- decide whether reported behaviour is a bug or a request for unsupported behaviour. Ask the user before treating it as a feature;
- keep test programs to supported constructs apart from the behaviour under test, so a test fails only because of that behaviour;
- update the subset document as part of a change, as the repository's AGENTS.md requires.

A feature request expands the supported scope by definition; its expected results come from the specification, not from the subset document.

Issue comments, implementation output and test goldens do not determine correctness. jBallerina is a cross-check only; resolve disagreements against the specification rather than copying its output.

If the specification is ambiguous, show the relevant clauses and competing interpretations to the user before designing around one. Do not switch to the generic policy to settle language semantics.

Include the relevant rules and expected results in the spec. A fresh subagent that sees the test programs without their markers derives their results from the clauses independently; show any disagreement to the user. Follow `nutcracker-review-guidelines` during review and `nutcracker-test-oracle` during testing.
