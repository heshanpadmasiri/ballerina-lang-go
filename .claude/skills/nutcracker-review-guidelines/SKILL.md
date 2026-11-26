---
name: nutcracker-review-guidelines
description: Project-specific review guidelines for the Ballerina Go compiler/interpreter (ballerina-lang-go repo, Go module github.com/ballerina-nutcracker/ballerina, a.k.a. nutcracker). Use whenever you review, audit or validate any code, commit, PR, spec or corpus test in this repository.
---

Use together with the repository's `AGENTS.md`, the `review-oracle` skill and its Go reference.

## Finding priorities

Use a single P0–P3 tag in each finding's title, with no separate severity field. These project-specific definitions override the generic impact examples and urgency definitions in `review-oracle`, including its restriction of P0 to input-independent failures:

- **P0:** Any input that causes the front end to crash or report an internal error. Invalid input must produce an ordinary diagnostic, not a crash or internal error; valid input must not cause either. An input-specific failure still qualifies.
- **P1:** Valid input is handled incorrectly even though it falls within the scope of the current implementation. Examples include rejecting a supported valid program, computing the wrong result, or accepting it with incorrect type analysis. Do not count intentionally unsupported features as P1.
- **P2:** Any performance concern, or violation of stated programming patterns in project guidelines, the agreed design or repository instructions. Cite the pattern and the violating code; performance claims need concrete evidence rather than speculation.
- **P2 (outside scope):** A bug discovered during review that is outside the current feature's scope. Report it separately and suggest opening an issue rather than fixing it as part of the current change. This scope rule applies to unrelated, pre-existing bugs; it must not be used to downgrade a regression introduced by the change.
- **P3:** Nitpicks, such as minor wording or cosmetic consistency issues that do not violate a stated programming pattern.

For in-scope findings matching multiple categories, use the highest priority. Do not classify a stated-pattern violation as a P3 nitpick. Explicit guardrail and completion gates still apply regardless of priority. Priority does not authorize widening the agreed implementation scope or opening an issue without the user's approval.

## Compiler invariants

- Symbols and types are resolved after their resolvers, and malformed ASTs do not survive semantic analysis. A defensive nil/zero check in a post-resolution stage that hides one of these invariants is a defect. Establish module invariants up front instead of lazy initialisation and repeated checks.
- Never make a gate pass by hiding the failure: no internal error conditional on other diagnostics (`if !cx.HasErrors() { internalError(...) }`), no nil/zero check that hides a broken invariant, no guard in every node visit or declaration loop for one rare case. Fix the stage that produces the case.
- Callers check applicability before calling a helper on optional input. A helper does the requested work; it does not silently succeed when it does not apply, unless handling absence is explicitly its contract.
- Keep stage-specific temporal state in the stage context, not in shared AST or symbol structs. Make temporal dependencies explicit: ordered public stages return the values the next stage consumes instead of mutating hidden shared state.
- Handle a construct where the stage's dispatch puts it. A visitor or tree walk used to find a nested construct (e.g. walking every statement for a query action) runs on every node, analyses modelled subtrees twice and hides the supported positions. Ask for an explicit branch and a dedicated handler shaped like similar constructs (e.g. `foreach` in the CFG). A visitor is right only when the construct can appear at arbitrary depth.
- When a change normalises a name differently at some sites (e.g. keeps or strips a quote), list every lookup that matches names from two differently normalised sources (parameters vs included-record fields, named arguments, mapping keys, error details), run a reproducer for each and flag each lookup that fails.
- Analyses use the structure the design assigns them (the CFG, resolved symbols), not a proxy (source offsets, statement order, names). If the structure is missing for a case, ask for it to be built or ask the user.
- In a switch where only some cases return early, each case's handling lives inside that case or a helper called from it.
- A function whose sole behaviour is to raise an internal error or panic obscures the failing path: raise it at the call site, or remove it if unreachable.
- When a change routes some calls of a helper through an ordering gate or deferral queue (work that must wait until the declarations it depends on are resolved), list every other call of that helper the change adds or relies on, write a reproducer that reaches each one before the gate opens (e.g. inside an anonymous-function body in a module-level record field default or type definition), and flag each call that bypasses the gate.

## Code shape

- Flag using a single mutex to protect multiple unrelated structs or data structures; give each independent protected state its own mutex.
- Inline shallow private functions with one caller, except data preparation, interface accessors, parent proxies and public inner-package proxies.
- When similar functions differ only in data access, use an existing or minimal interface; when their preparation differs, separate type-specific preparation from shared computation.
- A stored derived field can drift from its source. Accept it only with a clear performance reason recorded in a comment; make the source fields private and update source and derived fields together.
- When a change adds or changes a parameter of a shared constructor or function, check every caller, including native Go code under `lib/` that builds Ballerina values, against the Ballerina declaration it implements, and flag each caller passing a placeholder nil/zero where that declaration needs a real value.
- Cache repeated stable method calls within a function; avoid repeated shared-map lookups in one computation by holding stable handles or indexes. Consider struct-of-arrays where a hot loop reads one field from an array of structs.

