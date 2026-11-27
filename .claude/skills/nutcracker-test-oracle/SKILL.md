---
name: nutcracker-test-oracle
description: Find missing corpus tests for a nutcracker change by deriving cases from the Ballerina language specification. Use when testing a change in the ballerina-nutcracker/ballerina repository, or when `test-oracle` runs there.
---

# Nutcracker Test Oracle

## During design

When `design` uses these rules there is no diff yet. Derive cases from the request, the research and the specification sections it touches, using steps 2–4 and the input-space partitioning below, with these differences:

- The supported scope is the current subset plus the behaviour being designed. Other unsupported features are still not missing tests.
- New tests are expected to fail today. Cases that hit an unimplemented error are the point, not omissions.
- Do not use the findings format below. Each case becomes a scenario in `spec.md`, with its test written to its final corpus path.

## After implementation

1. From `spec.md` and the diff, identify which parts of the Ballerina language specification (https://ballerina.io/spec/lang/master/) the change implements. Restrict yourself to the current subset, `doc/lang/subset<latest>.md` including its restrictions; a feature outside the subset is not a missing test.
2. Read those specification sections in full.
3. Take each test scenario in `spec.md` and independently derive further cases from each applicable language-specification rule. Check the corpus for relevant error, panic and valid cases. If input partitions are supplied as supporting analysis, check those too; no input-space table is required. Read assertions or goldens before accepting existing coverage, and skip cases already covered. Surface language-specification ambiguities or missing behaviour decisions for the user rather than inventing expected results.
4. Write and run the corpus tests. Derive expected results from the specification by first principles, not from jBallerina or from the implementation's output. Follow the corpus conventions (location, naming, markers, one distinct scenario per test, user-reachable code only) from the `nutcracker-review-guidelines` skill and the repository's `manage-corpus-tests` skill. Each new test must fail without the change. For a case that passes with the change, verify that its behaviour assertion fails without it, not merely setup or an unrelated dependency; report the comparison. Do not add cases that already pass without the change merely to increase coverage.

- Cases that hit an unimplemented or internal error are intentional omissions. Do not report them.
- Do not fix the implementation.

## Findings format

Use these headings, omit empty subsections, and include the full corpus test source under each added or failing case. Passing tests added below must fail without the change. Report untested cases that already pass without the change separately as existing-behaviour coverage gaps, not as tests to add.

```markdown
## Added tests that pass with the change and fail without it
### Error
### Panic
### Valid

## Failing test cases that should work
### Error
### Panic
### Valid

## Existing-behaviour coverage gaps
```

## Input-space partitioning

Derive the relevant input dimensions from the language specification and the changed behavior. Partition each dimension into classes with distinct semantic outcomes, including applicable boundaries and rejection cases.

Identify interactions between dimensions required by the specification or affected by the change. Select representative cases for each partition and relevant interaction, grounding any claimed equivalence in the specification.

Map each case to a meaningful existing assertion or a new test that meets the before/after requirement above. Existing assertions count toward feature coverage even when they pass without the change. If an untested case already passes without the change, record the coverage gap rather than adding a passing baseline test or claiming it is covered. Verify that each added assertion detects breaking the behavior it targets. Report the partitions, interactions, coverage, and justified omissions.

Discover dimensions by tracing the changed rule through the complete relevant caller/callee flow and the applicable syntax in the specification. Include the input's surrounding and nested forms when they can change which rule is applied. Do not treat separate rule paths as equivalent merely because they share a type, output, or helper. For each affected interaction, record all distinct specification rules it can reach and select a witness for each uncovered rule. Tests of individual partitions do not establish coverage of their interactions.

For each pair of input dimensions made dependent by the changed rule, cover every valid combination of their partitions with an existing assertion or a new representative. If an applicable rule jointly depends on more than two dimensions, cover its relevant higher-order combinations. A witness covers only the partition combination it actually exercises; separate witnesses for individual dimensions do not establish combination coverage. Record justified exclusions and equivalences.
