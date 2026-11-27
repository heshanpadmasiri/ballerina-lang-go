#!/usr/bin/env python3
"""Check the structure and references of spec.md, research.md and rca.md.

Formats: ../references/spec-format.md and ../../fix/references/rca-format.md. Exit status 1 when any problem is found.

  lint_spec.py DOC                      check DOC
  lint_spec.py spec.md --list-new-tests print "<test>\t<Fails today>" for each staged new test
  lint_spec.py spec.md --stamp          record the approved new tests' hashes in the spec
  lint_spec.py spec.md --guard          check approved tests and the pending block are unchanged
  lint_spec.py spec.md --guard --slice-done N
                                        also check slices 1..N no longer have pending tests
  lint_spec.py DOC --rerun-probes       rerun every probe and report output differences
  lint_spec.py DOC --no-mermaid         skip rendering Mermaid diagrams

Citations are checked in prose and in views: fenced blocks tagged text, diff or mermaid, or
untagged, outside ## Probes and ## Reproducer. Mermaid diagrams are rendered with
../../show-me/scripts/check_mermaid.py.
"""

import argparse
import hashlib
import os
import re
import subprocess
import sys
from pathlib import Path

SECTIONS = {
    'spec': ['Goal', 'Non-goals', 'Repository context', 'Guardrails', 'API', 'Design',
             'Test scenarios', 'Slices', 'Success criteria'],
    'research': ['How it works today', 'Specification', 'Files read', 'Gaps', 'Questions asked'],
    'rca': ['Reproducer', 'Trace', 'Root cause', 'Rejected causes', 'Inputs sharing the cause',
            'Agreed fix'],
}
OPTIONAL_SECTIONS = {'research': {'Specification'}}
TRAILING_SECTIONS = ['Probes', 'Approved tests']
CATEGORIES = {'unimplemented', 'wrong-output', 'diagnostic-mismatch', 'panic-mismatch', 'crash'}
NUTCRACKER_MODULE = 'module github.com/ballerina-nutcracker/ballerina'
SKIP_FILE = 'test_util/skip.go'

CITATION = re.compile(r'(?<!`)(`{1,2})(?P<path>[^`\s:]+):(?P<line>\d+)\1(?!`)(?P<rest>[ \t]*(?P<tick>`{1,2})(?P<excerpt>.+?)(?P=tick)(?!`))?')
PROBE_REF = re.compile(r'\(P(\d+)\)')
PROBE_DEF = re.compile(r'^### P(\d+):')
PLACEHOLDER = re.compile(r'\bTODO\b|\bTBD\b|\bFIXME\b|<[a-z][a-z -]*>')
INLINE_CODE = re.compile(r'(`+).*?\1')
# A location written without backticks inside a view, e.g. `semantics/narrow.go:142`.
BARE_LOCATION = re.compile(r'(?<![\w/.`-])(?P<path>[\w.-]+(?:/[\w.-]+)+\.\w+):(?P<line>\d+)\b')
VIEW_LANGUAGES = {'', 'text', 'diff', 'mermaid'}
OUTPUT_SECTIONS = {'Probes', 'Reproducer'}
SHOW_ME_SCRIPTS = Path(__file__).resolve().parent.parent.parent / 'show-me' / 'scripts'


