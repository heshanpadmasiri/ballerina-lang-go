---
name: test-oracle
description: "Test an implementation against its spec: audit required scenarios and coverage, add the missing tests, report failures. Use when asked to consult the test oracle, or when another skill tells a tester to use `test-oracle`."
---

# Test Oracle

Read `spec.md` first: test the goals only, plus any non-goal with defined behaviour. Read the tests the change already adds so you do not duplicate them. Follow the project's documented testing preferences; otherwise prefer end-to-end tests for system behaviour and unit tests only for contracts that cannot be reached end to end.

1. If a project test oracle exists (a `<project>-test-oracle` skill), run it in a fresh subagent and merge its findings into the report below.
2. Scenario audit: for each scenario in the spec, find the test that exercises its input or preconditions and asserts its expected outcome. Write missing tests, respecting any justified alternative verification in the spec. Independently check the requirements and relevant code paths for missing cases within the approved behaviour. If the spec supplies input partitions as supporting analysis, audit those too; no input-space table is required. An explicitly unsupported scenario gets a test showing the specified rejection. A missing behaviour decision goes in the user-decision section below; do not invent its expected result.
3. Coverage: account for all relevant feature cases, including boundaries and interactions, by mapping them to meaningful assertions in existing or new tests. Run coverage to find additional gaps; executing a code path alone does not prove its behaviour is asserted. Panics, asserts and internal errors may stay uncovered. Add tests for uncovered changed behaviour, subject to the requirement below. If an untested case already works without the change, report it as an existing-behaviour coverage gap rather than adding a test merely to increase coverage; do not count it as covered.
4. Run everything. A new test must fail with the change reverted; say how you checked. For a test that passes with the change, verify that reverting the change makes its behaviour assertion fail, not merely its setup or an unrelated dependency. Existing tests count toward feature coverage when they exercise and assert the required behaviour; they need not fail without this change.

Do not fix the implementation. Report with this template:

```markdown
# Tests added
- <test>: <spec item or scenario it covers>

# Coverage audit
- <feature case>: <existing or new assertion, or uncovered gap and reason>

# Failing tests within the spec
Independently validate before fixing.
## <test>
- spec item
- expected behaviour
- actual behaviour

# Out of scope, needs a user decision
Not covered by the spec. Do not fix without approval.
## <test>
- expected behaviour
- actual behaviour
```
