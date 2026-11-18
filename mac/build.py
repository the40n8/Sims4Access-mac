"""Compile Scripts/ into TS4_Access.ts4script for the macOS game.

Same output format as upstream compile.py (Python 3.7 .pyc, PEP 552 header,
stored zip), but it writes only to the given folder and never deploys.
Usage: <python3.7> mac/build.py <repo root> <out dir>
"""
import marshal
import os
import struct
import sys
import zipfile
from importlib.util import MAGIC_NUMBER

# Never shipped on macOS even if an upstream merge brings them back before
# sync-upstream.sh deletes them: the game imports every module at load, and
# these need ctypes, which the Mac game does not have.
SKIP_DIRS = {'ctypes', '__pycache__'}
SKIP_FILES = {'tyjokr_a11y_prism.py',
              # Windows screen/DPI measurement and the SAPI kernel job (ctypes at load);
              # only the Windows build helpers that move the real mouse import them.
              'tyjokr_a11y_build_coordinates.py', 'tyjokr_a11y_display_diagnostic.py',
              'tyjokr_a11y_speech_process.py'}


def main():
    if sys.version_info[:2] != (3, 7):
        sys.exit('mac/build.py must run under Python 3.7 (the game loads 3.7 bytecode only)')
    src = os.path.join(sys.argv[1], 'Scripts')
    out = sys.argv[2]
    os.makedirs(out, exist_ok=True)
    target = os.path.join(out, 'TS4_Access.ts4script')
    count = 0
    with zipfile.ZipFile(target, 'w', zipfile.ZIP_STORED) as zf:
        for folder, subs, files in os.walk(src):
            subs[:] = sorted(d for d in subs if d not in SKIP_DIRS)
            for name in sorted(files):
                if not name.endswith('.py') or name in SKIP_FILES:
                    continue
                full = os.path.join(folder, name)
                rel = os.path.relpath(full, src).replace(os.sep, '/')
                with open(full, encoding='utf-8') as handle:
                    code = compile(handle.read(), rel, 'exec', optimize=2)
                header = MAGIC_NUMBER + struct.pack('<III', 0, 0, 0)
                zf.writestr(rel[:-3] + '.pyc', header + marshal.dumps(code))
                count += 1
    print('built {} modules -> {}'.format(count, target))


if __name__ == '__main__':
    main()
