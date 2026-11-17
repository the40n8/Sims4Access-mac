#!/bin/zsh
# Build the macOS mod into dist/TS4_Access:
#   TS4_Access.ts4script   compiled from Scripts/ with Python 3.7
#   ts4access-bridge       phone companion: web app + API (Go, universal)
#   ts4access-say          speech companion (Swift, universal)
#   TS4_Access_UI.package  from the pinned upstream release (checksum verified)
#   cloudflared            Cloudflare's tunnel client, pinned and signature-checked
# Env: PY37 = a Python 3.7 interpreter (default: micromamba env below).
# Needs: swiftc (Xcode Command Line Tools), go, pnpm.
set -euo pipefail

ROOT="${0:A:h:h}"
MAC="$ROOT/mac"
COMP="$ROOT/companion"
DIST="$ROOT/dist/TS4_Access"
CACHE="$MAC/.cache"
PY37="${PY37:-$HOME/.local/micromamba/envs/py37/bin/python}"

[[ -x "$PY37" ]] || { echo "Python 3.7 not found at $PY37 (set PY37=...). See .github/README.md"; exit 1; }
for tool in swiftc go pnpm; do
  command -v $tool >/dev/null || { echo "$tool not found. See .github/README.md"; exit 1; }
done
mkdir -p "$DIST" "$CACHE"

echo "== Python mod"
"$PY37" -I "$MAC/build.py" "$ROOT" "$DIST"

echo "== Speech companion"
for arch in arm64 x86_64; do
  swiftc -O -target "$arch-apple-macos13" -o "$CACHE/ts4access-say-$arch" "$COMP/speaker/TS4AccessSay.swift"
done
lipo -create -output "$DIST/ts4access-say" "$CACHE/ts4access-say-arm64" "$CACHE/ts4access-say-x86_64"

echo "== Phone app"
(cd "$COMP/web" && pnpm install --frozen-lockfile --silent && pnpm build >/dev/null)
rm -rf "$COMP/bridge/internal/webui/dist"
cp -R "$COMP/web/build" "$COMP/bridge/internal/webui/dist"

echo "== Bridge"
for arch in arm64 amd64; do
  (cd "$COMP/bridge" && CGO_ENABLED=0 GOOS=darwin GOARCH=$arch \
    go build -trimpath -ldflags "-s -w" -o "$CACHE/ts4access-bridge-$arch" ./cmd/ts4access-bridge)
done
lipo -create -output "$DIST/ts4access-bridge" "$CACHE/ts4access-bridge-arm64" "$CACHE/ts4access-bridge-amd64"
chmod +x "$DIST/ts4access-say" "$DIST/ts4access-bridge"

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

echo "== cloudflared (Cloudflare quick tunnel)"
read -r CF_VER CF_TEAM CF_ARM CF_AMD < <(python3 -I -c '
import json, sys; p = json.load(open(sys.argv[1]))
print(p["version"], p["team_id"], p["sha256"]["arm64"], p["sha256"]["amd64"])' "$MAC/cloudflared.json")
for arch in arm64 amd64; do
  TGZ="$CACHE/cloudflared-$CF_VER-$arch.tgz"
  [[ -f "$TGZ" ]] || curl -fsSL -o "$TGZ" "https://github.com/cloudflare/cloudflared/releases/download/$CF_VER/cloudflared-darwin-$arch.tgz"
  rm -rf "$CACHE/cf-$arch" && mkdir -p "$CACHE/cf-$arch" && tar -xzf "$TGZ" -C "$CACHE/cf-$arch"
  WANT=$([[ $arch == arm64 ]] && echo $CF_ARM || echo $CF_AMD)
  GOT=$(shasum -a 256 "$CACHE/cf-$arch/cloudflared" | cut -d' ' -f1)
  [[ "$GOT" == "$WANT" ]] || { echo "cloudflared $arch checksum mismatch ($GOT)"; exit 1; }
done
lipo -create -output "$DIST/cloudflared" "$CACHE/cf-arm64/cloudflared" "$CACHE/cf-amd64/cloudflared"
codesign --verify --strict --all-architectures "$DIST/cloudflared"
codesign -dv "$DIST/cloudflared" 2>&1 | grep -q "TeamIdentifier=$CF_TEAM" || { echo "cloudflared is not signed by Cloudflare"; exit 1; }
[[ -f "$CACHE/cloudflared-LICENSE-$CF_VER" ]] || curl -fsSL -o "$CACHE/cloudflared-LICENSE-$CF_VER" "https://raw.githubusercontent.com/cloudflare/cloudflared/$CF_VER/LICENSE"
cp "$CACHE/cloudflared-LICENSE-$CF_VER" "$DIST/cloudflared-LICENSE.txt"

cp "$ROOT/LICENSE" "$DIST/LICENSE.txt"
cp "$MAC/PLAYERS.md" "$DIST/README-MAC.md"
echo "== Done: $DIST"
ls -l "$DIST"
