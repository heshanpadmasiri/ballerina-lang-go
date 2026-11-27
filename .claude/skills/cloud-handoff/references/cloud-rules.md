# Cloud handoff session

This session runs unattended in a Claude Code cloud environment. It implements exactly one slice of `spec.md` and hands it back. The user reviews it locally, makes any touch-ups there, and opens pull requests from their own machine. Nobody answers questions in this session, so never end a turn to ask one.

These rules override `implement` where they differ.

## Start

1. Run `python3 .claude/skills/cloud-handoff/scripts/handoff.py resume --expect <sha>` with the SHA from the first message. If it fails, push nothing, report the error and stop: the session started from the wrong commit.
2. After `resume`, the design state is uncommitted exactly as the user left it: `spec.md`, `progress.md`, the pending tests and the pending block. `.claude/` is gitignored and holds the skills, rules and settings; never commit it. Never commit the design documents.
3. The slice to implement is the first slice in `spec.md` without a report in `progress.md`. Its base is HEAD after `resume`.

## Implement the slice, unattended

Follow `implement`'s "Start of a session", "Commits" and "Per slice" steps 1–6, with these changes:

- **Deviations.** Where `implement` says to raise a deviation and wait, choose the option you would recommend, apply it, and record it under Deviations in `slice-report.md`: the spec item, the blocker with `file:line` or the failing test, every option, the choice and why. Do not edit the approved tests or the stamp. If the only workable option changes an approved test, record it, leave the test unchanged and finish the rest of the slice.
- **Review.** Run one batch of `review-loop` (up to 3 rounds) without pausing. Fix as `review-loop` says, squash the fixups, and list every finding still open in `slice-report.md`.
- **Stop after the slice.** Do not start the next slice, do not edit `progress.md`, and do not run the after-the-last-slice steps; the user does those locally.

## Hand back

Write `slice-report.md` in the repository root using `implement`'s slice report format, with these sections filled in:

- `Deviations`: every decision made under the rule above, or `none`.
- `Open review findings`: every finding not fixed, with its priority, or `none`.
- `Checks`: the commands you ran on each commit, and their results.

Then push once:

```sh
python3 .claude/skills/cloud-handoff/scripts/handoff.py push --feature <feature> --status slice-ready --slice <n>
```

If you cannot finish the slice (the environment is broken, or a check fails and you cannot fix it), write what you tried and what is blocking in `slice-report.md` and push with `--status blocked` instead.

Never run `git push` yourself, never push to any other branch, and never open, edit or comment on pull requests or issues. End with the pushed branch and commit.

## Environment

Commands time out after 2 minutes by default. Run long test suites with a timeout of up to 10 minutes, or in the background and wait for them.
