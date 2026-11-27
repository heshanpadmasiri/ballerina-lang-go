# Research

Research establishes how the system works today, separately from what is to be built, so reading is not steered toward a design that is already forming.

## 1. Questions

Write neutral questions about current behaviour along the paths the request touches: how a construct flows through each stage, where data is stored and who reads it, which tests cover it, and what the system does today with representative inputs. Ask only "how does it work"; never "how would we", and never name the requested change, its motivation or a candidate design.

## 2. Explore

Give each fresh subagent its questions, the repository path and this prompt. Do not give it the request, the issue or any proposed cause.

```text
Document how the code in <repository> works today, to answer the questions below. Work read-only; run probes in a temporary workspace.

You are a documentarian, not a critic: describe what exists. Do not suggest improvements, designs or fixes.

Read the complete functions that determine the behaviour, their callers, data types and tests. Search hits and signatures are navigation, not evidence. Run probes to establish runtime behaviour: a program and its output, a temporary log or panic to show a path is reached, or a deliberate break to show which tests cover a behaviour. Choose inputs whose outcome distinguishes competing explanations.

Report, using the evidence format in <spec-format.md path>:
- Findings: each claim with a citation and excerpt, or a probe with its command and output.
- Files read.
- Boundaries: where the path you traced ends.
- Non-obvious things: surprises and invariants.
- Gaps: what you could not establish, and why. Be honest; a gap is better than a guess.

Questions:
<questions>
```

Split questions by area so each subagent's path fits its context. For Nutcracker, a program's path through the stages is easiest to trace with the CLI's stage dumps (`go run ./cli/cmd run --dump-tokens|--dump-st|--dump-ast|--dump-cfg|--dump-bir <file>`); desugared output has no flag, so use the desugar corpus driver in a temporary worktree.

## 3. Synthesise

Merge the reports into `research.md`. Resolve contradictions between reports by reading the code or running a probe yourself, not by choosing one. Verify any finding your design will depend on before relying on it.

Give the remaining gaps at most one more round of targeted questions. Label what is still unknown as an assumption.

When a behaviour policy applies, add the specification clauses the request touches and what they require under `## Specification`.

## 4. Present

Run `scripts/lint_spec.py research.md` and fix every problem. Present "how it works today" to the user: the current behaviour along the path, shown as a stage path or call tree with a citation or probe on each node (see [show-me](../../show-me/SKILL.md)), the specification clauses, the files read and the gaps, with the questions asked as a single footer line.

When the user corrects it or names a missing area, do not accept the correction on trust: research it the same way, update `research.md`, and present the change. Continue until the user accepts the picture.

Later factual questions, from the user's answers or from the design touching new code, go to targeted research subagents; add their findings to `research.md`.
