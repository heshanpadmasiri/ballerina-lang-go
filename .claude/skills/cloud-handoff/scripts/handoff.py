#!/usr/bin/env python3
"""Move uncommitted work between a local checkout and a Claude Code cloud session.

The only channel is a branch on the repository's remote: cloud/<feature>/<seq>.
Each branch tip is one handoff commit on top of the real work. It carries:

  - every uncommitted change in the worktree (spec.md, new tests, the pending block,
    progress.md, slice-report.md, ...),
  - the agent tooling the cloud does not have: skills, personal rules, settings, the CLAUDE.md link,
  - .claude/handoff.json: feature, seq, base, parent, status and slice.

Branches are never rewritten; every push creates the next seq. Commits between the base and
HEAD that are not on the remote yet (local touch-ups) are pushed with the handoff commit.

  handoff.py push   --feature F [--status S --slice N]  build the handoff commit and push it
  handoff.py launch --feature F                         start a cloud session on the latest seq
  handoff.py watch  --feature F --after SEQ             print one line when a later seq appears
  handoff.py land   --feature F                         fast-forward and restore the worktree
  handoff.py resume --expect SHA                        (cloud) check the start commit, restore
  handoff.py clean  --feature F                         delete the cloud/F/* branches

push never modifies the worktree or index: the commit is built in a temporary index.
"""

import argparse
import datetime
import json
import os
import re
import subprocess
import sys
import tempfile
import time
from pathlib import Path

SUBJECT = 'chore(handoff): {feature} #{seq} (drop before PR)'
SUBJECT_RE = re.compile(r'^chore\(handoff\): (\S+) #(\d+) \(drop before PR\)$')
MANIFEST = '.claude/handoff.json'
TOOLING_EXCLUDES = (':(exclude).claude', ':(exclude)CLAUDE.md')
# Design documents never committed; hidden from `git status` in the cloud so they are not staged.
SCRATCH = ('spec.md', 'research.md', 'progress.md', 'rca.md', 'slice-report.md')
STATUSES = ('handoff', 'slice-ready', 'blocked')
DEFAULT_SKILLS = (
    'cloud-handoff', 'implement', 'design', 'review-loop', 'review-oracle', 'commit',
    'fix-up-target', 'test-oracle', 'show-me', 'rebase', 'resolve-conflicts',
    'nutcracker-behavior-policy', 'nutcracker-review-guidelines', 'nutcracker-test-oracle',
)
# Cloud sessions do not read user settings; without this every commit gets attribution trailers.
CLOUD_SETTINGS = {'attribution': {'commit': '', 'pr': ''}, 'includeCoAuthoredBy': False}
HOME_SKILLS = Path.home() / '.claude' / 'skills'
HOME_RULES = Path.home() / '.claude' / 'CLAUDE.md'
SKILL_DIR = Path(__file__).resolve().parent.parent
IN_CLOUD = bool(os.environ.get('CLAUDE_CODE_REMOTE'))


def die(msg):
    print(f'error: {msg}', file=sys.stderr)
    sys.exit(1)


def git(*args, env=None, check=True, cwd=None):
    r = subprocess.run(['git', *args], capture_output=True, text=True, env=env, cwd=cwd)
    if check and r.returncode != 0:
        die(f'git {" ".join(args)}: {r.stderr.strip() or r.stdout.strip()}')
    return r.stdout.strip()


def ok(*args):
    return subprocess.run(['git', *args], capture_output=True).returncode == 0


def blob_of(text):
    return subprocess.run(['git', 'hash-object', '-w', '--stdin'], input=text,
                          capture_output=True, text=True, check=True).stdout.strip()


def root():
    return Path(git('rev-parse', '--show-toplevel'))


def remote_seqs(remote, feature):
    out = git('ls-remote', '--heads', remote, f'refs/heads/cloud/{feature}/*')
    seqs = {}
    for line in out.splitlines():
        sha, ref = line.split('\t')
        tail = ref.rsplit('/', 1)[1]
        if tail.isdigit():
            seqs[int(tail)] = sha
    return seqs


