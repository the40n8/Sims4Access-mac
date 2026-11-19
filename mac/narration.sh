#!/bin/zsh
# Switch SimSight narration without opening the game's settings.
#   ./narration.sh on      automatic narration (hover mode)
#   ./narration.sh off     silent
#   ./narration.sh status
# Takes effect at next game start, or immediately after typing
# simsight.reload in the cheat console (Cmd+Shift+C). In game you can also type
# simsight.speechmode off  /  simsight.speechmode hover
set -euo pipefail
CFG="$HOME/Documents/Electronic Arts/The Sims 4/Mods/SimSight/SimSight.cfg"

case "${1:-status}" in
  on)  mode=hover ;;
  off) mode=off ;;
  status)
    [[ -f "$CFG" ]] && grep -E '^mode *=' "$CFG" || echo "No settings file yet (defaults: narration on)."
    exit 0 ;;
  *) echo "usage: $0 on|off|status"; exit 2 ;;
esac

if [[ -f "$CFG" ]] && grep -qE '^mode *=' "$CFG"; then
  sed -i '' -E "s/^mode *=.*/mode = $mode/" "$CFG"
else
  printf '[speech]\nmode = %s\n' "$mode" >> "$CFG"
fi
echo "Narration: $mode"
