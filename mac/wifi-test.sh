#!/bin/zsh
# Can the phone reach this Mac directly over Wi-Fi (no internet tunnel)?
# Works without the game: starts the bridge in demo mode (a fake household)
# when the game is not running, shows a QR code with the Mac's Wi-Fi address,
# and says when the phone connects.
#   mac/wifi-test.sh            test (Ctrl+C to stop)
#   mac/wifi-test.sh --invert   QR for a light terminal background
set -euo pipefail
ROOT="${0:A:h:h}"
PORT=47800
BRIDGE=""
for b in "$ROOT/dist/SimSight/simsight-bridge" \
         "$HOME/Documents/Electronic Arts/The Sims 4/Mods/SimSight/simsight-bridge"; do
  [[ -x "$b" ]] && { BRIDGE="$b"; break; }
done
[[ -n "$BRIDGE" ]] || { echo "No bridge: run mac/build.sh first."; exit 1; }

DEMO_PID=""
cleanup() { [[ -n "$DEMO_PID" ]] && kill "$DEMO_PID" 2>/dev/null || true; }
trap cleanup EXIT INT TERM
if curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PORT/pair"; then
  echo "The game's bridge is running: testing with it."
else
  echo "Game not running: starting a demo bridge (fake household)."
  "$BRIDGE" --demo --tunnel=false --port "$PORT" >/dev/null 2>&1 &
  DEMO_PID=$!
  for _ in {1..20}; do curl -s -o /dev/null --max-time 1 "http://127.0.0.1:$PORT/pair" && break; sleep 0.3; done
fi

"$BRIDGE" link --port "$PORT" "$@"
echo
echo "1. Phone on the same Wi-Fi as this Mac (not a guest network)."
echo "2. Scan the QR code with the phone's camera and open the link."
echo "Waiting for the phone... (Ctrl+C to stop)"
seen=""
mine=" $(ifconfig | awk '/inet /{print $2}' | tr '\n' ' ')"   # this Mac's own addresses
for _ in {1..600}; do
  # Connections to the bridge from another device on the network.
  peers=$(netstat -an -p tcp | awk -v p=".$PORT" '$4 ~ p"$" && $6 == "ESTABLISHED" {print $5}' |
          sed 's/\.[0-9]*$//' | grep -v -E '^(127\.0\.0\.1|::1)$' | sort -u || true)
  for ip in ${(f)peers}; do
    [[ -z "$ip" || "$seen" == *" $ip"* || "$mine" == *" $ip "* ]] && continue
    seen="$seen $ip"
    echo "✅ Connected from $ip: the direct Wi-Fi link works."
  done
  sleep 1
done
[[ -n "$seen" ]] || echo "❌ No connection after 10 minutes. Check the phone is on the same Wi-Fi (not a guest network), and the Mac firewall allows simsight-bridge."
