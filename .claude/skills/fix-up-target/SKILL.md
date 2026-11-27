---
name: fix-up-target
description: Own fixup target selection, commit placement, descendant rebasing, verification and authorized squashing. Use whenever creating or squashing fixup commits, on a single branch or a PR stack.
---

# Fixups

This skill is the authority for fixup handling. Calling skills specify which changes need fixes and when squashing or pushing is authorized; they do not redefine targeting, placement or descendant propagation.

## Target and ownership

- Identify the original commit each fix belongs to. In a stack, a branch owns only the commits it adds beyond its parent, not commits inherited from lower branches.
- When a fix connects two commits, identify the source that introduced the cause and the target that introduced the affected usage or test. For API changes these are the API change and its usage; for golden changes, the change that affected output and the test that owns the golden. Target the later commit in their dependency history so the result remains buildable after squashing. If there is no unambiguous target or owning branch, ask rather than guess.
- Split unrelated fixes by target and reviewable concern. Keep each fix and its regression tests together.

## Placement and propagation

- Each fixup must be a separate commit on the branch that owns its original target commit, not on a descendant branch merely because it is checked out. Preserve the original commit until squashing is authorized.
- Rebase all descendant branches on top of the updated owning branch, preserving their own changes and stack relationships. When several branches receive fixups, the final stack must contain each fix on its owning branch and inherit it in every descendant, without duplication.
- A request to create fixups includes this descendant rebasing. It does not authorize unrelated history cleanup, squashing or pushing.
- Preserve unrelated staged and unstaged work, including its staging state. Do not discard work to navigate or update branches. If safe placement or propagation is blocked, report the blocker.
- Follow the `commit` skill's staging, testing and hook requirements, but use fixup subjects identifying the original target instead of its ordinary message format. Keep those subjects intact while the fixups await review.
- On a conflict, present the cause and proposed resolution and wait for approval before changing the resolution. Use `resolve-conflicts` for that approval and resolution workflow; its user-owned continuation rule still applies. Report any paused propagation as incomplete.

## Verification and reporting

- Run the fix's regression tests and affected checks. After propagation, run the project's full tests at every affected branch tip, including descendants, and build/lint where required by the project.
- Verify that each fixup belongs to the correct branch, every descendant inherits it, and no unrelated branch changes or commits were lost or duplicated.
- Report each fixup, its target commit, owning branch and PR, affected descendants, checks run and any incomplete propagation.

## Squashing and pushing

- Keep fixups separate until the user requests or approves squashing, or an explicitly invoked calling workflow authorizes it at a defined checkpoint. Do not infer permission from a request to create fixups.
- Squash each fixup into its target and keep descendant branches based on the rewritten owning branch. Preserve every affected branch's final tree; if conflict resolution changes a tree, disclose the difference and rerun the relevant checks.
- Fixup-only placement and squashing use this skill's verification requirements rather than the general `rebase` skill's per-commit testing procedure. Pure squashing that preserves every affected branch's final tree needs no additional tests.
- Push only when the user or the calling workflow authorizes it. Report which branches were squashed or pushed and anything left pending.
