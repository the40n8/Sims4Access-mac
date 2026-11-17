#!/bin/zsh
# Reach the phone app through a Cloudflare quick tunnel instead of the local
# network (for Wi-Fi setups where the phone cannot reach the Mac).
# No Cloudflare account needed: you get a random https://….trycloudflare.com
# address that lives while this script runs. The game API still requires
# the pairing token carried in the link.
#   mac/phone-tunnel.sh            tunnel to the game's bridge (port 47800)
#   mac/phone-tunnel.sh --invert   QR for a light terminal background
# Needs cloudflared (brew install cloudflared). Ctrl+C closes the tunnel.
set -euo pipefail
ROOT="${0:A:h:h}"
PORT=47800
command -v cloudflared >/dev/null || { echo "cloudflared not found: brew install cloudflared"; exit 1; }
curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PORT/" || echo "Note: no bridge answers on port $PORT yet (start the game); the tunnel works once it does."

BRIDGE=""
for b in "$ROOT/dist/TS4_Access/ts4access-bridge" \
         "$HOME/Documents/Electronic Arts/The Sims 4/Mods/TS4_Access/ts4access-bridge"; do
  # Feature check without running it (an old bridge would start serving).
  if [[ -x "$b" ]] && grep -qa "public address to link to" "$b"; then BRIDGE="$b"; break; fi
done
[[ -n "$BRIDGE" ]] || { echo "No bridge with tunnel support: run mac/build.sh first."; exit 1; }

LOG=$(mktemp -t ts4access-tunnel)
cloudflared tunnel --no-autoupdate --url "http://127.0.0.1:$PORT" >"$LOG" 2>&1 &
CF=$!
trap 'kill $CF 2>/dev/null; rm -f "$LOG"; echo; echo "Tunnel closed."' EXIT INT TERM

URL=""
for _ in {1..60}; do
  URL=$(grep -Eo 'https://[a-z0-9-]+\.trycloudflare\.com' "$LOG" | head -1 || true)
  [[ -n "$URL" ]] && break
  kill -0 $CF 2>/dev/null || { cat "$LOG"; exit 1; }
  sleep 0.5
done
[[ -n "$URL" ]] || { echo "Cloudflare did not give an address:"; tail -5 "$LOG"; exit 1; }

# A fresh name takes a few seconds to be published. A phone that looks it up
# too early caches "not found" for minutes, so show the QR only once public
# DNS has it and a request through the tunnel succeeds.
HOST=${URL#https://}
echo "Waiting for Cloudflare to publish $HOST…"
READY=""
for _ in {1..60}; do
  IP=$(dig +short "$HOST" @1.1.1.1 2>/dev/null | grep -Eo '^[0-9.]+$' | head -1 || true)
  if [[ -n "$IP" ]] && [[ "$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 --resolve "$HOST:443:$IP" "$URL/")" == 200 ]]; then
    READY=1; break
  fi
  sleep 2
done
[[ -n "$READY" ]] || echo "Still not published after 2 minutes; the link may need another minute."
sleep 3

OUT=$("$BRIDGE" link --base "$URL" "$@")
print -r -- "$OUT"
print -r -- "$OUT" | grep -m1 '^http' | tr -d '\n' | pbcopy
echo
echo "Scan with the phone's camera (any network), or paste the copied link."
echo "The address changes each time this script starts. Ctrl+C to close."
wait $CF
