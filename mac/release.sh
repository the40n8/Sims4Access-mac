#!/bin/zsh
# Build a player release: dist/release/TS4_Access-mac.zip, its SHA256SUMS and
# install.sh (the one-command installer, mac/get.sh). Nothing is published
# unless --publish is given (then: a GitHub release on the40n8/Sims4Access-mac).
#
#   mac/release.sh             build and check the files only
#   mac/release.sh --publish   also create the GitHub release (public)
set -euo pipefail
ROOT="${0:A:h:h}"
cd "$ROOT"
PUBLISH=0
[[ "${1:-}" == --publish ]] && PUBLISH=1

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Uncommitted changes: commit them first, a release is built from a commit."; exit 1
fi
TAG="mac-$(date +%Y.%m.%d)"
n=2
while git rev-parse -q --verify "refs/tags/$TAG" >/dev/null || gh release view "$TAG" >/dev/null 2>&1; do
  TAG="mac-$(date +%Y.%m.%d)-$n"; n=$((n + 1))
done
GAME=$(python3 -c 'import json;print(json.load(open("mac/ui-package.json"))["game_version"])')
UPSTREAM=$(python3 -c 'import json;print(json.load(open("mac/ui-package.json"))["release"])')

./mac/build.sh
./mac/test.sh

OUT="$ROOT/dist/release"
PKG="$OUT/TS4_Access-mac"
rm -rf "$OUT" && mkdir -p "$PKG"
cp -Rp dist/TS4_Access "$PKG/TS4_Access"
cp -p mac/install.sh mac/uninstall.sh "$PKG/"
cat > "$PKG/Install TS4 Access.command" <<'CMD'
#!/bin/zsh
# Double-click to install or update TS4 Access (quit the game first).
cd "${0:A:h}" && ./install.sh
echo; echo "You can close this window."
CMD
cat > "$PKG/Uninstall TS4 Access.command" <<'CMD'
#!/bin/zsh
# Double-click to remove TS4 Access (saves are not touched).
cd "${0:A:h}" && ./uninstall.sh
echo; echo "TS4 Access removed. You can close this window."
CMD
chmod +x "$PKG"/*.command "$PKG"/*.sh
# The optional meeting point, with the phone app it serves.
mkdir -p "$PKG/meeting-point"
cp -Rp companion/meeting-point/src "$PKG/meeting-point/src"
cp -Rp companion/web/build "$PKG/meeting-point/app"
sed 's#"directory": "../web/build"#"directory": "./app"#' companion/meeting-point/wrangler.jsonc > "$PKG/meeting-point/wrangler.jsonc"
grep -q '"./app"' "$PKG/meeting-point/wrangler.jsonc"
printf 'release=%s\ncommit=%s\ngame=%s\nupstream=%s\n' "$TAG" "$(git rev-parse --short HEAD)" "$GAME" "$UPSTREAM" > "$PKG/VERSION"
cat > "$PKG/READ ME FIRST.txt" <<TXT
TS4 Access for Mac ($TAG), accessibility mod for The Sims 4 (game $GAME).
Unofficial Mac port of TS4 Access by tyjokr, LeonianUniverse and contributors.

Install (quit the game first). Easiest, in Terminal:
  curl -fsSL https://github.com/the40n8/Sims4Access-mac/releases/latest/download/install.sh | zsh

Or double-click "Install TS4 Access.command". macOS may say it cannot
verify the developer: then open System Settings > Privacy & Security, find
the message about it near the bottom and choose Open Anyway, then
double-click it again.

Then start The Sims 4 and load a household. Option+Command+A opens the
accessibility menu; "phone companion" there connects a phone.

Guides: https://github.com/the40n8/Sims4Access-mac/tree/main/docs/mac
Remove: double-click "Uninstall TS4 Access.command" (saves are not touched).
TXT
( cd "$OUT" && ditto -c -k --sequesterRsrc --keepParent TS4_Access-mac TS4_Access-mac.zip )
cp -p mac/get.sh "$OUT/install.sh"
( cd "$OUT" && shasum -a 256 TS4_Access-mac.zip install.sh > SHA256SUMS )
echo "Release $TAG built:"; ls -la "$OUT"; cat "$OUT/SHA256SUMS"

if (( PUBLISH )); then
  NOTES="$OUT/notes.md"
  cat > "$NOTES" <<MD
TS4 Access for Mac, for The Sims 4 $GAME (UI from upstream $UPSTREAM).

**Install or update** (quit the game first), in Terminal:

\`\`\`sh
curl -fsSL https://github.com/the40n8/Sims4Access-mac/releases/latest/download/install.sh | zsh
\`\`\`

Or download TS4_Access-mac.zip and double-click "Install TS4 Access.command"
(see READ ME FIRST.txt inside). Guides: [docs/mac](https://github.com/the40n8/Sims4Access-mac/tree/main/docs/mac).
MD
  git tag "$TAG" && git push -q origin "$TAG"
  gh release create "$TAG" "$OUT/TS4_Access-mac.zip" "$OUT/install.sh" "$OUT/SHA256SUMS" \
    --repo the40n8/Sims4Access-mac --title "TS4 Access for Mac $TAG" --notes-file "$NOTES" --latest
fi