class Doc:
    def __init__(self, path):
        self.path = path
        self.lines = path.read_text().splitlines()
        self.prose = []  # (line number, text) outside fenced code
        self.fences = []  # (line of the opening fence, language, [(n, text)])
        fence = None
        for n, text in enumerate(self.lines, 1):
            stripped = text.lstrip()
            marker = re.match(r'(`{3,}|~{3,})', stripped)
            if fence:
                if marker and marker.group(1)[0] == fence[0] and len(marker.group(1)) >= len(fence) \
                        and stripped.strip() == marker.group(1):
                    fence = None
                else:
                    self.fences[-1][2].append((n, text))
                continue
            if marker:
                fence = marker.group(1)
                self.fences.append((n, stripped[len(fence):].strip().split(' ')[0].lower(), []))
                continue
            self.prose.append((n, text))
        self.sections = []  # (title, line, [(n, text)])
        for n, text in self.prose:
            if text.startswith('## '):
                self.sections.append((text[3:].strip(), n, []))
            elif self.sections:
                self.sections[-1][2].append((n, text))

    def section(self, title):
        return next((s for s in self.sections if s[0] == title), None)

    def section_at(self, line):
        """Title of the H2 section containing a line, or None before the first."""
        return next((t for t, n, _ in reversed(self.sections) if n < line), None)

    def views(self):
        """Fenced blocks that show the code's shape: trees, diffs and diagrams.

        Program sources and recorded output are not views: blocks under ## Probes and
        ## Reproducer, and blocks in a programming language, are left out."""
        return [f for f in self.fences
                if f[1] in VIEW_LANGUAGES and self.section_at(f[0]) not in OUTPUT_SECTIONS]

    def blocks(self, title):
        """H3 blocks inside an H2 section: (heading, line, [(n, text)])."""
        section = self.section(title)
        if not section:
            return []
        out = []
        for n, text in section[2]:
            if text.startswith('### '):
                out.append((text[4:].strip(), n, []))
            elif out:
                out[-1][2].append((n, text))
        return out


