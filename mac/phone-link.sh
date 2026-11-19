#!/bin/zsh
# Show the phone link as a QR code (scan it with the phone's camera) and
# copy it to the clipboard. The link stays the same across sessions, so the
# phone can keep it as a Home Screen app.
#   mac/phone-link.sh            QR + copy
#   mac/phone-link.sh --all      also list other addresses
#   mac/phone-link.sh --invert   QR for a light terminal background
#   mac/phone-link.sh --port 5173   link to the live-reload dev server (mac/dev.sh web)
set -euo pipefail
ROOT="${0:A:h:h}"

# With the game running, its bridge shows the current link (Cloudflare tunnel
# or local) as a QR code on a page: open that unless options were given.
if [[ $# -eq 0 ]] && curl -s -o /dev/null --max-time 2 http://127.0.0.1:47800/pair; then
  open http://127.0.0.1:47800/pair
  echo "Opened the phone link page (QR code) in your browser."
  exit 0
fi
BRIDGE=""
for b in "$ROOT/dist/SimSight/simsight-bridge" \
         "$HOME/Documents/Electronic Arts/The Sims 4/Mods/SimSight/simsight-bridge"; do
  # Feature check without running it (an old bridge would start serving).
  if [[ -x "$b" ]] && grep -qa "QR for terminals with a light background" "$b"; then BRIDGE="$b"; break; fi
done
[[ -n "$BRIDGE" ]] || { echo "No bridge with QR support: run mac/build.sh first."; exit 1; }
OUT=$("$BRIDGE" link "$@")
print -r -- "$OUT"
print -r -- "$OUT" | grep -m1 '^http' | tr -d '\n' | pbcopy
echo
echo "Scan with the phone's camera (same Wi-Fi as this Mac), or paste the copied link."
echo "Then in Safari: Share > Add to Home Screen."
