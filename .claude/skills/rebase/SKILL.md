---
name: rebase
description: Use this when user asks you to rebase against some branch
---

1. First identify what are the test/line commands
2. Then do an interactive rebase with that command as `exec`. IMPORTANT: you should make sure to run the tests for each rebased commit don't ever skip this unless explicitly asked.
  - Determine what commits should be actually included in the change. It is possible some commits were cherry picked from a branch that since have been merged to base and was modified. If that is the case (Check if there are commits with the same commit message) offer user to drop them. Precisely state both commits and from which branch they got merged in.
3. During rebase if you run into any test/lint failures or merge conflicts, diagnose the issue and present user with a fix. Only apply that when user agrees to that.
  - For test failures you should ammend the last commit with your fix when that is passing the test.
  - After applying the fix you should run test/lint commands you should loop on 3 until they pass
4. For fixup placement, descendant propagation or squashing, follow `../fix-up-target/SKILL.md`, including its authorization and verification requirements, instead of this skill's general rebase procedure.
5. Before a rebase with `exec`, check that the command's tools resolve in a non-interactive shell (e.g. `bash -c 'command -v go'`). For a stack, run the full tests only at each branch tip (`exec` before each `update-ref` line and at the end) and only build and lint on the other commits, unless the user asks for full tests on every commit.
