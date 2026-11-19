#!/bin/zsh
# Remove SimSight from the Mods folder. Saves are not touched.
# Script mods stay enabled unless you pass --restore-options, which puts back
# the Options.ini from the most recent install backup.
set -euo pipefail

USERDIR="$HOME/Documents/Electronic Arts/The Sims 4"
DEST="$USERDIR/Mods/SimSight"

if pgrep -f "The Sims 4.app/Contents/MacOS" >/dev/null; then
  echo "The Sims 4 is running. Quit the game first. Nothing was changed."; exit 1
fi
if [[ -d "$DEST" ]]; then
  PARK="$USERDIR/SimSight Backups/removed-$(date +%Y%m%d-%H%M%S)"
  mkdir -p "$PARK"
  mv "$DEST" "$PARK/"            # moved, not deleted: keeps your settings and log
  echo "Moved mod to: $PARK"
fi
if [[ "${1:-}" == "--restore-options" ]]; then
  LAST=$(ls -d "$USERDIR/SimSight Backups"/2* 2>/dev/null | tail -1)
  [[ -n "$LAST" && -f "$LAST/Options.ini" ]] && cp -p "$LAST/Options.ini" "$USERDIR/Options.ini" && echo "Restored Options.ini from $LAST"
fi
