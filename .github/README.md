# TS4 Access for macOS

An unofficial macOS port of **[TS4 Access](https://github.com/LeonianUniverse478/Sims4Access)**,
the screen-reader mod for The Sims 4 by tyjokr, LeonianUniverse and contributors.
It speaks menus, dialogs, tooltips, notifications and what your Sim is doing,
using the macOS system voices.

Made for playing together: a sighted player on the mouse, a blind player
hearing what happens. Keyboard-only play (scanner, build cursor, Ctrl+Alt
shortcuts) is Windows-only for now.

**Status:** works in game on an Apple Silicon Mac, game 1.129.39.1230
(hover narration, Sim actions). Early: expect rough edges.

## How it differs from upstream

| | Windows (upstream) | macOS (this fork) |
|---|---|---|
| Speech | NVDA/JAWS via Prism, or a SAPI voice | `ts4access-say` companion (AVSpeechSynthesizer) |
| Keyboard navigation, scanner | Windows key hooks | Off (the Mac game's Python has no `ctypes`) |
| Main menu / Create a Sim speech | Loopback socket | Off (Windows-only peer check; its thread crashed the Mac game) |
| UI package (patched Flash UI) | Built from `flash/` | Same file, taken from the pinned upstream release |
| Helper / updater .exe | Yes | `mac/install.sh` |

The Python changes are small `sys.platform == 'darwin'` branches in four
modules (`speech`, `settings`, `startup`, `plus_core`), so upstream changes
merge cleanly. Windows-only files are removed (see `mac/windows-paths.txt`).

## Build and install

Requirements: Xcode Command Line Tools (`swiftc`) and a **Python 3.7**
(the game only loads 3.7 bytecode). On Apple Silicon, one way:

```sh
mkdir -p ~/.local/micromamba && cd ~/.local/micromamba
curl -Ls https://micro.mamba.pm/api/micromamba/osx-64/latest | tar -xj bin/micromamba
MAMBA_ROOT_PREFIX=~/.local/micromamba ./bin/micromamba create -y \
  -p ~/.local/micromamba/envs/py37 --platform osx-64 -c conda-forge python=3.7
```

Then, with the game closed:

```sh
mac/build.sh      # -> dist/TS4_Access (downloads + verifies the UI package)
mac/install.sh    # backs up Options.ini and saves, installs, enables script mods
```

Start the game, accept the mod list, load a household. Player notes:
[`mac/PLAYERS.md`](../mac/PLAYERS.md). Remove with `mac/uninstall.sh`
(`--restore-options` also restores Options.ini). `mac/narration.sh on|off`
switches narration from outside the game.

## Keeping up with upstream

```sh
mac/sync-upstream.sh   # merge upstream/dev, re-remove Windows files, report conflicts
mac/pin-ui.sh <tag>    # when upstream publishes a release matching the merged source
mac/build.sh && mac/install.sh
```

Conflicts can only come from the four patched modules: keep upstream's
change and re-apply the `darwin` branch.

## Layout

```
Scripts/            the mod's Python (upstream + macOS branches)
mac/                everything macOS: build, install, sync, companion, docs
  companion/        TS4AccessSay.swift (speech companion)
  ui-package.json   pinned upstream release for TS4_Access_UI.package
flash/, tools/, docs/   upstream UI sources, tests and developer docs (reference)
```

## Licence and credits

MIT, as upstream (see `LICENSE`). TS4 Access by **tyjokr**; lead development
**LeonianUniverse**; TS4 Access Plus by **Michael Maslo**; Build cursor and
scanner additions by **DaRainbowGirlXX**. macOS port by the40n8.
Please report Mac-specific problems here, not upstream.
