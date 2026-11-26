---
name: cloud-handoff
description: Hand an approved spec to a Claude Code cloud session that implements it slice by slice, then bring the finished work back to the local checkout. Use when the user asks to implement in the cloud, start or continue a cloud slice, or land cloud work locally.
argument-hint: push|launch|land --feature <slug>
---

# Cloud handoff

Design stays local; `implement` runs in the cloud; pull requests are opened locally. The only channel between the two is a series of branches on the repository's remote, `cloud/<feature>/<seq>`. Each one is a handoff commit on top of the real work, carrying the uncommitted changes (spec, new tests, pending block, `progress.md`), the skills and personal rules the cloud lacks, and `.claude/handoff.json`. Branches are never rewritten: every push creates the next seq. Pushes from the cloud go through the GitHub proxy; no personal credentials leave this machine.

`scripts/handoff.py` is relative to this skill's directory. Run it from the repository checkout. Inside a cloud session, follow [the cloud rules](references/cloud-rules.md) instead of this file.

## Hand off (local)

Preconditions: the user approved `spec.md` and it is stamped; the base, every branch below it in the stack, and HEAD are pushed. `push` refuses otherwise.

1. `handoff.py push --feature <slug> --dry-run` and show the user the file list. Check that it contains the uncommitted design state and nothing else unexpected (secrets, build output).
2. `handoff.py push --feature <slug>`. It prints the branch and the handoff commit.
3. `handoff.py launch --feature <slug>` starts a cloud session with `/implement` on the latest branch, pinned to that commit. The session's first step, `resume`, refuses to continue from any other commit.

## Each slice

The session reports a slice and pushes `cloud/<feature>/<seq>`. The user reviews `git fetch origin cloud/<feature>/<seq>`, with the slice range from the report, locally or in the web diff, and replies in the session. Fixups are pushed as the next seq. Once the slice is approved, run `launch` again for the next slice: it starts a fresh session on the latest seq, as `implement` requires.

## Land (local)

1. From the local branch the work belongs on (its HEAD must be an ancestor of the cloud work), run `handoff.py land --feature <slug>`. It fast-forwards to the work under the latest handoff commit and restores the uncommitted state: `progress.md`, the spec and any still-pending tests. A local file that differs is kept as `<file>.orig`.
2. If the stack below moved since the handoff, land fails; rebase the cloud work with the `rebase` skill first.
3. Open the pull request locally (`create-pr`). Delete the `cloud/<feature>/*` branches once it is open, after asking the user.

## Cloud environment

Configure once at claude.ai/code, environment settings. Network: Limited, with package managers allowed. Setup script:

```bash
#!/bin/bash
set -e
# Go: go.mod's toolchain line is fetched from proxy.golang.org on first use when newer.
go version || true
command -v golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin || true
```

## Limits

- Skills and rules are a snapshot taken at each local push; the cloud carries them forward unchanged. Push again from local to refresh them.
- The handoff commit is built with `git commit-tree`, so commit hooks do not run for it; it never reaches a pull request.
- Desktop-only skills (herdr, hunk, session tools) are not carried.