def fetch_seq(remote, feature, seq):
    branch = f'cloud/{feature}/{seq}'
    git('fetch', '-q', remote, f'+refs/heads/{branch}:refs/remotes/{remote}/{branch}')
    return git('rev-parse', f'refs/remotes/{remote}/{branch}')


def pick_seq(a):
    seqs = remote_seqs(a.remote, a.feature)
    if not seqs:
        die(f'no cloud/{a.feature}/* branch on {a.remote}')
    seq = a.seq if a.seq is not None else max(seqs)
    if seq not in seqs:
        die(f'cloud/{a.feature}/{seq} does not exist')
    return seq, fetch_seq(a.remote, a.feature, seq)


def manifest_at(commit):
    r = subprocess.run(['git', 'show', f'{commit}:{MANIFEST}'], capture_output=True, text=True)
    if r.returncode != 0:
        die(f'{commit[:12]} has no {MANIFEST}; it is not a handoff commit')
    return json.loads(r.stdout)


def check_handoff(commit):
    subject = git('log', '-1', '--format=%s', commit)
    if not SUBJECT_RE.match(subject):
        die(f'{commit[:12]} is not a handoff commit (subject: {subject!r})')
    man = manifest_at(commit)
    parent = git('rev-parse', f'{commit}^')
    if man['parent'] != parent:
        die(f'handoff {commit[:12]} records parent {man["parent"][:12]} but sits on {parent[:12]}')
    if not ok('merge-base', '--is-ancestor', man['base'], parent):
        die(f'base {man["base"][:12]} is not an ancestor of {parent[:12]}')
    return man, parent


def worktree_tree(repo):
    """The tree of HEAD plus every uncommitted change and the scratch files, without touching the index."""
    fd, index = tempfile.mkstemp(prefix='handoff-index-')
    os.close(fd)
    env = dict(os.environ, GIT_INDEX_FILE=index)
    try:
        git('read-tree', 'HEAD', env=env)
        git('add', '-A', '--', '.', env=env)  # ignored files excluded
        scratch = [p for p in SCRATCH if (repo / p).is_file()]
        if scratch:
            git('add', '-f', '--', *scratch, env=env)  # resume hides them from git in the cloud
        return index, env
    except BaseException:
        os.unlink(index)
        raise


# ----- push -----

def tooling_files(repo, skills):
    """(path in repo, source file or symlink target, mode) for the injected tooling."""
    files = []
    if IN_CLOUD and (repo / MANIFEST).exists():  # a resumed cloud worktree already has the tooling
        for p in sorted((repo / '.claude').rglob('*')):
            rel = p.relative_to(repo).as_posix()
            if p.is_file() and not p.is_symlink() and rel != MANIFEST:
                files.append((rel, p, '100755' if os.access(p, os.X_OK) else '100644'))
    else:
        for name in skills:
            src = SKILL_DIR if name == 'cloud-handoff' else (HOME_SKILLS / name).resolve()
            if not (src / 'SKILL.md').exists():
                die(f'skill {name} not found in {HOME_SKILLS}')
            add_tree(files, src, f'.claude/skills/{name}')
        repo_skills = repo / '.agents' / 'skills'
        if repo_skills.is_dir():
            for d in sorted(repo_skills.iterdir()):
                if (d / 'SKILL.md').exists() and d.name not in skills:
                    add_tree(files, d, f'.claude/skills/{d.name}')
        if HOME_RULES.exists():
            text = HOME_RULES.read_text().replace('$(sysctl -n hw.ncpu)', '$(nproc)')  # the cloud runs Linux
            files.append(('.claude/rules/personal.md', text, 'text'))
        files.append(('.claude/rules/cloud-handoff.md', SKILL_DIR / 'references' / 'cloud-rules.md', '100644'))
        files.append(('.claude/settings.json', json.dumps(CLOUD_SETTINGS, indent=2) + '\n', 'text'))
    if (repo / 'AGENTS.md').exists() and not ok('-C', str(repo), 'ls-files', '--error-unmatch', 'CLAUDE.md'):
        files.append(('CLAUDE.md', 'AGENTS.md', '120000'))
    return files


