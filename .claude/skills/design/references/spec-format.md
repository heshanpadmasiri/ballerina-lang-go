# Document format

`scripts/lint_spec.py` checks this format. The `fix` skill's `rca.md` format is in `../fix/references/rca-format.md`; it uses the evidence rules below. Run it on every document you hand to the user and fix every line it prints.

## Evidence

Every claim about current behaviour carries its evidence in the same sentence:

- **Citation:** a backticked `path:line` immediately followed by a backticked excerpt of that line or one of its neighbours, e.g. `` `semantics/narrow.go:142` `if t.IsNever() {` ``. Use double backticks for an excerpt containing a backtick. Paths are relative to the repository root. Quote what you read; the linter checks that the excerpt appears within two lines of the cited line.
- **Probe:** `(P<n>)`, referring to a probe defined under `## Probes`. A probe is something you ran. Use one for any claim about what the code does at runtime, and to settle any question a reading leaves open.
- **Assumption:** `(assumption)`. Only for a claim you could neither cite nor probe; say why in the same sentence. The user decides whether to accept it.

Citations show where; probes show what happens. A behavioural claim needs a probe unless the cited code makes it evident.

The same rules apply inside views (trees, diffs and diagrams in `text`, `diff`, `mermaid` or untagged blocks; see [show-me](../../show-me/SKILL.md)). The linter checks backticked citations there as in prose, checks that each bare `path:line` names an existing file and line, checks probe references, and renders Mermaid diagrams; it reports each diagram that fails at its line. In a `diff` block, `+` lines are the proposed shape and are not checked. Blocks under `## Probes` and `## Reproducer`, and blocks tagged with a programming language, are output or source, not views.

Define each probe once:

````markdown
### P1: <what it shows>
Command: `<command, run from the repository root>`
Output:
```text
<relevant output>
```
````

## research.md

```markdown
# Research: <topic>
## How it works today
## Specification            (when a behaviour policy applies: the clauses and what they require)
## Files read
## Gaps
## Questions asked
## Probes
```

## spec.md

```markdown
# <title>
## Goal
## Non-goals
## Repository context
## Guardrails
## API
## Design
## Test scenarios
## Slices
## Success criteria
## Probes
## Approved tests            (written by `lint_spec.py --stamp`; never edit by hand)
```

`## Probes` may be omitted when nothing refers to a probe. `## Approved tests` appears only after the user approves the spec.

### API

The functions and types the change adds, changes or removes, public or private, with signatures and their contract, grouped by file. Show a changed definition as a `diff` block of its signature, before and after.

### Design

Lead with the structure: a shape diff of the call tree or file tree the change touches, then prose for the control flow, data flow and failure handling the view cannot show. Write this before the design: the design explains how these definitions are implemented.

### Test scenarios

One block per scenario:

```markdown
### S1: <title>
- Rule: <the requirement, or the specification clause and how it applies>
- Test: `corpus/bal/subset10/10-foo/bar-v.bal` (new)
- Expected: <observable outcome and the source establishing it>
- Fails today: wrong-output
- Slice: 1
```

- `Test` names a test file, project directory or Go test, then `(new)` or `(existing)`. A new test is written during design, before approval. An existing one must already assert the expected outcome.
- `Fails today` and `Slice` are required for new tests and forbidden for existing ones. `Fails today` is one of `unimplemented`, `wrong-output`, `diagnostic-mismatch`, `panic-mismatch`, `crash`: how the test fails before the change, as established by running it.
- For Nutcracker recovery fixtures (`corpus/recovery/`), which cannot be skipped yet (ballerina-nutcracker/ballerina#1578), write `(new, slice N)` instead of `(new)`; the slice writes the test. Omit `Fails today`.

### Slices

```markdown
### Slice 1: <behaviour that works when it is done>
- Scenarios: S1, S3
- Depends on: none
```

Each slice is a vertical, end-to-end increment: it makes at least one new scenario pass and depends only on earlier slices.

### Success criteria

A list. Each item names the scenarios that demonstrate it, e.g. `- Constant folding of bitwise complement (S1, S2)`. Every scenario supports at least one criterion.

## Nutcracker pending tests

New corpus tests are listed in `test_util.UnsupportedTests` (`test_util/skip.go`) between these two lines, one entry per test, so every commit stays green until the slice that implements them:

```go
	// BEGIN pending: <spec title>
	"subset10/10-foo/bar-v.bal",
	// END pending: <spec title>
```

Entries are the keys the corpus drivers use: the path relative to `corpus/bal/` for language fixtures, `lib/...` for library fixtures and `project/<dir>` for project fixtures. The block may only shrink after approval, and is removed by the last slice.
