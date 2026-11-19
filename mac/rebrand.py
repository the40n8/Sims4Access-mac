"""Apply the SimSight names to the source tree.

SimSight is based on TS4 Access (tyjokr, LeonianUniverse and contributors).
Its code keeps coming from there (mac/sync-upstream.sh), with TS4 Access
names; this script renames them the same way every time, so merges stay
mechanical:

- Python modules tyjokr_a11y_*.py -> simsight_*.py, and every reference.
- Console commands a11y.* -> simsight.*; each command is also still
  registered under its a11y.* name, because the prebuilt UI package (Flash,
  from the upstream release) calls those names.
- Files and folders: TS4_Access.ts4script, .cfg, .log... -> SimSight...;
  helpers ts4access-bridge / ts4access-say -> simsight-bridge / simsight-say;
  protobuf package ts4access.v1 -> simsight.v1.
- Words players hear and read: "TS4 Access" -> "SimSight", except where the
  original project is credited.

Idempotent: running it on renamed code changes nothing.
Usage: python3 mac/rebrand.py [repo root]
"""
import os
import re
import subprocess
import sys

ROOT = os.path.abspath(sys.argv[1] if len(sys.argv) > 1 else os.path.join(os.path.dirname(__file__), '..'))

# Never touched: the original's credits and licence, upstream's own release
# files and Windows-only list (named as upstream names them), generated
# code (regenerated from the renamed .proto files), and this script.
SKIP_DIRS = {'.git', 'node_modules', 'dist', 'build', '.cache', 'flash', 'a11y-reports',
             '__pycache__', '.svelte-kit', '.wrangler', 'gen'}
SKIP_FILES = {'LICENSE', 'LICENSE.txt', 'CONTRIBUTORS.md', 'README.txt', 'CONTROLS.txt', 'CHEATS.txt',
              'pnpm-lock.yaml', 'go.sum'}
SKIP_PATHS = {'mac/rebrand.py', 'mac/windows-paths.txt', 'mac/ui-package.json', 'mac/pin-ui.sh',
              'mac/build.sh', 'mac/sync-upstream.sh'}
TEXT = ('.py', '.md', '.txt', '.sh', '.go', '.swift', '.ts', '.js', '.svelte', '.json', '.jsonc',
        '.proto', '.yaml', '.yml', '.html', '.css', '.cfg', '.command', '.mod')

# Phrases that credit the original project keep its name.
KEEP = ['TS4 Access by **tyjokr', 'upstream TS4 Access', 'TS4 Access Helper', 'TS4 Access by tyjokr', '[TS4 Access]',
        'based on TS4 Access', "tyjokr's TS4 Access", 'original TS4 Access', 'TS4 Access test builds',
        'TS4AccessUpdater', 'LeonianUniverse478/Sims4Access', 'TS4_Access_Update.zip']

RULES = [
    # Commands keep their a11y.* alias (the prebuilt UI calls them).
    (re.compile(r"Command\((['\"])a11y\.([a-z0-9_]+)\1"), r"Command(\1simsight.\2\1, \1a11y@KEEP@.\2\1"),
    (re.compile(r'tyjokr_a11y_'), 'simsight_'),
    (re.compile(r'tyjokr_a11y'), 'simsight'),
    (re.compile(r'(?<![\w./-])a11y\.(?=[a-z])'), 'simsight.'),
    (re.compile(r"(['\"])a11y: "), r'\1simsight: '),
    (re.compile(r'TS4 Access Plus'), 'SimSight Plus'),
    (re.compile(r'TS4 Access'), 'SimSight'),
    (re.compile(r'TS4_Access'), 'SimSight'),
    (re.compile(r'TS4AccessSay'), 'SimSightSay'),
    (re.compile(r'TS4AccessCompanion'), 'SimSightCompanion'),
    (re.compile(r'TS4Access'), 'SimSight'),
    (re.compile(r'TS4ACCESS_'), 'SIMSIGHT_'),
    (re.compile(r'\bTS4A_'), 'SIMSIGHT_'),
    (re.compile(r'ts4accessv1'), 'simsightv1'),
    (re.compile(r'ts4access'), 'simsight'),
    (re.compile(r'\bts4k\b'), 'ssk'),
    (re.compile(r'\bts4g\b'), 'ssg'),
]


ALIAS = re.compile(r"(Command\((['\"])simsight\.[a-z0-9_]+\2, \2)a11y\.")


def rebrand_text(text):
    kept = {}
    for i, phrase in enumerate(KEEP):
        token = '\x01{}\x01'.format(i)
        if phrase in text:
            kept[token] = phrase
            text = text.replace(phrase, token)
    # Aliases already in place stay a11y.* (idempotent).
    text = ALIAS.sub(r'\1a11y@KEEP@.', text)
    for pattern, repl in RULES:
        text = pattern.sub(repl, text)
    for token, phrase in kept.items():
        text = text.replace(token, phrase)
    return text.replace('a11y@KEEP@.', 'a11y.')


def new_name(rel):
    parts = rel.split('/')
    out = []
    for p in parts:
        p = re.sub(r'^tyjokr_a11y_', 'simsight_', p)
        p = p.replace('TS4AccessSay', 'SimSightSay').replace('ts4accessv1', 'simsightv1')
        p = p.replace('ts4access', 'simsight')
        out.append(p)
    return '/'.join(out)


def tracked():
    out = subprocess.run(['git', 'ls-files', '-z'], cwd=ROOT, stdout=subprocess.PIPE, check=True).stdout
    return [p for p in out.decode().split('\0') if p]


def skipped(rel):
    parts = rel.split('/')
    return (any(p in SKIP_DIRS for p in parts[:-1]) or parts[-1] in SKIP_FILES or rel in SKIP_PATHS
            or not rel.endswith(TEXT))


def main():
    renamed = changed = 0
    for rel in tracked():
        if any(p in SKIP_DIRS for p in rel.split('/')[:-1]) or rel in SKIP_PATHS:
            continue
        dest = new_name(rel)
        if dest != rel and os.path.exists(os.path.join(ROOT, rel)):
            os.makedirs(os.path.dirname(os.path.join(ROOT, dest)) or ROOT, exist_ok=True)
            subprocess.run(['git', 'mv', rel, dest], cwd=ROOT, check=True)
            renamed += 1
    for rel in tracked():
        path = os.path.join(ROOT, rel)
        if skipped(rel) or not os.path.isfile(path):
            continue
        with open(path, encoding='utf-8', newline='') as f:
            try:
                text = f.read()
            except UnicodeDecodeError:
                continue
        new = rebrand_text(text)
        if new != text:
            with open(path, 'w', encoding='utf-8', newline='') as f:
                f.write(new)
            changed += 1
    print('rebrand: {} renamed, {} files updated'.format(renamed, changed))


if __name__ == '__main__':
    main()