def add_tree(files, src, dest):
    for p in sorted(src.rglob('*')):
        if p.is_dir() or '__pycache__' in p.parts or p.name == '.DS_Store':
            continue
        real = p.resolve()
        if real.is_file():
            files.append((f'{dest}/{p.relative_to(src).as_posix()}', real,
                          '100755' if os.access(real, os.X_OK) else '100644'))


def build_commit(repo, man, skills):
    index, env = worktree_tree(repo)
    try:
        for rel, src, mode in tooling_files(repo, skills):
            if mode == '120000':
                blob, mode = blob_of(src), '120000'
            elif mode == 'text':
                blob, mode = blob_of(src), '100644'
            else:
                blob = git('hash-object', '-w', '--', str(src))
            git('update-index', '--add', '--cacheinfo', f'{mode},{blob},{rel}', env=env)
        git('update-index', '--add', '--cacheinfo', f'100644,{blob_of(json.dumps(man, indent=2) + chr(10))},{MANIFEST}',
            env=env)
        tree = git('write-tree', env=env)
    finally:
        os.unlink(index)
    return git('commit-tree', tree, '-p', man['parent'], '-m', SUBJECT.format(feature=man['feature'], seq=man['seq']))


def skills_source():
    if IN_CLOUD:
        return 'carried over'
    r = subprocess.run(['git', '-C', str(SKILL_DIR), 'rev-parse', 'HEAD'], capture_output=True, text=True)
    dirty = subprocess.run(['git', '-C', str(SKILL_DIR), 'status', '--porcelain', '--', '.'],
                           capture_output=True, text=True).stdout.strip()
    return (r.stdout.strip() + (' (dirty)' if dirty else '')) if r.returncode == 0 else 'unknown'


def cmd_push(a):
    repo = root()
    os.chdir(repo)
    seqs = remote_seqs(a.remote, a.feature)
    seq = max(seqs) + 1 if seqs else 0
    if seqs:
        base = manifest_at(fetch_seq(a.remote, a.feature, max(seqs)))['base']
        if a.base and git('rev-parse', a.base) != base:
            die(f'--base differs from the recorded base {base[:12]}')
    else:
        base = git('rev-parse', a.base or 'HEAD')
    head = git('rev-parse', 'HEAD')
    if not ok('merge-base', '--is-ancestor', base, head):
        die(f'base {base[:12]} is not an ancestor of HEAD')
    if not IN_CLOUD:
        # The base must already be on the remote; commits above it travel with the handoff commit.
        git('fetch', '-q', a.remote)
        if not git('branch', '-r', '--contains', base):
            msg = f'base {base[:12]} is on no {a.remote} branch; push the stack first'
            if not a.dry_run:
                die(msg)
            print(f'warning: {msg}', file=sys.stderr)
    if a.status == 'slice-ready' and not (repo / 'slice-report.md').is_file():
        die('slice-ready needs slice-report.md')
    man = {
        'feature': a.feature, 'seq': seq, 'base': base, 'parent': head,
        'status': a.status, 'slice': a.slice, 'skills_source': skills_source(),
        'created': datetime.datetime.now(datetime.timezone.utc).isoformat(timespec='seconds'),
    }
    commit = build_commit(repo, man, tuple(a.skills.split(',')) if a.skills else DEFAULT_SKILLS)
    branch = f'cloud/{a.feature}/{seq}'
    if a.dry_run:
        print(git('show', '--stat', '--format=%H %s', commit))
        return
    git('push', a.remote, f'{commit}:refs/heads/{branch}')
    if remote_seqs(a.remote, a.feature).get(seq) != commit:
        die(f'{a.remote}/{branch} does not point at {commit}')
    print(json.dumps({'branch': branch, 'commit': commit, 'parent': head, 'base': base, 'status': a.status}))


# ----- launch -----

PROMPT = """/implement
Cloud handoff session for feature `{feature}`, started from `{branch}`.
Before anything else, run `python3 .claude/skills/cloud-handoff/scripts/handoff.py resume --expect {commit}`. If it fails, report the error and stop.
Then implement the next slice unattended, following `.claude/rules/cloud-handoff.md`. It overrides `implement` where they differ.
{extra}"""


