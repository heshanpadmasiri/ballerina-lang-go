#!/usr/bin/env python3
"""Check that new Nutcracker corpus tests fail today, and fail the way the spec predicts.

Reads "<test path>\t<predicted category>" lines on stdin, the output of
`lint_spec.py spec.md --list-new-tests`. Run from the Nutcracker checkout:

  lint_spec.py spec.md --list-new-tests | nutcracker_check_failing.py

The tests run end to end (the integration drivers only) in a temporary worktree
holding HEAD plus the working tree's uncommitted and untracked changes, with
the tests' skip-list entries removed. Each test runs twice:

  1. with -update, which writes its integration golden from today's output
     (the harness cannot check markers without a golden);
  2. without -update, so only the @output, @error and @panic markers decide.

The goldens in the temporary worktree record wrong output and are discarded.
A test that passes, fails differently from its prediction, or cannot be
classified is a problem. Exit status 1 when any problem is found.
"""

import argparse
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

DRIVERS = [
    # (path prefix, driver, subtest name from the remainder, golden path from the remainder)
    ('corpus/bal/', 'TestIntegration', lambda r: r, lambda r: f'corpus/integration/{r[:-4]}.txtar'),
    ('corpus/lib/', 'TestLibIntegration', lambda r: f'lib/{r}', lambda r: f'corpus/integration/lib/{r[:-4]}.txtar'),
    ('corpus/project/', 'TestProjectIntegration', lambda r: r.split('/')[0],
     lambda r: f'corpus/integration/project/{r.split("/")[0]}.txtar'),
    ('corpus/workspace/', 'TestWorkspaceIntegration', lambda r: r.split('/')[0],
     lambda r: f'corpus/integration/workspace/{r.split("/")[0]}.txtar'),
]
CATEGORIES = ['unimplemented', 'wrong-output', 'diagnostic-mismatch', 'panic-mismatch', 'crash']


def run(argv, cwd, check=True):
    result = subprocess.run(argv, cwd=cwd, capture_output=True, text=True)
    if check and result.returncode != 0:
        raise SystemExit(f'{" ".join(argv)} failed:\n{result.stdout}{result.stderr}')
    return result


def skip_key(path):
    if path.startswith('corpus/bal/'):
        return path[len('corpus/bal/'):]
    if path.startswith('corpus/lib/'):
        return path[len('corpus/'):]
    parts = path.split('/')
    if len(parts) >= 3 and parts[1] in ('project', 'workspace'):
        return 'project/' + parts[2]
    return None


def driver_for(path):
    for prefix, driver, name, golden in DRIVERS:
        if path.startswith(prefix):
            rest = path[len(prefix):]
            return driver, name(rest), golden(rest)
    return None


def prepare(repo, tmp):
    run(['git', 'worktree', 'add', '--detach', str(tmp), 'HEAD'], repo)
    diff = run(['git', 'diff', '--binary', 'HEAD'], repo).stdout
    if diff:
        subprocess.run(['git', 'apply', '--whitespace=nowarn'], cwd=tmp, input=diff, text=True, check=True)
    untracked = run(['git', 'ls-files', '--others', '--exclude-standard', '-z'], repo).stdout
    for rel in filter(None, untracked.split('\0')):
        target = tmp / rel
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(repo / rel, target)


def unskip(tmp, keys):
    skip = tmp / 'test_util' / 'skip.go'
    lines = skip.read_text().splitlines(keepends=True)
    kept = [t for t in lines if not ((m := re.match(r'\s*"([^"]+)"', t)) and m.group(1) in keys)]
    skip.write_text(''.join(kept))


def run_pattern(driver, name):
    return '^' + driver + '$/' + '/'.join('^' + re.escape(part) + '$' for part in name.split('/'))


def outcome(output):
    if re.search(r'^\s*--- PASS', output, re.M):
        return 'pass'
    if re.search(r'^\s*--- SKIP', output, re.M):
        return 'skip'
    if re.search(r'^\s*--- FAIL', output, re.M):
        return 'fail'
    return 'none'


def golden_streams(path):
    if not path.is_file():
        return None, None
    streams, current = {}, None
    for line in path.read_text().splitlines(keepends=True):
        m = re.match(r'-- (\S+) --\s*$', line)
        if m:
            current = m.group(1)
            streams[current] = ''
        elif current:
            streams[current] += line
    return streams.get('stdout', ''), streams.get('stderr', '')


