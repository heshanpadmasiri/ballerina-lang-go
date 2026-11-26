#!/usr/bin/env python3
"""Move uncommitted work between a local checkout and a Claude Code cloud session.

The only channel is a branch on the repository's remote: cloud/<feature>/<seq>.
Each branch tip is one handoff commit on top of the real work. It carries:

  - every uncommitted change in the worktree (spec.md, new tests, the pending block, ...),
  - the agent tooling the cloud does not have: skills, personal rules, the CLAUDE.md link,
  - .claude/handoff.json: feature, base, parent and seq.

Branches are never rewritten; every push creates the next seq.

  handoff.py push   --feature F   build the handoff commit from the worktree and push it
  handoff.py launch --feature F   start a cloud session on the latest cloud/F/<seq>
  handoff.py resume --expect SHA  (cloud) check the start commit, then restore the worktree
  handoff.py land   --feature F   fast-forward the local branch and restore the worktree

The worktree and index are never modified by push: the commit is built in a temporary index.
"""

import argparse
import datetime
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

SUBJECT = 'chore(handoff): {feature} #{seq} (drop before PR)'
SUBJECT_RE = re.compile(r'^chore\(handoff\): (\S+) #(\d+) \(drop before PR\)$')
MANIFEST = '.claude/handoff.json'
TOOLING = ('.claude/', 'CLAUDE.md')
# Design documents never committed; hidden from `git status` in the cloud so they are not staged.
SCRATCH = ('spec.md', 'research.md', 'progress.md', 'rca.md', 'CLAUDE.md')
DEFAULT_SKILLS = (
    'cloud-handoff', 'implement', 'design', 'review-loop', 'review-oracle', 'commit',
    'fix-up-target', 'test-oracle', 'show-me', 'rebase', 'resolve-conflicts',
    'nutcracker-behavior-policy', 'nutcracker-review-guidelines', 'nutcracker-test-oracle',
)
HOME_SKILLS = Path.home() / '.claude' / 'skills'
HOME_RULES = Path.home() / '.claude' / 'CLAUDE.md'
SKILL_DIR = Path(__file__).resolve().parent.parent


def die(msg):
    print(f'error: {msg}', file=sys.stderr)
    sys.exit(1)


def git(*args, env=None, check=True, cwd=None):
    r = subprocess.run(['git', *args], capture_output=True, text=True, env=env, cwd=cwd)
    if check and r.returncode != 0:
        die(f'git {" ".join(args)}: {r.stderr.strip() or r.stdout.strip()}')
    return r.stdout.strip()


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


def manifest_at(commit):
    r = subprocess.run(['git', 'show', f'{commit}:{MANIFEST}'], capture_output=True, text=True)
    if r.returncode != 0:
        die(f'{commit} has no {MANIFEST}; it is not a handoff commit')
    return json.loads(r.stdout)


def check_handoff(commit):
    subject = git('log', '-1', '--format=%s', commit)
    m = SUBJECT_RE.match(subject)
    if not m:
        die(f'{commit[:12]} is not a handoff commit (subject: {subject!r})')
    man = manifest_at(commit)
    parent = git('rev-parse', f'{commit}^')
    if man['parent'] != parent:
        die(f'handoff {commit[:12]} records parent {man["parent"][:12]} but sits on {parent[:12]}')
    if subprocess.run(['git', 'merge-base', '--is-ancestor', man['base'], parent]).returncode != 0:
        die(f'base {man["base"][:12]} is not an ancestor of {parent[:12]}')
    return man, parent


# ----- push -----