def cmd_launch(a):
    repo = root()
    if not a.dry_run and not (sys.stdin.isatty() and sys.stdout.isatty()):
        die('`claude --cloud` needs an interactive terminal; run launch in one (an agent: the Terminal panel)')
    seq, commit = pick_seq(a)
    branch = f'cloud/{a.feature}/{seq}'
    man, _ = check_handoff(commit)
    if man.get('status') not in (None, 'handoff'):
        die(f'{branch} has status {man.get("status")}; push the reviewed work first (push --feature {a.feature})')
    prompt = PROMPT.format(feature=a.feature, branch=branch, commit=commit, extra=a.message or '')
    # `claude --cloud` clones the remote at the current branch, so run it from a checkout of exactly that branch.
    wt = Path(tempfile.mkdtemp(prefix='handoff-launch-'))
    try:
        git('worktree', 'add', '-q', '--detach', str(wt), commit, cwd=repo)
        git('switch', '-q', '-C', branch, '--track', f'{a.remote}/{branch}', cwd=wt)
        cmd = ['claude', '--cloud', prompt] + (['--environment', a.environment] if a.environment else [])
        if a.dry_run:
            print(' '.join(cmd[:2]), repr(prompt), f'(in {wt})')
        else:
            subprocess.run(cmd, cwd=wt, check=True)
    finally:
        git('worktree', 'remove', '--force', str(wt), cwd=repo, check=False)
        git('branch', '-D', branch, cwd=repo, check=False)


# ----- watch -----

def cmd_watch(a):
    """Exit with one line when cloud/F/<seq> with seq > --after exists, or at the deadline."""
    end = time.monotonic() + a.deadline
    while True:
        seqs = {s: c for s, c in remote_seqs(a.remote, a.feature).items() if s > a.after}
        if seqs:
            seq = max(seqs)
            man = manifest_at(fetch_seq(a.remote, a.feature, seq))
            print(json.dumps({'event': 'push', 'seq': seq, 'status': man.get('status'), 'slice': man.get('slice')}),
                  flush=True)
            return
        if time.monotonic() >= end:
            print(json.dumps({'event': 'deadline', 'after': a.after}), flush=True)
            return
        time.sleep(a.interval)


# ----- resume (cloud) -----

def cmd_resume(a):
    repo = root()
    os.chdir(repo)
    head = git('rev-parse', 'HEAD')
    if not head.startswith(a.expect):
        die(f'session started at {head[:12]}, expected {a.expect[:12]}: wrong branch, do not continue')
    man, parent = check_handoff(head)
    # The handoff's changes become uncommitted again, as they were before the push.
    git('reset', '-q', '--soft', parent)
    git('reset', '-q')
    exclude = Path(git('rev-parse', '--git-path', 'info/exclude'))
    exclude.parent.mkdir(parents=True, exist_ok=True)
    lines = exclude.read_text().splitlines() if exclude.exists() else []
    lines += [f'/{p}' for p in (*SCRATCH, 'CLAUDE.md') if f'/{p}' not in lines]
    exclude.write_text('\n'.join(lines) + '\n')
    if (repo / '.githooks').is_dir():
        git('config', 'core.hooksPath', '.githooks')
    print(json.dumps({**man, 'resumed_at': parent}, indent=2))
    print('uncommitted after resume:\n' + (git('status', '--short') or '(none)'))


# ----- land -----

def discard_pushed_state(repo, head, remote, feature):
    """Drop uncommitted changes if they are exactly what an earlier handoff on HEAD carried."""
    if not git('status', '--porcelain'):
        return
    index, env = worktree_tree(repo)
    try:
        current = git('write-tree', env=env)
    finally:
        os.unlink(index)
    for seq, sha in sorted(remote_seqs(remote, feature).items(), reverse=True):
        commit = fetch_seq(remote, feature, seq)
        if git('rev-parse', f'{commit}^') != head:
            continue
        if ok('diff', '--quiet', current, commit, '--', '.', *TOOLING_EXCLUDES):
            added = git('diff', '--name-only', '--diff-filter=A', head, commit, '--', '.',
                        *TOOLING_EXCLUDES).splitlines()
            git('reset', '-q', '--hard', head)
            for path in added:
                (repo / path).unlink(missing_ok=True)
            return
    die('the worktree has uncommitted changes that no handoff on HEAD carried; commit or set them aside first')