def classify(update_out, check_out, stdout, stderr):
    """Return (category, evidence)."""
    text = '\n'.join(filter(None, [update_out, check_out, stdout, stderr]))
    if 'UNIMPLEMENTED' in text and 'fatal[' in text:
        return 'unimplemented', first_line(text, 'fatal[')
    if 'INTERNAL_ERROR' in text:
        return 'crash', first_line(text, 'INTERNAL_ERROR')
    if stdout and re.search(r'^panic: ', stdout, re.M):
        return 'crash', first_line(stdout, 'panic: ')
    if 'has a runtime panic in expected stdout' in update_out:
        return 'crash', first_line(update_out, 'panic')
    rules = [
        ('diagnostic-mismatch', ['leaked a compile diagnostic', 'has empty expected stderr', 'not covered by any',
                                 'produced no error diagnostics', 'is not a compile diagnostic',
                                 'begins with a compile diagnostic']),
        ('panic-mismatch', ['stack frame mismatch', 'runtime error message for @panic',
                            'expected non-zero exit code']),
        ('wrong-output', ['@output annotation mismatch', 'test produced stdout but has no @output']),
    ]
    for category, needles in rules:
        for needle in needles:
            if needle in text:
                return category, first_line(text, needle)
    return 'unknown', tail(text)


def first_line(text, needle):
    return next((t.strip() for t in text.splitlines() if needle in t), '')


def tail(text, n=6):
    return ' | '.join(t.strip() for t in text.strip().splitlines()[-n:])


def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument('--repo', type=Path, help='Nutcracker checkout (default: git top level of the cwd)')
    parser.add_argument('--keep', action='store_true', help='keep the temporary worktree')
    args = parser.parse_args()

    repo = args.repo or Path(run(['git', 'rev-parse', '--show-toplevel'], Path.cwd()).stdout.strip())
    repo = repo.resolve()
    tests = []
    for line in sys.stdin:
        if line.strip():
            path, _, predicted = line.rstrip('\n').partition('\t')
            tests.append((path.strip(), predicted.strip()))
    if not tests:
        print('no tests given on stdin', file=sys.stderr)
        return 1

    tmp = Path(tempfile.mkdtemp(prefix='nutcracker-failcheck-')) / 'worktree'
    problems = []
    try:
        prepare(repo, tmp)
        unskip(tmp, {skip_key(p) for p, _ in tests})
        binary = tmp / 'corpus.test'
        run(['go', 'test', '-c', '-o', str(binary), './corpus'], tmp)
        print(f'{"test":60} {"predicted":20} {"actual":20} evidence')
        for path, predicted in tests:
            target = driver_for(path)
            if target is None:
                problems.append(f'{path}: not a corpus fixture this script can run; check it by hand')
                continue
            driver, name, golden = target
            pattern = run_pattern(driver, name)
            update = run([str(binary), '-test.run', pattern, '-test.v', '-test.count=1', '-update'],
                         tmp / 'corpus', check=False)
            check = run([str(binary), '-test.run', pattern, '-test.v', '-test.count=1'], tmp / 'corpus', check=False)
            update_out = update.stdout + update.stderr
            check_out = check.stdout + check.stderr
            if outcome(update_out) in ('none', 'skip') or outcome(check_out) in ('none', 'skip'):
                problems.append(f'{path}: {driver}/{name} did not run ({outcome(check_out)}); '
                                f'check the path and skip lists')
                continue
            if outcome(check_out) == 'pass' and outcome(update_out) == 'pass':
                actual, evidence = 'passes', 'the markers already hold'
            else:
                stdout, stderr = golden_streams(tmp / golden)
                actual, evidence = classify(update_out if outcome(update_out) == 'fail' else '',
                                            check_out if outcome(check_out) == 'fail' else '', stdout, stderr)
            print(f'{path:60} {predicted:20} {actual:20} {evidence[:160]}')
            if actual == 'passes':
                problems.append(f'{path}: already passes, so it does not test the change')
            elif actual == 'unknown':
                problems.append(f'{path}: failed in a way this script cannot classify: {evidence}')
            elif predicted and actual != predicted:
                problems.append(f'{path}: predicted {predicted}, but it fails with {actual}')
            elif predicted not in CATEGORIES:
                problems.append(f'{path}: no valid prediction ("{predicted}"); it fails with {actual}')
    finally:
        if args.keep:
            print(f'kept {tmp}', file=sys.stderr)
        else:
            run(['git', 'worktree', 'remove', '--force', str(tmp)], repo, check=False)
            shutil.rmtree(tmp.parent, ignore_errors=True)
    for p in problems:
        print(p, file=sys.stderr)
    print(f'{len(problems)} problems', file=sys.stderr)
    return 1 if problems else 0


if __name__ == '__main__':
    sys.exit(main())
