#!/bin/zsh
# Install SimSight Companion (the TS4 Access mod and its helpers) into The Sims 4 Mods folder: from a release folder
# (TS4_Access next to this script) or from the source tree (dist/TS4_Access,
# made by mac/build.sh). Reversible: backs up Options.ini and saves first;
# uninstall.sh undoes it. Your settings (TS4_Access.cfg) are kept.
#
# TS4A_SPEAK=1 says each step with the Mac voice (the one-command installer
# turns it on when VoiceOver is off); TS4A_SPEAK=0 keeps it quiet.
set -euo pipefail

HERE="${0:A:h}"
if [[ -f "$HERE/TS4_Access/TS4_Access.ts4script" ]]; then
  SRC="$HERE/TS4_Access"
else
  SRC="${HERE:h}/dist/TS4_Access"
fi
[[ -f "$SRC/TS4_Access.ts4script" ]] || { echo "Run mac/build.sh first."; exit 1; }
USERDIR="$HOME/Documents/Electronic Arts/The Sims 4"
MODS="$USERDIR/Mods"
DEST="$MODS/TS4_Access"
STAMP=$(date +%Y%m%d-%H%M%S)
BACKUP="$USERDIR/TS4 Access Backups/$STAMP"

tell() {
  print -r -- "$1"
  [[ "${TS4A_SPEAK:-0}" == 1 ]] && say -- "$1" 2>/dev/null || true
}

if pgrep -f "The Sims 4.app/Contents/MacOS" >/dev/null; then
  tell "The Sims 4 is running. Save, quit the game, then try again. Nothing was changed."; exit 1
fi
[[ -f "$USERDIR/Options.ini" ]] || {
  tell "The Sims 4 settings were not found. Start the game once, quit it, then try again. Nothing was changed."; exit 1; }

# The mod is made for one game version; a newer game may break parts of it.
if [[ -f "$HERE/VERSION" && -f "$USERDIR/GameVersion.txt" ]]; then
  want=$(sed -n 's/^game=//p' "$HERE/VERSION")
  have=$(LC_ALL=C tr -cd '0-9.\n' < "$USERDIR/GameVersion.txt" | grep -Eo '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)
  if [[ -n "$want" && -n "$have" && "$have" != "$want" ]]; then
    tell "Note: this SimSight Companion is made for game version $want, your game is $have. It usually works; if something breaks, look for a newer SimSight Companion."
  fi
fi

tell "Step 1 of 3: backing up your game settings and saves."
mkdir -p "$BACKUP"
cp -p "$USERDIR/Options.ini" "$BACKUP/Options.ini"
[[ -d "$USERDIR/saves" ]] && cp -Rp "$USERDIR/saves" "$BACKUP/saves"   # copy only; saves are never modified
ls -la "$MODS" > "$BACKUP/Mods-listing.txt" 2>/dev/null || true
if [[ -d "$DEST" ]]; then
  mkdir -p "$BACKUP/Mods"
  cp -Rp "$DEST" "$BACKUP/Mods/"
fi
print -r -- "Backup: $BACKUP"

tell "Step 2 of 3: copying the mod."
mkdir -p "$DEST"
for f in TS4_Access.ts4script TS4_Access_UI.package ts4access-say ts4access-bridge cloudflared cloudflared-LICENSE.txt README.txt CONTROLS.txt CHEATS.txt LICENSE.txt README-MAC.md; do
  cp -p "$SRC/$f" "$DEST/$f.new" && mv -f "$DEST/$f.new" "$DEST/$f"
done
cp -p "$HERE/uninstall.sh" "$DEST/uninstall.sh" 2>/dev/null || true
# The optional meeting point, ready to deploy (docs/mac/MEETING-POINT.md).
if [[ -d "$HERE/meeting-point" ]]; then
  rm -rf "$DEST/meeting-point" && cp -R "$HERE/meeting-point" "$DEST/meeting-point"
fi
chmod +x "$DEST/ts4access-say" "$DEST/ts4access-bridge" "$DEST/cloudflared" "$DEST/uninstall.sh" 2>/dev/null || true
# Downloaded with a browser, macOS would refuse to start the mod's helpers.
xattr -dr com.apple.quarantine "$DEST" 2>/dev/null || true

tell "Step 3 of 3: allowing mods in the game settings."
# Same change the Windows TS4 Access Helper makes: allow mods and script mods.
sed -i '' -E 's/^([[:space:]]*scriptmodsenabled[[:space:]]*=[[:space:]]*)[0-9]+/\11/; s/^([[:space:]]*modsdisabled[[:space:]]*=[[:space:]]*)[0-9]+/\10/' "$USERDIR/Options.ini"
grep -E '^(scriptmodsenabled|modsdisabled)' "$USERDIR/Options.ini" || true
print -r -- "Installed to: $DEST"
tell "SimSight Companion is installed. Start The Sims 4 and load a household. Then Option Command A opens the accessibility menu."
