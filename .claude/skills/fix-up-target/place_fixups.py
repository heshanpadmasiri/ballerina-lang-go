#!/usr/bin/env python3
"""GIT_SEQUENCE_EDITOR for `git rebase -i --no-autosquash --update-refs <stack-base>`.

Moves every `fixup! <subject>` commit to the top of the stack branch that owns
its target: directly before the first `update-ref` that follows the target's
`pick` line. Fixups whose target is on the top branch stay at the end. Every
other line keeps its order. Exits non-zero, leaving the todo untouched, when a
fixup's target cannot be found, so git aborts the rebase before it starts.
"""

import re
import sys

PICK = re.compile(r"^(pick|p)\s+(\S+)\s+(?:# )?(.*)$")


def subject(line: str) -> str | None:
    match = PICK.match(line)
    return match.group(3).strip() if match else None


def main() -> int:
    path = sys.argv[1]
    lines = open(path, encoding="utf-8").read().splitlines()
    body = [line for line in lines if line and not line.startswith("#")]

    fixups = [line for line in body if (subject(line) or "").startswith("fixup! ")]
    rest = [line for line in body if line not in fixups]

    for fixup in fixups:
        target = subject(fixup).removeprefix("fixup! ").strip()
        while target.startswith("fixup! "):
            target = target.removeprefix("fixup! ").strip()
        target_index = next(
            (i for i, line in enumerate(rest) if subject(line) == target), None
        )
        if target_index is None:
            print(f"place_fixups: no commit with subject {target!r} for {fixup!r}", file=sys.stderr)
            return 1
        insert_at = next(
            (i for i in range(target_index + 1, len(rest)) if rest[i].startswith("update-ref ")),
            len(rest),
        )
        rest.insert(insert_at, fixup)

    with open(path, "w", encoding="utf-8") as out:
        out.write("\n".join(rest) + "\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
