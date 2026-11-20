#!/bin/zsh
# Double-click to open SimSight setup in the browser: phone address,
# meeting point, disconnect phones. Works with the game running or not.
here="${0:A:h}"
installed="$HOME/Documents/Electronic Arts/The Sims 4/Mods/SimSight/simsight-bridge"
for bridge in "$here/simsight-bridge" "$installed" "$here/SimSight/simsight-bridge"; do
  [[ -x "$bridge" ]] && break
done
if [[ ! -x "$bridge" ]]; then
  echo "SimSight is not installed: install it first."
  exit 1
fi
"$bridge" setup
echo; echo "You can close this window."