def cmd_land(a):
    repo = root()
    os.chdir(repo)
    seq, commit = pick_seq(a)
    branch = f'cloud/{a.feature}/{seq}'
    man, parent = check_handoff(commit)
    head = git('rev-parse', 'HEAD')
    if not ok('merge-base', '--is-ancestor', head, parent):
        die(f'HEAD {head[:12]} is not an ancestor of the cloud work {parent[:12]}; '
            f'rebase the cloud work onto your branch first (base {man["base"][:12]})')
    discard_pushed_state(repo, head, a.remote, a.feature)
    for path in git('diff', '--name-only', '--diff-filter=A', parent, commit, '--', '.',
                    *TOOLING_EXCLUDES).splitlines():
        if (repo / path).exists():
            die(f'{path} exists locally but is not tracked; move it aside first')
    git('merge', '-q', '--ff-only', parent)
    diff = subprocess.run(['git', 'diff', '--binary', parent, commit, '--', '.', *TOOLING_EXCLUDES],
                          capture_output=True, check=True).stdout
    r = subprocess.run(['git', 'apply', '-'], input=diff, capture_output=True)
    if r.returncode != 0:
        die('could not restore the uncommitted changes: ' + r.stderr.decode())
    commits = git('log', '--reverse', '--format=%H %s', f'{head}..{parent}').splitlines()
    print(json.dumps({
        'landed': branch, 'status': man.get('status'), 'slice': man.get('slice'),
        'range': f'{head}..{parent}', 'commits': commits,
        'report': 'slice-report.md' if (repo / 'slice-report.md').is_file() else None,
    }, indent=2))


# ----- clean -----

def cmd_clean(a):
    seqs = remote_seqs(a.remote, a.feature)
    refs = [f':refs/heads/cloud/{a.feature}/{s}' for s in sorted(seqs)]
    if not refs:
        print('nothing to delete')
        return
    if a.dry_run:
        print('would delete:', ' '.join(r[1:] for r in refs))
        return
    git('push', a.remote, *refs)
    print(f'deleted {len(refs)} branches')


def main():
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = p.add_subparsers(dest='cmd', required=True)
    for name in ('push', 'launch', 'watch', 'land', 'clean'):
        s = sub.add_parser(name)
        s.add_argument('--feature', required=True, help='short slug; branches are cloud/<feature>/<seq>')
        s.add_argument('--remote', default='origin')
        s.add_argument('--dry-run', action='store_true')
        if name == 'push':
            s.add_argument('--base', help='base revision (first push only; default HEAD)')
            s.add_argument('--status', choices=STATUSES, default='handoff')
            s.add_argument('--slice', type=int)
            s.add_argument('--skills', help='comma-separated personal skills to carry')
        if name in ('launch', 'land'):
            s.add_argument('--seq', type=int, help='default: the latest')
        if name == 'launch':
            s.add_argument('--environment', help='cloud environment id')
            s.add_argument('--message', help='extra instructions appended to the prompt')
        if name == 'watch':
            s.add_argument('--after', type=int, required=True, help='the last seq already seen')
            s.add_argument('--deadline', type=int, default=25 * 60, help='seconds')
            s.add_argument('--interval', type=int, default=60, help='seconds')
    s = sub.add_parser('resume')
    s.add_argument('--expect', required=True, help='the handoff commit the session must start at')
    a = p.parse_args()
    {'push': cmd_push, 'launch': cmd_launch, 'watch': cmd_watch, 'resume': cmd_resume,
     'land': cmd_land, 'clean': cmd_clean}[a.cmd](a)


if __name__ == '__main__':
    main()
