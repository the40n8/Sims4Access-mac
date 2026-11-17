#!/bin/zsh
# Copy the phone link to the clipboard (and print where to send it).
# The link is written by the bridge each time the game starts; it stays the
# same across sessions, so the phone can keep it as a home-screen app.
#   mac/phone-link.sh          copy the best link
#   mac/phone-link.sh --all    list every address (if the first does not open)
set -euo pipefail
FILE="$HOME/Library/Application Support/TS4AccessCompanion/phone-link.txt"
[[ -f "$FILE" ]] || { echo "No link yet: start The Sims 4 once with the mod installed."; exit 1; }
if [[ "${1:-}" == "--all" ]]; then
  cat "$FILE"
  exit 0
fi
head -1 "$FILE" | tr -d '\n' | pbcopy
echo "Phone link copied. Send it to the phone (AirDrop, Messages) and open it in Safari."
echo "Then: Share > Add to Home Screen. The phone must be on the same Wi-Fi as this Mac."
