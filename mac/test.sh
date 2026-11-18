#!/bin/zsh
# Offline checks of the Mac-only mod code with the game's Python 3.7 and
# stand-ins for the game (mac/tests), plus the bridge's Go tests.
# They cover our logic, not EA's APIs: those still need a game session.
set -euo pipefail
ROOT="${0:A:h:h}"
PY37="${PY37:-$HOME/.local/micromamba/envs/py37/bin/python}"
echo "== Mod (Python 3.7, fake game)"
for t in "$ROOT"/mac/tests/test_*.py; do
  (cd "$ROOT/mac/tests" && "$PY37" -E -s -B "$t" 2>&1 | tail -1 | sed "s|^|${t:t}: |")
done
echo "== Bridge (Go)"
(cd "$ROOT/companion/bridge" && go test ./... 2>&1 | grep -v "no test files")