## Documentation

- A field comment adds what the name and type do not. If a better name would make the comment unnecessary, ask for the rename.
- A function comment gives the behaviour and the contract (inputs, results, preconditions), never the implementation, its cost, or why it is shaped for particular callers. Flag comments longer than the thing they document.
- Check every claim a changed comment makes about what the code does (e.g. which items fail, what is always or never true) against each test fixture the change adds, by tracing or instrumenting the code on that fixture, and flag any claim a fixture contradicts.

## Diagnostics and restrictions

- List every diagnostic the change adds, and every existing diagnostic that gains a new caller or new arguments. Trace each value interpolated into its message (names, types, positions) back through every caller: a compiler-generated symbol, desugared function or other synthetic name must never reach user-facing text. Read the `stderr` of each added or changed `-e` golden and flag any identifier its source does not contain.
- For each new way a program is rejected, decide from the spec whether a conforming implementation must reject it. For rejections that come from the compiler's own evaluation strategy (dependency or ordering checks, eager evaluation, folding), write a program where the rejected construct is never used, or every use avoids it, and check whether the spec accepts it (cross-check with jBallerina). A rejection the spec does not require is a restriction: it is stated under its construct in the latest `doc/lang/subset<N>.md` and covered by an `-e` test. Flag each undocumented restriction separately; documenting one does not cover another.
- A `doc/lang/subset<N>.md` sub-bullet states only a restriction or implementation-specific behaviour: flag any added sub-bullet that restates spec behaviour or enumerates supported cases (e.g. "Supports X in Y"), because a fix that brings a construct to full spec conformance satisfies the subset-doc requirement by removing the obsolete restriction or, if none was documented, by leaving the doc unchanged.
- For each restriction a change documents in `doc/lang/subset<N>.md`, check the code that decides it for every exception it already makes (special-cased functions, types or contexts) and flag a restriction text that omits one.

## Tests and corpus

- Prefer corpus tests. A Go unit test needs a comment saying why the scenario cannot be reached through a corpus test.
- Run coverage between `main` and the branch. Every error other than an unimplemented or internal error is covered by a `-e` (front end) or `-p` (back end) test; everything else by corpus tests written in user-reachable code.
- Internal errors and panics that are unreachable by construction need no test. Those that check invariants guaranteed by previous stages get corpus tests showing they cannot fire.
- New corpus tests live in the latest subset: features under `<latest-subset>/<id>-<feature>/`, bug regressions under `<latest-subset>/<id>-bug/`, even when the feature belongs to an older subset, with goldens and embedded source paths aligned. Flag additions to historical subsets.
- Each test fails if the behaviour it targets is broken. When the property is invisible in printed output (mutability, readonly-ness, a value's type, identity), the test checks it at runtime in a way the compiler cannot constant-fold (e.g. through an `any` parameter before the `is` check).
- Each new corpus test fails without the change. For a test passing with the change, verify that its behaviour assertion fails without it, not merely setup or an unrelated dependency. Existing tests count toward feature-case coverage when they exercise and assert the required behaviour; they need not fail without this change. An untested case already passing without the change is an existing-behaviour coverage gap, not a test to add merely to increase coverage.
- No two tests added by the change exercise the same scenario, and new tests do not duplicate existing coverage. A test differing from another new one only in comments, identifiers or literals is removed with its goldens or changed to a distinct case.
- Corpus tests reach behaviour only through code a user program can write. Flag imports or calls that exist for compiler-generated code (e.g. `lang.'__internal`). If no user program can reach a check, it is an internal invariant: remove the direct test.
- Derive expected behaviour from the Ballerina spec using a fresh subagent. jBallerina is a cross-check only; flag any disagreement instead of copying jBallerina.
- Run every whole `-v` corpus file the change adds or edits through jBallerina (`bal run`), and for each statement jBallerina rejects or whose output differs, decide from the spec which side is right and flag any golden that records output for a program the spec rejects.
- A front-end change that adds a language feature is recorded in the latest `doc/lang/subset<N>.md`.
- No test, CI check or script enforces code shape by reading the repository's own source (`go/parser`, grep over `.go` files, identifier allowlists or snapshots). State such properties as review criteria in specs and guardrails instead. Standard linters, and `go/analysis` analyzers that enforce a stated guardrail invariant (see the Go reference), are not covered by this rule.
- A test that drives an external tool (git, a shell, a compiler toolchain) passes under every supported configuration of that tool, not only the reviewer's defaults: for each fixture value the test hard-codes (IDs, hashes, paths, branch names, formats), name the tool default or environment variable its validity depends on, re-run the test with that default changed, and flag any failure.
- A script that classifies changes from Git output identifies each allowed case by its defining property, not by a Git heuristic (rename/copy similarity thresholds): build a scratch repo where the heuristic flips (e.g. a rename that also rewrites most of the file) and flag a misclassification.
