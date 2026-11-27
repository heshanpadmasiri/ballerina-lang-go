---
name: review-oracle
description: Generic code and design review guidelines. Use when asked to consult the review oracle, or when another skill tells a reviewer to use `review-oracle`.
---

# Review Oracle

Generic guidelines, used together with the project's own and the language reference for the files changed:

- Go: `references/go.md` in this skill directory.

## API and type definitions

- Illegal states should be unrepresentable. Several enum or flag fields that allow invalid combinations become one enum of the valid cases. A value with a fixed set of forms is a sum type (or the language's closest equivalent, e.g. a sealed interface with a private marker method plus a linter).
- A value known at construction is required at construction, not set later. Setters make state changes hard to reason about.
- An interface with one implementation and no real boundary is indirection, not abstraction.

## Patterns to flag

- Callbacks and hooks that decide behaviour mid-flow. If the decision can be made ahead of time, pass the decision, not a function.
- Shallow helpers: one caller or almost no body, hiding what happens behind an extra hop.
- Duplicated logic that an existing or parameterised function already covers. When a change adds a helper, every existing site that could use it must.
- Hidden failures: errors reported or suppressed depending on unrelated state; checks that hide a broken invariant instead of fixing what broke it.
- Symptom patches at the crash site when the root cause is known or findable.
- Edge-case guards on the common path: a check in every loop, visit or call for one rare case. Fix the source so the case cannot reach them.
- Speculative abstraction: wrappers or extension points for needs the change does not have.

## Tests

- A new test must fail without the change. Check by reverting the change or breaking the behaviour it targets.
- Use coverage on the change to find scenarios without tests.
- Internal illegal states (not inputs) should be unrepresentable first; where they remain representable, the system must detect them and fail meaningfully, and that is tested.
- Tests that read the project's own source to enforce code shape (parsing source, grepping identifiers, snapshots of names) are not tests of behaviour. Leave code-shape rules to review.

## Documentation

- Follow the language's and the codebase's comment conventions.
- A doc comment adds what the signature does not: caller restrictions (nil behaviour, ordering requirements, invariants to keep). Prefer enforcing a contract in code to documenting it.
- Implementation notes (performance, clever algorithms) go inside the body, briefly, not in the public comment.
- A field comment says what the field represents, and only when the name does not. A function comment says what it does and the contract, never how it is implemented or why it is shaped for a caller.

## Things to flag for a human reviewer

Only when the audience is human:

- Changes to sections unrelated to the spec (check `spec.md`), with why each was needed.
- Changes not traceable to a spec item, with the mechanism each adds.
- New dependencies, and API changes the user did not agree to.
- Disabled tests, and tests that pass without the change.
- Scenarios still missing tests.

## Presenting findings

Use one priority tag in each finding's title; do not add a separate severity or high/low field:

- **[P0]** — Drop everything to fix. Blocking release/operations. Only for universal issues that do not depend on assumptions about inputs.
- **[P1]** — Urgent. Should be addressed in the next cycle.
- **[P2]** — Normal. To be fixed eventually.
- **[P3]** — Low. Nice to have.

Order findings P0 through P3. Assign priority from impact and urgency, not confidence or the mere existence of proof. For a bug, a reproducing test is the best evidence. An unverified claim is unresolved, not a lower-priority confirmed defect. Project guidelines may supply focused examples using this same scale.
- To learn what was agreed, read `spec.md`. When reviewing one slice of it, judge only that slice's scenarios.