class Linter:
    def __init__(self, doc, kind, repo):
        self.doc, self.kind, self.repo = doc, kind, repo
        self.problems = []

    def fail(self, line, message):
        self.problems.append(f'{self.doc.path}:{line}: {message}')

    # ----- shared checks -----

    def check_sections(self):
        expected = SECTIONS[self.kind]
        titles = [s[0] for s in self.doc.sections]
        cursor = 0
        for title in expected:
            if title not in titles:
                if title not in OPTIONAL_SECTIONS.get(self.kind, set()):
                    self.fail(1, f'missing section "## {title}"')
                continue
            at = titles.index(title)
            if at < cursor:
                self.fail(self.doc.section(title)[1], f'section "## {title}" is out of order')
            cursor = max(cursor, at)
        for title, n, _ in self.doc.sections:
            if title not in expected and title not in TRAILING_SECTIONS:
                self.fail(n, f'unexpected section "## {title}"')
        if not any(t.startswith('# ') for _, t in self.doc.prose):
            self.fail(1, 'no H1 title')

    def check_placeholders(self):
        for n, text in self.doc.prose:
            if PLACEHOLDER.search(INLINE_CODE.sub('', text)):
                self.fail(n, f'placeholder left in text: {text.strip()}')

    def check_citations(self, lines=None):
        for n, text in self.doc.prose if lines is None else lines:
            for m in CITATION.finditer(text):
                path, line = m.group('path'), int(m.group('line'))
                if '/' not in path and '.' not in path:
                    continue
                source = self.source(n, path, line)
                if source is None:
                    continue
                if not m.group('excerpt'):
                    self.fail(n, f'citation {path}:{line} has no excerpt')
                    continue
                excerpt = normalise(m.group('excerpt'))
                window = source[max(0, line - 3):line + 2]
                if not any(excerpt in normalise(s) for s in window):
                    self.fail(n, f'excerpt not found within two lines of {path}:{line}: {m.group("excerpt")}')

    def source(self, n, path, line):
        """The cited file's lines, or None after reporting a missing file or line."""
        target = self.repo / path
        if not target.is_file():
            self.fail(n, f'cited file does not exist: {path}')
            return None
        source = target.read_text(errors='replace').splitlines()
        if not 1 <= line <= len(source):
            self.fail(n, f'citation {path}:{line} is past the end of the file ({len(source)} lines)')
            return None
        return source

    def check_views(self, mermaid=True):
        """Locations, citations and probe references inside trees, diffs and diagrams.

        A backticked citation in a view is checked as in prose, excerpt included. A bare
        `path:line` must name an existing file and line. In a diff, `+` lines show the
        proposed shape and are not checked against today's code."""
        diagrams = []
        for start, lang, body in self.doc.views():
            lines = [(n, t) for n, t in body if not (lang == 'diff' and t.startswith('+'))]
            self.check_citations(lines)
            for n, text in lines:
                for m in BARE_LOCATION.finditer(INLINE_CODE.sub('', text)):
                    path = m.group('path')
                    if '.' not in path.split('/')[0]:  # skip module paths such as github.com/...
                        self.source(n, path, int(m.group('line')))
            self.check_probe_refs(body)
            if lang == 'mermaid':
                diagrams.append((start, body))
        if diagrams and mermaid:
            self.check_mermaid(diagrams)

    def check_mermaid(self, diagrams):
        sys.path.insert(0, str(SHOW_ME_SCRIPTS))
        try:
            import check_mermaid
        except ImportError:
            self.fail(diagrams[0][0], f'cannot check Mermaid: {SHOW_ME_SCRIPTS}/check_mermaid.py not found')
            return
        finally:
            sys.path.pop(0)
        # Keep each diagram on its own lines so mmdc's line numbers map back to the document.
        text = [''] * len(self.doc.lines)
        for start, body in diagrams:
            text[start - 1] = '```mermaid'
            for n, t in body:
                text[n - 1] = t
            text[(body[-1][0] if body else start)] = '```'
        try:
            problems, _ = check_mermaid.check_texts([(str(self.doc.path), '\n'.join(text))])
        except check_mermaid.Unavailable as e:
            self.fail(diagrams[0][0], f'cannot check Mermaid, so the diagrams are unverified: {e}')
            return
        for p in problems:
            self.problems.append(p)

    def probes(self, report=True):
        """Probe definitions: number -> (line, command, output)."""
        defs = {}
        lines = self.doc.lines
        heads = [(i, int(m.group(1))) for i, t in enumerate(lines) if (m := PROBE_DEF.match(t))]
        for idx, (i, number) in enumerate(heads):
            end = heads[idx + 1][0] if idx + 1 < len(heads) else len(lines)
            body = lines[i + 1:end]
            command = next((re.match(r'Command:\s*`(.+)`\s*$', t).group(1) for t in body
                            if re.match(r'Command:\s*`(.+)`\s*$', t)), None)
            output = None
            for j, t in enumerate(body):
                if t.strip().startswith('Output:'):
                    rest = body[j + 1:]
                    if rest and re.match(r'\s*(`{3,}|~{3,})', rest[0]):
                        fence = re.match(r'\s*(`{3,}|~{3,})', rest[0]).group(1)
                        out = []
                        for t2 in rest[1:]:
                            if t2.strip() == fence:
                                break
                            out.append(t2)
                        output = '\n'.join(out)
                    break
            if number in defs and report:
                self.fail(i + 1, f'probe P{number} is defined twice')
            defs[number] = (i + 1, command, output)
        return defs

    def check_probes(self):
        defs = self.probes()
        if defs and not self.doc.section('Probes'):
            self.fail(1, 'probes are defined outside a "## Probes" section')
        for number, (line, command, output) in defs.items():
            if command is None:
                self.fail(line, f'probe P{number} has no "Command: `...`" line')
            if output is None:
                self.fail(line, f'probe P{number} has no fenced "Output:" block')
        self.check_probe_refs(self.doc.prose, defs)

    def check_probe_refs(self, lines, defs=None):
        defs = self.probes(report=False) if defs is None else defs
        for n, text in lines:
            for m in PROBE_REF.finditer(text):
                if int(m.group(1)) not in defs:
                    self.fail(n, f'probe P{m.group(1)} is referenced but not defined')

    def rerun_probes(self):
        for number, (line, command, output) in sorted(self.probes().items()):
            if command is None or output is None:
                continue
            result = subprocess.run(command, shell=True, cwd=self.repo, capture_output=True, text=True)
            actual = (result.stdout + result.stderr).strip()
            if normalise(actual) != normalise(output) and normalise(output) not in normalise(actual):
                self.fail(line, f'probe P{number} output changed; now:\n{actual}')

    # ----- spec checks -----

    def scenarios(self):
        out = {}
        for heading, n, body in self.doc.blocks('Test scenarios'):
            m = re.match(r'S(\d+):', heading)
            if not m:
                self.fail(n, f'scenario heading is not "### S<n>: <title>": {heading}')
                continue
            fields = {}
            for _, text in body:
                fm = re.match(r'\s*-\s*([A-Z][A-Za-z ]+):\s*(.*)$', text)
                if fm:
                    fields[fm.group(1).strip()] = fm.group(2).strip()
            out[int(m.group(1))] = (n, fields)
        return out

    def slices(self):
        out = {}
        for heading, n, body in self.doc.blocks('Slices'):
            m = re.match(r'Slice (\d+):', heading)
            if not m:
                self.fail(n, f'slice heading is not "### Slice <n>: <behaviour>": {heading}')
                continue
            fields = {}
            for _, text in body:
                fm = re.match(r'\s*-\s*([A-Z][A-Za-z ]+):\s*(.*)$', text)
                if fm:
                    fields[fm.group(1).strip()] = fm.group(2).strip()
            out[int(m.group(1))] = (n, fields)
        return out

    def parse_test(self, value):
        m = re.match(r'`([^`]+)`\s*\((new|existing|new, slice (\d+))\)\s*$', value)
        if not m:
            return None
        return m.group(1), 'existing' if m.group(2) == 'existing' else 'new', m.group(3)

    def new_tests(self):
        """(scenario, path, fails_today, slice, staged) for each new test."""
        out = []
        for number, (_, fields) in sorted(self.scenarios().items()):
            test = self.parse_test(fields.get('Test', ''))
            if not test or test[1] != 'new':
                continue
            path, _, late_slice = test
            staged = late_slice is None
            slice_no = fields.get('Slice') if staged else late_slice
            out.append((number, path, fields.get('Fails today'), slice_no, staged))
        return out

    def check_spec(self):
        scenarios, slices = self.scenarios(), self.slices()
        if not scenarios:
            self.fail(1, 'no scenarios under "## Test scenarios"')
        if not slices:
            self.fail(1, 'no slices under "## Slices"')
        if sorted(slices) != list(range(1, len(slices) + 1)):
            self.fail(1, f'slices are numbered {sorted(slices)}, expected 1..{len(slices)}')
        owned = {s: [] for s in slices}
        for number, (n, fields) in sorted(scenarios.items()):
            for key in ('Rule', 'Test', 'Expected'):
                if not fields.get(key):
                    self.fail(n, f'S{number} has no "{key}"')
            test = self.parse_test(fields.get('Test', ''))
            if fields.get('Test') and not test:
                self.fail(n, f'S{number} Test is not "`<test>` (new)", "(existing)" or "(new, slice <n>)"')
                continue
            if not test:
                continue
            path, status, late_slice = test
            if status == 'existing' or late_slice is None:
                if not self.test_exists(path):
                    self.fail(n, f'S{number} test does not exist: {path}')
            if status == 'existing':
                for key in ('Fails today', 'Slice'):
                    if key in fields:
                        self.fail(n, f'S{number} is an existing test and must not have "{key}"')
                if self.nutcracker and self.unsupported(path):
                    self.fail(n, f'S{number} existing test is skipped in {SKIP_FILE}: {path}')
                continue
            if late_slice is not None:
                if not path.startswith('corpus/recovery/'):
                    self.fail(n, f'S{number} "(new, slice <n>)" is only for tests that cannot be skipped')
                if 'Fails today' in fields:
                    self.fail(n, f'S{number} is written by its slice and must not have "Fails today"')
                slice_no = late_slice
            else:
                if fields.get('Fails today') not in CATEGORIES:
                    self.fail(n, f'S{number} "Fails today" must be one of {", ".join(sorted(CATEGORIES))}')
                slice_no = fields.get('Slice')
                if not slice_no:
                    self.fail(n, f'S{number} has no "Slice"')
                    continue
            if not slice_no.isdigit() or int(slice_no) not in slices:
                self.fail(n, f'S{number} names slice {slice_no}, which does not exist')
                continue
            owned[int(slice_no)].append(number)
        for number, (n, fields) in sorted(slices.items()):
            listed = {int(x) for x in re.findall(r'S(\d+)', fields.get('Scenarios', ''))}
            if not listed:
                self.fail(n, f'Slice {number} lists no scenarios')
            if listed != set(owned.get(number, [])):
                self.fail(n, f'Slice {number} lists {fmt(listed)} but the scenarios naming it are '
                             f'{fmt(owned.get(number, []))}')
            depends = fields.get('Depends on')
            if depends is None:
                self.fail(n, f'Slice {number} has no "Depends on"')
            else:
                for d in map(int, re.findall(r'(\d+)', depends)):
                    if d >= number:
                        self.fail(n, f'Slice {number} depends on slice {d}, which is not earlier')
        criteria = self.doc.section('Success criteria')
        referenced = set()
        if criteria:
            items = [(n, t) for n, t in criteria[2] if re.match(r'\s*[-*] ', t)]
            if not items:
                self.fail(criteria[1], 'no success criteria')
            for n, text in items:
                refs = {int(x) for x in re.findall(r'\bS(\d+)\b', text)}
                if not refs:
                    self.fail(n, 'success criterion names no scenario')
                for r in refs - set(scenarios):
                    self.fail(n, f'success criterion names S{r}, which does not exist')
                referenced |= refs
        for number in sorted(set(scenarios) - referenced):
            self.fail(scenarios[number][0], f'S{number} supports no success criterion')
        if self.nutcracker:
            self.check_pending()

    # ----- Nutcracker adapter -----

    @property
    def nutcracker(self):
        go_mod = self.repo / 'go.mod'
        return go_mod.is_file() and NUTCRACKER_MODULE in go_mod.read_text().splitlines()

    def test_exists(self, path):
        return (self.repo / path).exists() or bool(re.search(r'\bTest\w+', path))

    def skip_key(self, path):
        parts = path.split('/')
        if path.startswith('corpus/bal/'):
            return path[len('corpus/bal/'):]
        if path.startswith('corpus/lib/'):
            return path[len('corpus/'):]
        if len(parts) >= 3 and parts[0] == 'corpus' and parts[1] in ('project', 'workspace'):
            return 'project/' + parts[2]
        return None

    def skip_file_entries(self):
        target = self.repo / SKIP_FILE
        return target.read_text().splitlines() if target.is_file() else []

    def unsupported(self, path):
        key = self.skip_key(path)
        return key is not None and any(
            re.match(r'\s*"([^"]+)"', t) and re.match(r'\s*"([^"]+)"', t).group(1) == key
            for t in self.skip_file_entries())

    def pending_entries(self):
        title = self.title()
        lines = self.skip_file_entries()
        begin = f'// BEGIN pending: {title}'
        end = f'// END pending: {title}'
        starts = [i for i, t in enumerate(lines) if t.strip() == begin]
        if not starts:
            return None
        entries = []
        for t in lines[starts[0] + 1:]:
            if t.strip() == end:
                return entries
            m = re.match(r'\s*"([^"]+)"', t)
            if m:
                entries.append(m.group(1))
        self.fail(1, f'{SKIP_FILE}: "{begin}" has no matching "{end}"')
        return entries

    def title(self):
        return next((t[2:].strip() for _, t in self.doc.prose if t.startswith('# ')), '')

    def check_pending(self, require_block=True):
        staged = {self.skip_key(p): (s, p) for s, p, _, _, st in self.new_tests() if st and self.skip_key(p)}
        entries = self.pending_entries()
        if entries is None:
            if staged and require_block:
                self.fail(1, f'{SKIP_FILE} has no "// BEGIN pending: {self.title()}" block for the new tests')
            return
        for key in entries:
            if key not in staged:
                self.fail(1, f'{SKIP_FILE}: pending entry "{key}" is not a new test of this spec')
        return entries, staged

    # ----- approval stamp and guard -----

    def digest(self, path):
        target = self.repo / path
        h = hashlib.sha256()
        files = sorted(p for p in target.rglob('*') if p.is_file()) if target.is_dir() else [target]
        for f in files:
            h.update(str(f.relative_to(self.repo)).encode())
            h.update(f.read_bytes())
        return h.hexdigest()

    def stamp(self):
        tests = [p for _, p, _, _, staged in self.new_tests() if staged]
        missing = [p for p in tests if not (self.repo / p).exists()]
        if missing:
            for p in missing:
                self.fail(1, f'cannot stamp missing test: {p}')
            return
        text = self.doc.path.read_text()
        text = re.sub(r'\n## Approved tests\n.*\Z', '\n', text, flags=re.S).rstrip('\n')
        block = '\n'.join(f'- `{self.digest(p)}` `{p}`' for p in tests)
        self.doc.path.write_text(f'{text}\n\n## Approved tests\n\n{block}\n')
        print(f'stamped {len(tests)} tests in {self.doc.path}')

    def guard(self, slice_done):
        section = self.doc.section('Approved tests')
        if not section:
            self.fail(1, 'no "## Approved tests" section; the spec has not been stamped')
            return
        approved = {}
        for n, text in section[2]:
            m = re.match(r'\s*-\s*`([0-9a-f]{64})`\s*`([^`]+)`', text)
            if m:
                approved[m.group(2)] = (n, m.group(1))
        current = {p for _, p, _, _, staged in self.new_tests() if staged}
        for path in sorted(current - set(approved)):
            self.fail(1, f'new test {path} was added to the spec after approval')
        for path, (n, digest) in sorted(approved.items()):
            if path not in current:
                self.fail(n, f'approved test {path} was removed from the spec')
            elif not (self.repo / path).exists():
                self.fail(n, f'approved test {path} no longer exists')
            elif self.digest(path) != digest:
                self.fail(n, f'approved test {path} was modified')
        if not self.nutcracker:
            return
        result = self.check_pending(require_block=slice_done is None)
        if result is None or slice_done is None:
            return
        entries, staged = result
        slice_of = {self.skip_key(p): int(sl) for _, p, _, sl, st in self.new_tests()
                    if st and self.skip_key(p) and sl and sl.isdigit()}
        for key in entries:
            if slice_of.get(key, 0) <= slice_done:
                self.fail(1, f'{SKIP_FILE}: "{key}" belongs to slice {slice_of.get(key)} '
                             f'but is still pending after slice {slice_done}')


