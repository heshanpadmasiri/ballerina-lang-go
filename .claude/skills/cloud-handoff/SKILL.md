---
name: cloud-handoff
description: Implement an approved spec.md in Claude Code cloud sessions, one unattended slice per session, while this local session watches them, lands each slice, shows its diff for review and applies the user's touch-ups locally. Use when the user asks to implement a spec in the cloud.
argument-hint: <feature slug>
disable-model-invocation: true
---

# Cloud handoff

Design stays local. Each slice of an approved `spec.md` is implemented in a fresh cloud session that never asks questions. This session watches for the result, lands it, shows the user the diff, makes the user's touch-ups locally, and hands the next slice to a new cloud session. Pull requests are opened locally; no personal credentials leave this machine.

Run it in the Claude desktop app, in the worktree that holds the approved design (normally the design session itself): it needs the Terminal panel, notifications and the diff view.

The only channel is a series of branches on the remote, `cloud/<feature>/<seq>`. Each is one handoff commit on top of the real work. It carries the uncommitted design state, the skills, rules and settings the cloud lacks, and `.claude/handoff.json` with the status: `handoff` (from here), `slice-ready` or `blocked` (from the cloud). Branches are never rewritten; every push creates the next seq. The cloud session follows [the cloud rules](references/cloud-rules.md).

`scripts/handoff.py` is relative to this skill's directory; run it from the worktree. `H` below means `python3 <skill dir>/scripts/handoff.py`.

## 1. Hand off a slice

Preconditions: the user approved `spec.md` and it is stamped (`lint_spec.py spec.md --guard` passes); the base, and every branch below it in the stack, is on the remote. Commits above the base travel with the handoff commit.

1. `H push --feature <slug> --dry-run`. Check the list: the uncommitted design state and the tooling, and nothing unexpected such as secrets or build output. On the first push, tell the user which commits above the base will be pushed.
2. `H push --feature <slug>`.
3. Run `H launch --feature <slug>` in the Terminal panel (`run_in_terminal`); `claude --cloud` refuses to run without a terminal. Read the tab for the session's URL and tell the user.

## 2. Watch

Watch with Monitor: `H watch --feature <slug> --after <seq just pushed>`. It prints one line and exits when a later seq appears, or after 25 minutes.

- On `{"event": "deadline"}`: check the cloud session's state in the session list (`ListAgents`). If it is still working, start the watch again. If it is idle and nothing was pushed, it stopped without handing back: send a push notification with its URL and tell the user, then keep watching.
- On `{"event": "push"}`: continue below.

## 3. Land and show

1. Send a push notification: `slice <n> ready` or `slice <n> blocked`.
2. `H land --feature <slug>`. It fast-forwards this branch to the cloud work, restores the design state with `slice-report.md`, and prints the range and commits. It discards the uncommitted state here only when it is exactly what an earlier handoff carried.
3. Open the diff view on the slice: `show_pane` with `pane: "diff"` and `diff_scope` set to the commit SHA. With several commits, open the first and name the others so the user can switch.
4. Then report from `slice-report.md`, in this order:
   - what the slice implements and its commits;
   - every deviation: spec item, blocker, options, the choice made. Ask the user to accept or reject each;
   - open review findings, checks run, pending tests that already pass;
   - for `blocked`, what blocked it and what was tried.

## 4. Review and touch up locally

The user comments on lines in the diff view and in chat. Handle each comment and each rejected deviation here, under `implement`'s Findings and Deviations rules: as fixup commits through `fix-up-target`, rerunning the affected checks. Show the diff again after each round. Do not send feedback to the cloud session.

When the user approves the slice:

1. Squash the fixups with `git rebase --autostash`: the uncommitted design state otherwise blocks the rebase. Every slice commit must still pass the project's checks.
2. Append the slice's report to `progress.md`, updated with the user's decisions on each deviation and the final base and tip. Delete `slice-report.md`. Record approved spec changes in `spec.md`; restamp only after the user approves changed tests.
3. Run `lint_spec.py spec.md --guard --slice-done <n>`.
4. If slices remain, go back to step 1 (push and launch). A `blocked` slice can be relaunched after fixing the cause, or finished here.

## 5. Finish

After the last slice, run `implement`'s "After the last slice" steps here. Then:

1. Open the diff view on the whole change (`diff_scope: "all"`) and summarise it: the commits, deviations and their decisions, and the open findings.
2. Ask before deleting the remote branches: `H clean --feature <slug>`.
3. Open the pull request when the user asks (`create-pr`).

## Cloud environment

Configure once at claude.ai/code in the environment's settings. Network: Limited, with package managers allowed. Setup script:

```bash
#!/bin/bash
set -e
# Go fetches the go.mod toolchain from proxy.golang.org on first use.
go version || true
command -v golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin || true
```

## Limits

- The skills and rules are a snapshot taken by each push from here; a cloud push carries them forward unchanged.
- The handoff commit is built with `git commit-tree`, so commit hooks do not run for it. It never reaches a pull request.
- There is no way to read a cloud session's messages from here: the branch and the session's idle state are the only signals.
