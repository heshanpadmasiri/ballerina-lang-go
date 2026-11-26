---
name: show-me
description: Help the user understand the current topic visually with concise call trees, stage paths, shape diffs, checked Mermaid diagrams and focused HTML pages. Other skills read this file for their visual conventions.
disable-model-invocation: true
---

# Show me

Help the user understand the current topic visually. Skip the preamble and keep prose brief. Pick the smallest view that makes the key point clear; one view is often enough, several is fine, all of them is never right.

Scripts are relative to this skill's directory, not the repository; run them by their full path.

## Views

- **Stage path:** how one input moves through the pipeline, with the first wrong stage marked. For Nutcracker the stages are tokens → ST → AST → desugar → semantics → CFG → BIR → runtime; dump them with `go run ./cli/cmd run --dump-tokens|--dump-st|--dump-ast|--dump-cfg|--dump-bir <file>`.

  ```text
  parse      ok
  desugar    ok
  semantics  ✗ `semantics/narrow.go:142` `if t.IsNever() {`  returns before nil is removed (P2)
  BIR        not reached
  ```

- **Call tree:** runtime control flow inside a stage.

  ```text
  analyzeFunction
    checkStatements
      narrowCondition      semantics/narrow.go:120
        diffType           semtypes/diff.go:48
  ```

- **File tree:** what each file is responsible for, or how a refactor moves responsibilities. Keep it shallow.

  ```text
  semantics/
  ├── narrow.go     # narrows types on conditions
  └── typecheck.go  # assigns types to expressions
  ```

- **Pseudocode:** an algorithm or decision, without the language's noise.

- **Shape diff:** use `diff` when the point is what changes in a shape that already exists: a call tree, a file tree, an API signature, a type, a CFG or BIR fragment. ` ` lines are today's shape, `-` lines leave it, `+` lines are proposed.

  ```diff
   narrowCondition
  -  diffType
  +  if t.IsNever()
  +    return never
  +  diffType
  ```

- **Full block:** when most of the shape is new, when omitted context would hide ownership or order, or when the user needs a copyable target.

- **Mermaid:** for a graph a tree cannot show: joins and loops (a CFG), messages between components (`sequenceDiagram`), or states (`stateDiagram-v2`). Choose Mermaid or a text view by what the content needs, never because a diagram failed to render.

- **HTML page:** for a comparison or concept too dense for the above, such as a type lattice or a large CFG side by side. Write one focused file in a scratch directory, use real labels and data, and open it (`open <file>`).

## Evidence

A view of today's code is a set of claims, and the same evidence rules apply as in prose. Give each node that states what the code does a citation (`` `path:line` `excerpt` ``), a bare `path:line` for location only, or a probe reference `(P<n>)`. Mark proposed shapes as proposals: `+` lines in a diff, or a heading such as "Proposed". `design/scripts/lint_spec.py` checks citations, locations and probe references inside `text`, `diff`, `mermaid` and untagged blocks of `spec.md`, `research.md` and `rca.md`.

## Mermaid must render

Never show the user a Mermaid diagram you have not rendered. Write it to a `.mmd` or Markdown file and run:

```sh
<this skill>/scripts/check_mermaid.py <file>
```

It renders every diagram with mermaid-cli and prints `file:line: mermaid: <error>` for each failure. Fix every error and rerun until it passes. `lint_spec.py` runs the same check on its documents. Exit status 2 means the checker could not run (mmdc or a browser is missing): tell the user the diagrams are unverified and why; do not present them as checked.

Frequent causes of failure:

- Unquoted labels containing `( ) [ ] { } < > " : ; |` or `#`: quote them, `A["narrow(t): Type"]`. Use `#quot;` for a quote inside a label.
- `end`, `graph` or `style` as a node ID: rename the node, for example `endNode["end"]`.
- Edges: flowcharts use `-->`, `-.->`, `==>` and `-->|label|`; sequence diagrams use `->>` and `-->>`. `->` is invalid in a flowchart.
- `subgraph id ["title"]` and every `subgraph` closed by `end`.

When the same diagram fails three times, simplify it: fewer nodes, plain IDs, all labels quoted. Do not switch to ASCII art to escape the checker.

In a terminal conversation Mermaid shows as source. Use it in documents the user reads in a renderer (a PR body, `spec.md` in an editor preview), or render it for them: `mmdc -i <file>.mmd -o <file>.svg && open <file>.svg`.

## Guidance

Place each view next to the short text it supports. Keep only the calls, files, types, stages and boundaries needed for the user's current question, or for the options of the current decision.
