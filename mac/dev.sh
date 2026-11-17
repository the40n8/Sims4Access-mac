#!/bin/zsh
# Change the companion while the game runs (no game restart).
#   mac/dev.sh web       phone app with live reload: open the printed link on the phone
#   mac/dev.sh bridge    rebuild the bridge and swap it in (the mod restarts it)
#   mac/dev.sh speaker   rebuild the Mac voice and swap it in
# Python mod changes still need: mac/build.sh, quit the game, mac/install.sh.
set -euo pipefail
ROOT="${0:A:h:h}"
COMP="$ROOT/companion"
MODS="$HOME/Documents/Electronic Arts/The Sims 4/Mods/TS4_Access"
LINKS="$HOME/Library/Application Support/TS4AccessCompanion/phone-link.txt"

# Replace a binary that may be running: write next to it, then rename over it
# (copying over a running binary on macOS can get the new one killed).
swap() {
  cp "$1" "$MODS/$2.new" && chmod +x "$MODS/$2.new" && mv -f "$MODS/$2.new" "$MODS/$2"
}

case "${1:-}" in
  web)
    [[ -f "$LINKS" ]] || { echo "Start the game first (the bridge writes the link)."; exit 1; }
    LINK=$(head -1 "$LINKS")
    echo "On the phone, open (live reload, same pairing token):"
    echo "  ${LINK/:47800\//:5173/}"
    cd "$COMP/web" && exec pnpm dev
    ;;
  bridge)
    (cd "$COMP/web" && pnpm build >/dev/null)
    rm -rf "$COMP/bridge/internal/webui/dist" && cp -R "$COMP/web/build" "$COMP/bridge/internal/webui/dist"
    (cd "$COMP/bridge" && go build -o "$ROOT/mac/.cache/ts4access-bridge-dev" ./cmd/ts4access-bridge)
    swap "$ROOT/mac/.cache/ts4access-bridge-dev" ts4access-bridge
    pkill -f "$MODS/ts4access-bridge" || true
    echo "Bridge swapped; the game restarts it within a few seconds and the phone reconnects."
    ;;
  speaker)
    swiftc -O -o "$ROOT/mac/.cache/ts4access-say-dev" "$COMP/speaker/TS4AccessSay.swift"
    swap "$ROOT/mac/.cache/ts4access-say-dev" ts4access-say
    pkill -f "$MODS/ts4access-say" || true
    echo "Voice swapped; the next sentence starts the new one."
    ;;
  *)
    sed -n '2,6p' "$0"; exit 2
    ;;
esac
