#!/bin/zsh
# Install dist/TS4_Access (made by mac/build.sh) into The Sims 4 Mods folder.
# Reversible: backs up Options.ini and saves first; uninstall.sh undoes it.
set -euo pipefail

SRC="${0:A:h:h}/dist/TS4_Access"
[[ -f "$SRC/TS4_Access.ts4script" ]] || { echo "Run mac/build.sh first."; exit 1; }
USERDIR="$HOME/Documents/Electronic Arts/The Sims 4"
MODS="$USERDIR/Mods"
DEST="$MODS/TS4_Access"
STAMP=$(date +%Y%m%d-%H%M%S)
BACKUP="$USERDIR/TS4 Access Backups/$STAMP"

if pgrep -f "The Sims 4.app/Contents/MacOS" >/dev/null; then
  echo "The Sims 4 is running. Quit the game first. Nothing was changed."; exit 1
fi
[[ -f "$USERDIR/Options.ini" ]] || { echo "Options.ini not found. Start the game once, quit, then retry."; exit 1; }

mkdir -p "$BACKUP"
cp -p "$USERDIR/Options.ini" "$BACKUP/Options.ini"
cp -Rp "$USERDIR/saves" "$BACKUP/saves"          # copy only; saves are never modified
ls -la "$MODS" > "$BACKUP/Mods-listing.txt"
if [[ -d "$DEST" ]]; then
  mkdir -p "$BACKUP/Mods"
  cp -Rp "$DEST" "$BACKUP/Mods/"
fi
echo "Backup: $BACKUP"

mkdir -p "$DEST"
# Keep the user's TS4_Access.cfg (settings) if one exists.
for f in TS4_Access.ts4script TS4_Access_UI.package ts4access-say ts4access-bridge cloudflared cloudflared-LICENSE.txt README.txt CONTROLS.txt CHEATS.txt LICENSE.txt README-MAC.md; do
  cp -p "$SRC/$f" "$DEST/$f.new" && mv -f "$DEST/$f.new" "$DEST/$f"
done
chmod +x "$DEST/ts4access-say" "$DEST/ts4access-bridge" "$DEST/cloudflared"

# Same change the Windows TS4 Access Helper makes: allow mods and script mods.
sed -i '' -E 's/^([[:space:]]*scriptmodsenabled[[:space:]]*=[[:space:]]*)[0-9]+/\11/; s/^([[:space:]]*modsdisabled[[:space:]]*=[[:space:]]*)[0-9]+/\10/' "$USERDIR/Options.ini"
grep -E '^(scriptmodsenabled|modsdisabled)' "$USERDIR/Options.ini"
echo "Installed to: $DEST"