def normalise(text):
    return re.sub(r'\s+', ' ', text).strip()


def fmt(numbers):
    return ', '.join(f'S{n}' for n in sorted(numbers)) or 'none'


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument('doc', type=Path)
    parser.add_argument('--kind', choices=sorted(SECTIONS))
    parser.add_argument('--repo', type=Path, help='repository root (default: git top level of the cwd)')
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument('--list-new-tests', action='store_true')
    mode.add_argument('--stamp', action='store_true')
    mode.add_argument('--guard', action='store_true')
    parser.add_argument('--slice-done', type=int)
    parser.add_argument('--rerun-probes', action='store_true')
    parser.add_argument('--no-mermaid', action='store_true',
                        help='skip rendering Mermaid diagrams (they stay unverified)')
    args = parser.parse_args()

    kind = args.kind or args.doc.stem
    if kind not in SECTIONS:
        parser.error(f'cannot infer the document kind from {args.doc.name}; pass --kind')
    if (args.list_new_tests or args.stamp or args.guard) and kind != 'spec':
        parser.error('--list-new-tests, --stamp and --guard apply to spec.md only')
    if args.slice_done is not None and not args.guard:
        parser.error('--slice-done requires --guard')
    repo = args.repo
    if repo is None:
        top = subprocess.run(['git', 'rev-parse', '--show-toplevel'], capture_output=True, text=True)
        repo = Path(top.stdout.strip()) if top.returncode == 0 else Path.cwd()
    linter = Linter(Doc(args.doc), kind, repo.resolve())

    if args.list_new_tests:
        for _, path, fails_today, _, staged in linter.new_tests():
            if staged:
                print(f'{path}\t{fails_today or ""}')
        return 0
    if args.stamp:
        linter.check_sections()
        linter.check_spec()
        if not linter.problems:
            linter.stamp()
    elif args.guard:
        linter.guard(args.slice_done)
    else:
        linter.check_sections()
        linter.check_placeholders()
        linter.check_citations()
        linter.check_probes()
        linter.check_views(mermaid=not args.no_mermaid)
        if kind == 'spec':
            linter.check_spec()
        if args.rerun_probes:
            linter.rerun_probes()
    for p in linter.problems:
        print(p, file=sys.stderr)
    print(f'{len(linter.problems)} problems', file=sys.stderr)
    return 1 if linter.problems else 0


if __name__ == '__main__':
    sys.exit(main())
