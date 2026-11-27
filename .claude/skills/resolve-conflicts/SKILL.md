---
name: resolve-conflicts
description: Plan and resolve merge conflicts with user approval, then verify tests pass. Use when resolving merge or rebase conflicts in git or jj.
---
Resolve all merge conflicts in the current repository:

1. Identify every conflict and analyze the code on both sides (the incoming change and the current head) to understand the intent of each. Before planning, read in full the commits on each side that touch the conflicted files: their messages, diffs and tests (in git, `git log --merge -p -- <file>`). The markers show what changed; the commits show why.
2. For each conflict, summarize what causes it and how you plan to resolve it. Do not modify any code until the user approves the plan.
3. After the user approves, apply the resolutions. Then run the unit tests: all of them must pass, not just the build.

IMPORTANT: Never continue the merge or rebase after resolving the conflicts (no `git merge --continue`, `git rebase --continue`, or `git commit`). The user does that.
IMPORTANT: Never run `git add` on your changes.
IMPORTANT: In jj, never squash the current commit.