def tooling_files(repo, skills):
    """(path in repo, source file or symlink target, mode) for the injected tooling."""
    files = []
    on_disk = (repo / MANIFEST).exists()  # a resumed cloud worktree already has the tooling
    if on_disk:
        for p in sorted((repo / '.claude').rglob('*')):
            rel = p.relative_to(repo).as_posix()
            if p.is_file() and not p.is_symlink() and rel != MANIFEST:
                files.append((rel, p, '100644' if not os.access(p, os.X_OK) else '100755'))
    else:
        for name in skills:
            src = (HOME_SKILLS / name).resolve() if name != 'cloud-handoff' else SKILL_DIR
            if not (src / 'SKILL.md').exists():
                die(f'skill {name} not found in {HOME_SKILLS}')
            add_tree(files, src, f'.claude/skills/{name}')
        repo_skills = repo / '.agents' / 'skills'
        if repo_skills.is_dir():
            for d in sorted(repo_skills.iterdir()):
                if (d / 'SKILL.md').exists() and d.name not in skills:
                    add_tree(files, d, f'.claude/skills/{d.name}')
        if HOME_RULES.exists():
            text = HOME_RULES.read_text()
            # The cloud runs Linux.
            text = text.replace('$(sysctl -n hw.ncpu)', '$(nproc)')
            tmp = Path(tempfile.mkstemp(suffix='.md')[1])
            tmp.write_text(text)
            files.append(('.claude/rules/personal.md', tmp, '100644'))
        files.append(('.claude/rules/cloud-handoff.md', SKILL_DIR / 'references' / 'cloud-rules.md', '100644'))
    if (repo / 'AGENTS.md').exists() and not tracked(repo, 'CLAUDE.md'):
        files.append(('CLAUDE.md', 'AGENTS.md', '120000'))
    return files


def add_tree(files, src, dest):
    for p in sorted(src.rglob('*')):
        if p.is_dir() or '__pycache__' in p.parts or p.name == '.DS_Store':
            continue
        real = p.resolve()
        if real.is_file():
            mode = '100755' if os.access(real, os.X_OK) else '100644'
            files.append((f'{dest}/{p.relative_to(src).as_posix()}', real, mode))


def tracked(repo, path):
    return subprocess.run(['git', 'ls-files', '--error-unmatch', path], cwd=repo,
                          capture_output=True).returncode == 0


def build_commit(repo, feature, seq, base, skills):
    head = git('rev-parse', 'HEAD')
    fd, index = tempfile.mkstemp(prefix='handoff-index-')
    os.close(fd)
    env = dict(os.environ, GIT_INDEX_FILE=index)
    try:
        git('read-tree', 'HEAD', env=env)
        git('add', '-A', '--', '.', env=env)  # every uncommitted change, ignored files excluded
        # resume hides the design documents from git; carry them anyway.
        scratch = [p for p in SCRATCH if p != 'CLAUDE.md' and (repo / p).is_file()]
        if scratch:
            git('add', '-f', '--', *scratch, env=env)
        for rel, src, mode in tooling_files(repo, skills):
            if mode == '120000':
                blob = subprocess.run(['git', 'hash-object', '-w', '--stdin'], input=str(src),
                                      capture_output=True, text=True, check=True).stdout.strip()
            else:
                blob = git('hash-object', '-w', '--', str(src))
            git('update-index', '--add', '--cacheinfo', f'{mode},{blob},{rel}', env=env)
        man = {
            'feature': feature, 'seq': seq, 'base': base, 'parent': head,
            'skills_source': skills_source(),
            'created': datetime.datetime.now(datetime.timezone.utc).isoformat(timespec='seconds'),
        }
        blob = subprocess.run(['git', 'hash-object', '-w', '--stdin'], input=json.dumps(man, indent=2) + '\n',
                              capture_output=True, text=True, check=True).stdout.strip()
        git('update-index', '--add', '--cacheinfo', f'100644,{blob},{MANIFEST}', env=env)
        tree = git('write-tree', env=env)
    finally:
        os.unlink(index)
    return git('commit-tree', tree, '-p', head, '-m', SUBJECT.format(feature=feature, seq=seq)), man


