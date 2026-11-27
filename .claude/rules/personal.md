# Global agent instructions

## Nutcracker (ballerina-lang-go, Go module github.com/ballerina-nutcracker/ballerina)

### Running the full test suite

When you need to run the full test suite, do not run `make test` or
`.github/scripts/run_native_tests.py`. They also run the standard-library (`lib`) and
language-server (`ls`) tests. Run this from the repository root instead:

```sh
go test -count=1 -p "$(nproc)" \
  -skip 'TestParseCorpusFiles|TestJBalUnitTests|TestJBalUnitBIRTests|TestLibIntegration' \
  ./ast/... ./bir/... ./birgen/... ./common/... ./context/... ./decimal/... \
  ./desugar/... ./model/... ./nodebuilder/... ./parser/... ./platform/... \
  ./prettyprint/... ./projects/... ./runtime/... ./semantics/... ./semtypes/... \
  ./st/... ./values/... \
  ./corpus/... ./test_util/...
```

- This covers the frontend modules and the `corpus/` tests. It leaves out `lib/...`
  (stdlib and langlib Go packages), `TestLibIntegration` (`corpus/lib`), `ls/...`,
  `cli`, `tools` and `compiler-tools`.
- If `go.work` gains a new compiler or runtime module, add it to the list.
- Run the stdlib or language-server tests only when the user explicitly asks for them,
  or when the change touches `lib/` or `ls/`. In that case, run only the affected
  packages, such as `./lib/...` or `./ls/...`.
