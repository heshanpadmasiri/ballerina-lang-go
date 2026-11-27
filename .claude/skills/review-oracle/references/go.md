# Go review reference

- Tests call `t.Parallel()` unless they share mutable state (globals, files, env vars); a test that cannot run in parallel says why in a comment.
- Every lint exception or suppression carries a comment saying why it is necessary.
- A failure signal that can never fire (`ok bool` that is always true, `error` that is always nil) comes out of the signature, and its dead checks come out of every call site. For each `if !ok` / `if err != nil` on a diff line, read every return of the callee.
- Mutually exclusive booleans become one enum; independent flags become a compact bit set with explicit set/check methods.
- Prefer a field's zero value to a separate "is set" boolean when a safe non-zero representation exists (e.g. an index offset).
- Structs that share fields and the methods on them embed one private base struct that implements those methods.
- A group of functions or types sharing a name prefix is usually a package.
- New exported symbols used only inside their package become unexported. Symbols with no non-test caller, and helpers reachable only from them, are removed.
- Keep superseded methods or interfaces only as a documented deprecation with a removal plan, never as silent forwarding proxies.
- An invariant a spec's guardrails state becomes a lint when it can be checked mechanically: a `golangci-lint` rule where one exists, otherwise a custom `go/analysis` analyzer kept with the project's tools, documented with the guardrail it enforces and run in CI. Review then checks that the lint exists and runs, not the invariant by hand.