def skills_source():
    if os.environ.get('CLAUDE_CODE_REMOTE'):
        return 'carried over'
    r = subprocess.run(['git', '-C', str(SKILL_DIR), 'rev-parse', 'HEAD'], capture_output=True, text=True)
    dirty = subprocess.run(['git', '-C', str(SKILL_DIR), 'status', '--porcelain', '--', '.'],
                           capture_output=True, text=True).stdout.strip()
    return (r.stdout.strip() + (' (dirty)' if dirty else '')) if r.returncode == 0 else 'unknown'


def cmd_push(a):
    repo = root()
    os.chdir(repo)
    if git('rev-parse', '--abbrev-ref', 'HEAD') == 'HEAD' and not os.environ.get('CLAUDE_CODE_REMOTE'):
        print('note: HEAD is detached', file=sys.stderr)
    seqs = remote_seqs(a.remote, a.feature)
    seq = max(seqs) + 1 if seqs else 0
    local_man = repo / MANIFEST
    if local_man.exists():
        base = json.loads(local_man.read_text())['base']
    else:
        base = git('rev-parse', a.base or 'HEAD')
    head = git('rev-parse', 'HEAD')
    if subprocess.run(['git', 'merge-base', '--is-ancestor', base, head]).returncode != 0:
        die(f'base {base[:12]} is not an ancestor of HEAD')
    if not os.environ.get('CLAUDE_CODE_REMOTE'):
        # The cloud clones from the remote: the parent must already be there.
        git('fetch', '-q', a.remote)
        if not git('branch', '-r', '--contains', head):
            msg = f'HEAD {head[:12]} is on no {a.remote} branch; push the stack first'
            if not a.dry_run:
                die(msg)
            print(f'warning: {msg}', file=sys.stderr)
    skills = tuple(a.skills.split(',')) if a.skills else DEFAULT_SKILLS
    commit, man = build_commit(repo, a.feature, seq, base, skills)
    branch = f'cloud/{a.feature}/{seq}'
    if a.dry_run:
        print(git('show', '--stat', '--format=%H %s', commit))
        return
    git('push', a.remote, f'{commit}:refs/heads/{branch}')
    remote_sha = remote_seqs(a.remote, a.feature).get(seq)
    if remote_sha != commit:
        die(f'{a.remote}/{branch} is {remote_sha}, expected {commit}')
    print(json.dumps({'branch': branch, 'commit': commit, 'parent': head, 'base': base}))


# ----- launch -----

PROMPT = """/implement
This is a cloud handoff session for feature `{feature}` (branch `{branch}`).
Before anything else, run `python3 .claude/skills/cloud-handoff/scripts/handoff.py resume --expect {commit}`. If it fails, report the error and stop.
Follow `.claude/rules/cloud-handoff.md` for pushing: never push except through that script.
{extra}"""


def cmd_launch(a):
    repo = root()
    seqs = remote_seqs(a.remote, a.feature)
    if not seqs:
        die(f'no cloud/{a.feature}/* branch on {a.remote}')
    seq = a.seq if a.seq is not None else max(seqs)
    if seq not in seqs:
        die(f'cloud/{a.feature}/{seq} does not exist')
    branch, commit = f'cloud/{a.feature}/{seq}', seqs[seq]
    git('fetch', '-q', a.remote, f'refs/heads/{branch}:refs/remotes/{a.remote}/{branch}')
    check_handoff(commit)
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


# ----- resume (cloud) -----

