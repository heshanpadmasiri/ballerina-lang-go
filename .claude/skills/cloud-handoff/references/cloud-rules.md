# Cloud handoff session

This session runs in a Claude Code cloud environment, started from a `cloud/<feature>/<seq>` handoff branch. The user works locally and opens pull requests from their own machine.

- The first thing the session does is `python3 .claude/skills/cloud-handoff/scripts/handoff.py resume --expect <sha>` with the SHA from the first message. If it fails, report the error and stop: the session started from the wrong commit.
- After `resume`, `spec.md`, `progress.md`, `research.md` and the new tests are uncommitted, exactly as the design session left them. `.claude/` is gitignored and holds the skills and these rules; never commit it.
- Push only with `python3 .claude/skills/cloud-handoff/scripts/handoff.py push --feature <feature>`. It pushes the next `cloud/<feature>/<seq>` branch and leaves this worktree unchanged. Never run `git push` yourself, never push to any other branch, and never open, edit or comment on pull requests or issues.
- Push whenever the user needs to see your work: when you report a slice, after each round of fixups for the user's comments, and before you stop. Include the printed branch and commit in your report.
- Where a skill says to tell the user to start the next slice in a new session, push first and tell them the branch; the user starts the next session from it.
- Commands time out after 2 minutes by default. Run the long test suites with a timeout of up to 10 minutes, or in the background.
