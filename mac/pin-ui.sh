#!/bin/zsh
# Pin the TS4 Access release whose prebuilt game interface (TS4_Access_UI.package,
# shipped as SimSight_UI.package) SimSight uses.
#   mac/pin-ui.sh build-2026-11-12_220652
# Uses the release's published .sha256 file; does not download the zip.
set -euo pipefail
cd "${0:A:h:h}"
TAG="${1:?usage: mac/pin-ui.sh <release-tag>}"
REPO=LeonianUniverse478/Sims4Access
ASSET=TS4_Access_Update.zip
SHA=$(curl -fsSL "https://github.com/$REPO/releases/download/$TAG/$ASSET.sha256" | awk '{print $1}')
[[ "$SHA" =~ ^[0-9a-f]{64}$ ]] || { echo "No valid checksum published for $TAG"; exit 1; }
python3 -I - mac/ui-package.json "$TAG" "$SHA" <<'EOF'
import json, sys
path, tag, sha = sys.argv[1:4]
data = json.load(open(path))
data.update(release=tag, sha256=sha, source_commit="(check)", game_version="(check)")
json.dump(data, open(path, 'w'), indent=2)
open(path, 'a').write('\n')
EOF
echo "Pinned $TAG ($SHA). Fill in source_commit and game_version in mac/ui-package.json."
