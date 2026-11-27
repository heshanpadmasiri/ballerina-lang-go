"""Tests for the view checks in lint_spec.py: citations and diagrams inside fenced blocks.

Run: python3 -m unittest discover -s design/tests   (from the skills directory)
Mermaid tests are skipped when mmdc is not installed.
"""

import shutil
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path

LINT = Path(__file__).resolve().parent.parent / 'scripts' / 'lint_spec.py'

SOURCE = textwrap.dedent('''\
    package semantics

    func narrow(t Type) Type {
    \tif t.IsNever() {
    \t\treturn t
    \t}
    \treturn t.Diff(nilType)
    }
''')


def rca(trace, fix='Change `semantics/narrow.go:4` `if t.IsNever() {`.', reproducer='Program.'):
    return textwrap.dedent('''\
        # Root cause: narrowing
        ## Reproducer
        {reproducer}
        ## Trace
        {trace}
        ## Root cause
        The check `semantics/narrow.go:4` `if t.IsNever() {` returns early.
        ## Rejected causes
        None.
        ## Inputs sharing the cause
        None.
        ## Agreed fix
        {fix}
    ''').replace('{trace}', trace).replace('{fix}', fix).replace('{reproducer}', reproducer)


class LintViews(unittest.TestCase):
    def setUp(self):
        self.repo = Path(tempfile.mkdtemp(prefix='lint-spec-test-'))
        self.addCleanup(shutil.rmtree, self.repo)
        (self.repo / 'semantics').mkdir()
        (self.repo / 'semantics' / 'narrow.go').write_text(SOURCE)

    def lint(self, text, *flags):
        doc = self.repo / 'rca.md'
        doc.write_text(text)
        result = subprocess.run([sys.executable, str(LINT), str(doc), '--repo', str(self.repo), *flags],
                                capture_output=True, text=True)
        return result.returncode, result.stderr

    def test_valid_tree_passes(self):
        code, out = self.lint(rca('```text\nparse ok\nsemantics `semantics/narrow.go:4` `if t.IsNever() {`\n```'))
        self.assertEqual(code, 0, out)

    def test_wrong_excerpt_in_tree(self):
        code, out = self.lint(rca('```text\nsemantics `semantics/narrow.go:4` `if t.IsNil() {`\n```'))
        self.assertEqual(code, 1)
        self.assertIn('excerpt not found within two lines of semantics/narrow.go:4', out)

    def test_bare_location_missing_file(self):
        code, out = self.lint(rca('```\ndesugar  desugar/expr.go:10\n```'))
        self.assertIn('cited file does not exist: desugar/expr.go', out)

    def test_bare_location_past_end(self):
        code, out = self.lint(rca('```text\nsemantics  semantics/narrow.go:99\n```'))
        self.assertIn('semantics/narrow.go:99 is past the end of the file', out)

    def test_bare_location_valid(self):
        code, out = self.lint(rca('```text\nsemantics  semantics/narrow.go:7  (P1)\n```\n## Probes\n'
                                  '### P1: x\nCommand: `true`\nOutput:\n```text\nok\n```'))
        self.assertEqual(code, 0, out)

    def test_undefined_probe_in_view(self):
        code, out = self.lint(rca('```text\nsemantics ✗ (P3)\n```'))
        self.assertIn('probe P3 is referenced but not defined', out)

    def test_diff_added_lines_not_checked(self):
        diff = '```diff\n semantics `semantics/narrow.go:4` `if t.IsNever() {`\n+  semantics/widen.go:3\n```'
        code, out = self.lint(rca('Trace.', fix=diff))
        self.assertEqual(code, 0, out)

    def test_diff_removed_lines_checked(self):
        code, out = self.lint(rca('Trace.', fix='```diff\n-  semantics/gone.go:3\n```'))
        self.assertIn('cited file does not exist: semantics/gone.go', out)

    def test_code_and_output_blocks_ignored(self):
        reproducer = '```text\npanic at semantics/other.go:12\n```'
        trace = '```go\nx := "a/b.go:3"\n```\n```text\ngithub.com/ballerina-nutcracker/ballerina/x.go:3\n```'
        code, out = self.lint(rca(trace, reproducer=reproducer))
        self.assertEqual(code, 0, out)


@unittest.skipUnless(shutil.which('mmdc'), 'mmdc not installed')
class LintMermaid(LintViews.__base__):
    setUp = LintViews.setUp
    lint = LintViews.lint

    def test_valid_mermaid_passes(self):
        code, out = self.lint(rca('```mermaid\nflowchart LR\n  parse --> desugar\n```'))
        self.assertEqual(code, 0, out)

    def test_invalid_mermaid_reports_document_line(self):
        doc = rca('```mermaid\nflowchart LR\n  A[call f(x)] --> B\n```')
        code, out = self.lint(doc)
        self.assertEqual(code, 1)
        line = doc.splitlines().index('  A[call f(x)] --> B') + 1
        self.assertIn(f'rca.md:{line}: mermaid: Error: Parse error', out)

    def test_no_mermaid_flag_skips(self):
        code, out = self.lint(rca('```mermaid\nflowchart LR\n  A[call f(x)] --> B\n```'), '--no-mermaid')
        self.assertEqual(code, 0, out)


if __name__ == '__main__':
    unittest.main()
