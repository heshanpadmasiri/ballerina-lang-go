#!/usr/bin/env python3
"""Check Mermaid diagrams by rendering them with mermaid-cli (mmdc).

  check_mermaid.py FILE...    check each .mmd file, or every ```mermaid block in a Markdown file
  check_mermaid.py -          read Markdown or Mermaid from stdin

Prints `<file>:<line>: mermaid: <error>` for each diagram that does not render, with
<line> pointing at the failing line in FILE. Exit status: 0 all render, 1 a diagram
failed, 2 the checker cannot run (mmdc or a browser is missing, or mmdc timed out). Status 2 never means
the diagrams are valid.

Rendering, not just parsing, is what GitHub and Markdown viewers do, so a diagram that
passes here renders there. Requires `brew install mermaid-cli`; it reuses an installed
Chrome when Puppeteer has not downloaded its own.
"""

import os
import re
import shutil
import subprocess
import sys
import tempfile
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

FENCE = re.compile(r'^(\s*)(`{3,}|~{3,})\s*(\S*)')
ERROR_LINE = re.compile(r'(?:Parse error|Lexical error) on line (\d+)')
TIMEOUT = 120


class Unavailable(Exception):
    pass


def browser_env():
    """Point Puppeteer at a usable browser when it has none of its own."""
    env = dict(os.environ)
    if env.get('PUPPETEER_EXECUTABLE_PATH'):
        return env
    if list(Path.home().glob('.cache/puppeteer/chrome-headless-shell/*')):
        return env
    candidates = [Path('/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'),
                  Path('/Applications/Chromium.app/Contents/MacOS/Chromium')]
    candidates += sorted(Path.home().glob('Library/Caches/ms-playwright/chromium_headless_shell-*/'
                                          'chrome-headless-shell-*/chrome-headless-shell'), reverse=True)
    candidates += [Path(p) for p in map(shutil.which, ('google-chrome', 'chromium', 'chromium-browser')) if p]
    for path in candidates:
        if path.is_file() and os.access(path, os.X_OK):
            env['PUPPETEER_EXECUTABLE_PATH'] = str(path)
            return env
    raise Unavailable('no browser for mmdc: install Chrome, or run '
                      '`npx puppeteer browsers install chrome-headless-shell`')


def diagrams(text, is_markdown):
    """(first line of the diagram body in the file, source) for each diagram."""
    if not is_markdown:
        return [(1, text)]
    out, fence, lang, start, body = [], None, '', 0, []
    for n, line in enumerate(text.splitlines(), 1):
        if fence:
            if line.strip() == fence:
                if lang == 'mermaid':
                    out.append((start, '\n'.join(body) + '\n'))
                fence = None
            else:
                body.append(line[indent:] if line[:indent].isspace() else line.lstrip())
            continue
        m = FENCE.match(line)
        if m:
            indent, fence, lang, start, body = len(m.group(1)), m.group(2), m.group(3).lower(), n + 1, []
    return out


def render(source, env):
    """None when the diagram renders, else mmdc's error message."""
    with tempfile.TemporaryDirectory(prefix='check-mermaid-') as tmp:
        src, out = Path(tmp, 'd.mmd'), Path(tmp, 'd.svg')
        src.write_text(source)
        try:
            result = subprocess.run(['mmdc', '-q', '-i', str(src), '-o', str(out)], env=env,
                                    capture_output=True, text=True, timeout=TIMEOUT)
        except subprocess.TimeoutExpired:
            raise Unavailable(f'mmdc timed out after {TIMEOUT}s; the browser may be slow to start, so retry')
        if result.returncode == 0 and out.is_file():
            return None
        return result.stderr or result.stdout or f'mmdc exited with status {result.returncode}'


def summarise(error):
    """Keep mmdc's message and drop the JavaScript stack."""
    kept = []
    for line in error.strip().splitlines():
        if line.startswith(('    at ', 'Parser.')) or 'mermaid-cli-intercept' in line:
            break
        kept.append(line)
    if kept and kept[0].strip().startswith('Error: Could not find'):
        raise Unavailable(kept[0].strip())
    return '\n    '.join(l for l in kept if l.strip()) or error.strip().splitlines()[0]


def check(name, text, env):
    """Problems as '<name>:<line>: mermaid: <message>'."""
    is_markdown = not name.endswith('.mmd') and ('```' in text or '~~~' in text)
    found = diagrams(text, is_markdown)
    with ThreadPoolExecutor(max_workers=4) as pool:
        errors = list(pool.map(lambda d: render(d[1], env), found))
    problems = []
    for (start, _), error in zip(found, errors):
        if error is None:
            continue
        message = summarise(error)
        m = ERROR_LINE.search(message)
        line = start + int(m.group(1)) - 1 if m else start
        problems.append(f'{name}:{line}: mermaid: {message}')
    return problems, len(found)


def check_texts(items):
    """Check [(name, text)]; returns (problems, diagram count). Raises Unavailable."""
    if not shutil.which('mmdc'):
        raise Unavailable('mmdc not found: run `brew install mermaid-cli`')
    env = browser_env()
    problems, count = [], 0
    for name, text in items:
        p, c = check(name, text, env)
        problems += p
        count += c
    return problems, count


def main(argv):
    if not argv or argv[0] in ('-h', '--help'):
        print(__doc__)
        return 0 if argv else 2
    items = [('<stdin>', sys.stdin.read()) if a == '-' else (a, Path(a).read_text()) for a in argv]
    try:
        problems, count = check_texts(items)
    except Unavailable as e:
        print(f'cannot check Mermaid: {e}', file=sys.stderr)
        return 2
    for p in problems:
        print(p, file=sys.stderr)
    print(f'{count} diagrams, {len(problems)} failed', file=sys.stderr)
    return 1 if problems else 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