def cmd_resume(a):
    repo = root()
    os.chdir(repo)
    head = git('rev-parse', 'HEAD')
    if not head.startswith(a.expect):
        die(f'session started at {head[:12]}, expected {a.expect[:12]}: wrong branch, do not continue')
    man, parent = check_handoff(head)
    # Restore the worktree as it was before the push: the handoff's changes become uncommitted again.
    git('reset', '-q', '--soft', parent)
    git('reset', '-q')
    exclude = Path(git('rev-parse', '--git-path', 'info/exclude'))
    exclude.parent.mkdir(parents=True, exist_ok=True)
    lines = exclude.read_text().splitlines() if exclude.exists() else []
    exclude.write_text('\n'.join(lines + [f'/{p}' for p in SCRATCH if f'/{p}' not in lines]) + '\n')
    if (repo / '.githooks').is_dir():
        git('config', 'core.hooksPath', '.githooks')
    status = git('status', '--short')
    print(json.dumps({**man, 'resumed_at': parent, 'next_push': 'handoff.py push --feature ' + man['feature']},
                     indent=2))
    print('uncommitted after resume:\n' + (status or '(none)'))


# ----- land -----

def cmd_land(a):
    repo = root()
    os.chdir(repo)
    seqs = remote_seqs(a.remote, a.feature)
    if not seqs:
        die(f'no cloud/{a.feature}/* branch on {a.remote}')
    seq = a.seq if a.seq is not None else max(seqs)
    branch, commit = f'cloud/{a.feature}/{seq}', seqs[seq]
    git('fetch', '-q', a.remote, f'refs/heads/{branch}')
    man, parent = check_handoff(commit)
    if git('status', '--porcelain', '--untracked-files=no'):
        die('the worktree has uncommitted changes to tracked files; commit or stash them first')
    head = git('rev-parse', 'HEAD')
    if subprocess.run(['git', 'merge-base', '--is-ancestor', head, parent]).returncode != 0:
        die(f'HEAD {head[:12]} is not an ancestor of the cloud work {parent[:12]}; '
            f'rebase cloud work onto your branch first (base was {man["base"][:12]})')
    diff = subprocess.run(['git', 'diff', '--binary', parent, commit, '--', '.',
                           *[f':(exclude){p}' for p in TOOLING]], capture_output=True, check=True).stdout
    # Untracked files the handoff adds back, such as spec.md: keep identical ones, back up the rest.
    added = git('diff', '--name-only', '--diff-filter=A', parent, commit, '--', '.',
                *[f':(exclude){p}' for p in TOOLING]).splitlines()
    git('merge', '-q', '--ff-only', parent)
    skipped = []
    for path in added:
        p = repo / path
        if p.exists():
            blob = git('rev-parse', f'{commit}:{path}')
            if git('hash-object', str(p)) == blob:
                skipped.append(path)
                continue
            shutil.move(p, p.with_name(p.name + '.orig'))
            print(f'{path}: local copy differs, moved to {path}.orig')
    r = subprocess.run(['git', 'apply', *[f'--exclude={s}' for s in skipped], '-'], input=diff,
                       capture_output=True)
    if r.returncode != 0:
        die('could not restore the uncommitted changes: ' + r.stderr.decode())
    print(f'landed {branch}: HEAD is now {parent[:12]}; uncommitted changes restored')
    print(git('status', '--short'))


def main():
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = p.add_subparsers(dest='cmd', required=True)
    for name in ('push', 'launch', 'land'):
        s = sub.add_parser(name)
        s.add_argument('--feature', required=True, help='short slug; branches are cloud/<feature>/<seq>')
        s.add_argument('--remote', default='origin')
        s.add_argument('--dry-run', action='store_true')
        if name == 'push':
            s.add_argument('--base', help='base revision (first push only; default HEAD)')
            s.add_argument('--skills', help='comma-separated personal skills to carry')
        else:
            s.add_argument('--seq', type=int, help='default: the latest')
        if name == 'launch':
            s.add_argument('--environment', help='cloud environment id')
            s.add_argument('--message', help='extra instructions appended to the prompt')
    s = sub.add_parser('resume')
    s.add_argument('--expect', required=True, help='the handoff commit the session must start at')
    a = p.parse_args()
    {'push': cmd_push, 'launch': cmd_launch, 'resume': cmd_resume, 'land': cmd_land}[a.cmd](a)


if __name__ == '__main__':
    main()
