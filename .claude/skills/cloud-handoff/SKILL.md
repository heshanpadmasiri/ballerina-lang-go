---
name: cloud-handoff
description: Implement the next slice of an approved spec.md in an unattended Claude Code cloud session, wait for it, land it locally, open its diff in hunk beside this Herdr pane and show the slice report. Use when the user asks to implement a spec in the cloud.
argument-hint: <feature slug>
disable-model-invocation: true
allowed-tools: Bash(python3:*), Bash(git:*), Bash(cat:*)
---

# Cloud handoff

Design stays local. Each slice of an approved `spec.md` is implemented by a fresh cloud session that runs unattended and pushes its result. This session hands the slice off, waits, lands the result in this worktree, opens the diff in hunk to the left of this pane and shows the slice report. The user reviews it here; touch-ups are made locally. Pull requests are opened locally; no personal credentials leave this machine.

Run it in Claude Code inside a Herdr GUI terminal, in the worktree that holds the approved design.

The only channel to the cloud is a series of branches on the remote, `cloud/<feature>/<seq>`, each one handoff commit on top of the real work. It carries the uncommitted design state, the skills, rules and settings the cloud lacks, and `.claude/handoff.json` with the status: `handoff` (from here), `slice-ready` or `blocked` (from the cloud). The cloud session follows [the cloud rules](references/cloud-rules.md).

`H` below means `python3 <this skill's directory>/scripts/handoff.py`, run from the worktree root. Every command prints JSON; read values from it, never guess them. If a command fails, stop and show the user its error.

## 1. Hand off

Preconditions, checked before anything else:

- `spec.md` is approved and stamped: `python3 <skills>/design/scripts/lint_spec.py spec.md --guard` passes (`<skills>` is the directory holding this skill).
- HEAD is on the feature's own branch, not a branch that exists on the remote: `land` fast-forwards it. `push` refuses otherwise; then ask the user for a branch name and `git switch -c` it.
- The base (HEAD on the first push) is on the remote.

Then:

1. `H push --feature <slug> --dry-run`. Check the file list holds the uncommitted design state and the tooling and nothing unexpected (secrets, build output). If anything is unexpected, stop and ask.
2. `H push --feature <slug>`. Note `seq` from the branch name.
3. `H launch --feature <slug>`. Tell the user the `session_url` in one line.

## 2. Wait

Run `H watch --feature <slug> --after <seq>` with Bash `run_in_background`, and end the turn with one line saying you are waiting for the cloud session. You are woken when it exits; read its output.

- `{"event": "deadline"}`: three hours passed without a push; the user was notified. Start the watch again.
- `{"event": "push"}`: continue.

## 3. Land and show

1. `H land --feature <slug>`. It fast-forwards this branch to the cloud's commits with their attribution trailers removed, restores the design state with `slice-report.md`, and prints `range` and `commits`.
2. `H show --range <range>`. It opens hunk on the slice to the left of this pane.
3. Your message to the user is, in this order:
   - the contents of `slice-report.md`, unedited (`cat` it and reproduce it exactly);
   - one line: the range shown in hunk on the left, and the cloud session URL;
   - one line: comment in hunk or here; touch-ups are made locally; say "approve" when the slice is right.

If the status is `blocked`, the report's `### Blocked` section says why; ask the user whether to fix it here and relaunch, or finish the slice locally.

## 4. Review and touch up

Handle the user's hunk comments with the `hunk-review` skill, and chat comments directly, under `implement`'s Findings and Deviations rules. Make agreed changes as fixup commits through `fix-up-target`, rerunning the affected checks, then run `H show --range <range of the slice>` again.

When the user approves the slice:

1. Squash the fixups with `git rebase --autostash` (the uncommitted design state otherwise blocks the rebase). Every slice commit must pass the project's checks.
2. Append the report to `progress.md`, with the user's decision on each deviation and the final base and tip, and delete `slice-report.md`. Record approved spec changes in `spec.md`; restamp only after the user approves changed tests.
3. `lint_spec.py spec.md --guard --slice-done <n>`.
4. If slices remain, go back to step 1. Otherwise run `implement`'s "After the last slice" steps here, show the whole change (`H show --range <base>..HEAD`), and ask before deleting the remote branches with `H clean --feature <slug>`. Open the pull request when the user asks.

## Cloud environment

Configure once at claude.ai/code in the environment's settings. Network: Limited, with package managers allowed. Setup script:

```bash
#!/bin/bash
# Build tools with the repository's toolchain (go.mod `toolchain` line), fetched from proxy.golang.org.
export GOTOOLCHAIN=go1.27.1 GOBIN=/usr/local/bin
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 || true
```

## Limits

- This session cannot read the cloud session's messages; its push is the only signal. The cloud never asks questions; it records its decisions in the report.
- The skills and rules are a snapshot taken by each push from here.
- The handoff commit is built with `git commit-tree`, so commit hooks do not run for it. It never reaches a pull request.
