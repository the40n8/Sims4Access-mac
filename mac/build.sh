#!/bin/zsh
# Build the macOS mod into dist/TS4_Access:
#   TS4_Access.ts4script   compiled from Scripts/ with Python 3.7
#   ts4access-say          speech companion (universal: Apple Silicon + Intel)
#   TS4_Access_UI.package  from the pinned upstream release (checksum verified)
# Env: PY37 = a Python 3.7 interpreter (default: micromamba env below).
set -euo pipefail

ROOT="${0:A:h:h}"
MAC="$ROOT/mac"
DIST="$ROOT/dist/TS4_Access"
CACHE="$MAC/.cache"
PY37="${PY37:-$HOME/.local/micromamba/envs/py37/bin/python}"

[[ -x "$PY37" ]] || { echo "Python 3.7 not found at $PY37 (set PY37=...). See .github/README.md"; exit 1; }
command -v swiftc >/dev/null || { echo "swiftc not found: xcode-select --install"; exit 1; }
mkdir -p "$DIST" "$CACHE"

echo "== Python mod"
"$PY37" -I "$MAC/build.py" "$ROOT" "$DIST"

echo "== Speech companion"
for arch in arm64 x86_64; do
  swiftc -O -target "$arch-apple-macos13" -o "$CACHE/ts4access-say-$arch" "$MAC/companion/TS4AccessSay.swift"
done
lipo -create -output "$DIST/ts4access-say" "$CACHE/ts4access-say-arm64" "$CACHE/ts4access-say-x86_64"
chmod +x "$DIST/ts4access-say"

echo "== UI package (upstream release)"
read -r REPO RELEASE ASSET SHA < <(python3 -I -c '
import json, sys; p = json.load(open(sys.argv[1]))
print(p["repo"], p["release"], p["asset"], p["sha256"])' "$MAC/ui-package.json")
ZIP="$CACHE/$RELEASE-$ASSET"
if [[ ! -f "$ZIP" ]]; then
  curl -fL --progress-bar -o "$ZIP.part" "https://github.com/$REPO/releases/download/$RELEASE/$ASSET"
  mv "$ZIP.part" "$ZIP"
fi
GOT=$(shasum -a 256 "$ZIP" | cut -d' ' -f1)
[[ "$GOT" == "$SHA" ]] || { rm -f "$ZIP"; echo "Checksum mismatch for $ASSET ($GOT). Not used."; exit 1; }
unzip -o -q -j "$ZIP" "Mods/TS4_Access/TS4_Access_UI.package" \
  "Mods/TS4_Access/README.txt" "Mods/TS4_Access/CONTROLS.txt" "Mods/TS4_Access/CHEATS.txt" -d "$DIST"

cp "$ROOT/LICENSE" "$DIST/LICENSE.txt"
cp "$MAC/PLAYERS.md" "$DIST/README-MAC.md"
echo "== Done: $DIST"
ls -l "$DIST"
