---
name: commit
description: "Commit message format and staging rules. Read before creating any git or jj commit."
---

First detect whether the project uses jj (a `.jj` directory exists at the repository root) or git. jj takes precedence over git.

For fixup requests, follow `../fix-up-target/SKILL.md` for targeting, subjects, placement, descendant rebasing and verification. The staging, testing and hook requirements here still apply; the ordinary message format below does not.

## Format

`<type>(<scope>): <summary>`

- `type` REQUIRED. Use `feat` for new features, `fix` for bug fixes. Other common types: `docs`, `refactor`, `chore`, `test`, `perf`.
- `scope` OPTIONAL. Short noun in parentheses for the affected area (e.g., `api`, `parser`, `ui`).
- `summary` REQUIRED. Short, imperative, <= 72 chars, no trailing period.

## Notes

- Body is OPTIONAL. If needed, add a blank line after the subject and write short paragraphs.
- Do NOT include breaking-change markers or footers.
- Do NOT add sign-offs (no `Signed-off-by`).
- Only commit; do NOT push.
- If it is unclear whether a file should be included, ask the user which files to commit.
- If any candidate file is a design document (`spec.md`, `research.md`, `rca.md`, `progress.md`, `plan.md`, `deferred.md`, `design-ledger.md`), explicitly ask the user whether to include it before staging or committing it. Do not infer consent from broad instructions such as "commit all" or from the file already being staged.
- Treat any caller-provided arguments as additional commit guidance. Common patterns:
  - Freeform instructions should influence scope, summary, and body.
  - File paths or globs should limit which files to commit. If files are specified, only stage/commit those unless the user explicitly asks otherwise.
  - If arguments combine files and instructions, honor both.
- If the user has already staged some files, consider only the staged files for committing (git only; jj has no staging area).

## Steps

1. Infer from the prompt if the user provided specific file paths/globs and/or additional instructions.
2. Resolve ambiguous files and design documents as described in Notes before committing.
3. Run the project's tests. If they fail, ask the user whether to commit anyway.
4. Stage only the intended files. If nothing was pre-staged and no files were specified, stage all changes.
5. Commit:
   - git: `git commit -m "<subject>"` (and `-m "<body>"` if needed).
   - jj: `jj commit -m "<message>" [paths...]`; pass paths to limit the commit to specific files.

IMPORTANT: never work around commit hooks (for example with `--no-verify`) unless the user specifically authorizes it.
IMPORTANT: never modify code as part of committing unless the user specifically authorizes it.
