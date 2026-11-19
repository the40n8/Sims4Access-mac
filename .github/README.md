# SimSight Companion

Play The Sims 4 with a screen reader, on the Mac and from your phone. An
independent project based on **[TS4 Access](https://github.com/LeonianUniverse478/Sims4Access)**,
the screen-reader mod for The Sims 4 by tyjokr, LeonianUniverse and contributors
(MIT licence; their credits and licence are kept here). In the game, the mod
still calls itself "TS4 Access".
It speaks menus, dialogs, tooltips, notifications and what your Sim is doing,
using the macOS system voices.

Made for playing together: a sighted player on the mouse, a blind player
hearing what happens. Keyboard-only play (scanner, build cursor, Ctrl+Alt
shortcuts) is Windows-only for now.

**Status:** works in game on an Apple Silicon Mac, game 1.129.39.1230
(hover narration, Sim actions). The phone companion works in game (status,
feed, speed, Play menus, House map); its newest parts (game questions,
notifications, Needs attention) are still being tried. Early: expect rough
edges.

**Players:** read the [phone guide](../docs/mac/PHONE-GUIDE.md) (also in the
app's Help tab). To ask ChatGPT or another assistant about the mod, point it
at [`llms.txt`](https://raw.githubusercontent.com/the40n8/simsight-companion/main/llms.txt).

## Install (players)

Quit The Sims 4, open Terminal (Command+Space, type Terminal, Return), paste
this line and press Return:

```sh
curl -fsSL https://github.com/the40n8/simsight-companion/releases/latest/download/install.sh | zsh
```

It says each step (VoiceOver reads them; with VoiceOver off, the Mac voice
says them), asks you to press Return, backs up your settings and saves,
installs the mod and allows mods in the game. Run the same line again to
update; add `-s uninstall` at the end to remove it. Then start the game and
load a household: Option+Command+A opens the accessibility menu.

Without Terminal: download `SimSight-Companion-mac.zip` from the
[latest release](https://github.com/the40n8/simsight-companion/releases/latest)
and double-click **Install SimSight Companion.command** (macOS may ask you to allow
it in System Settings > Privacy & Security > Open Anyway).

Nothing else to set up: the voice follows the game's language, and the phone
app works right away (QR code or 6-digit code each session). Optional later:
a [fixed address](../docs/mac/FIXED-ADDRESS.md) (your own domain) or a
[meeting point](../docs/mac/MEETING-POINT.md) (free Cloudflare account).

## How it differs from upstream

| | Windows (upstream) | macOS (this fork) |
|---|---|---|
| Speech | NVDA/JAWS via Prism, or a SAPI voice | `ts4access-say` companion (AVSpeechSynthesizer) |
| Keyboard navigation, scanner | Windows key hooks | Off (the Mac game's Python has no `ctypes`) |
| Main menu / Create a Sim speech | Loopback socket | Off (Windows-only peer check; its thread crashed the Mac game) |
| UI package (patched Flash UI) | Built from `flash/` | Same file, taken from the pinned upstream release |
| Helper / updater .exe | Yes | One Terminal line (`mac/get.sh`), or `mac/install.sh` from source |

The Python changes are small `sys.platform == 'darwin'` branches in four
modules (`speech`, `settings`, `startup`, `plus_core`), so upstream changes
merge cleanly. Windows-only files are removed (see `mac/windows-paths.txt`).

## Build from source

Requirements: Xcode Command Line Tools (`swiftc`), Go, pnpm and a
**Python 3.7** (the game only loads 3.7 bytecode). On Apple Silicon, one way
to get Python 3.7:

```sh
mkdir -p ~/.local/micromamba && cd ~/.local/micromamba
curl -Ls https://micro.mamba.pm/api/micromamba/osx-64/latest | tar -xj bin/micromamba
MAMBA_ROOT_PREFIX=~/.local/micromamba ./bin/micromamba create -y \
  -p ~/.local/micromamba/envs/py37 --platform osx-64 -c conda-forge python=3.7
```

Then, with the game closed:

```sh
mac/build.sh      # -> dist/TS4_Access (mod, bridge + phone app, speaker, UI package)
mac/test.sh       # offline checks (fake game for the Mac-only mod code, Go tests)
mac/install.sh    # backs up Options.ini and saves, installs, enables script mods
mac/release.sh    # player release in dist/release (zip, SHA256SUMS, install.sh);
                  # --publish creates the GitHub release
```

Start the game, accept the mod list, load a household. Player notes:
[`mac/PLAYERS.md`](../mac/PLAYERS.md). Remove with `mac/uninstall.sh`
(`--restore-options` also restores Options.ini). `mac/narration.sh on|off`
switches narration from outside the game.

## Phone companion

While the game runs, `ts4access-bridge` serves a small web app: live
household status, a feed announced by VoiceOver, the game's own interaction
menus, a room-by-room house map, game questions answered from the phone, and
quick actions for what needs attention. See the
[phone guide](../docs/mac/PHONE-GUIDE.md). No server of ours is involved.

1. Start the game. The bridge opens a Cloudflare quick tunnel by itself
   (bundled `cloudflared`, no account) so the phone works from any network;
   the Mac voice says "Phone link ready" and how to open the QR code.
2. Show the QR code: in game, accessibility menu (`a11y.hub` in the cheat
   console) > phone companion, or `a11y.companion`; from a terminal,
   `mac/phone-link.sh`. Both open http://127.0.0.1:47800/pair (this Mac only).
3. Scan it with the phone, then in Safari: Share > Add to Home Screen. The
   tunnel address changes each game session, so scan again next time.
4. To revoke the pairing token, delete
   `~/Library/Application Support/TS4AccessCompanion/config.json`.
   `ts4access-bridge --tunnel=false` keeps everything on the local network.
   `mac/wifi-test.sh` checks whether the phone reaches the Mac directly over
   Wi-Fi (works without the game). For a fixed address instead of a random
   one each session, see [docs/mac/FIXED-ADDRESS.md](../docs/mac/FIXED-ADDRESS.md).

The first time, macOS may ask whether `ts4access-bridge` may accept incoming
connections: allow it, or the phone cannot connect.

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
Scripts/              the mod's Python (upstream + macOS branches)
  tyjokr_a11y_bridge.py   game side of the phone companion (new in this fork)
companion/
  proto/              API and pipe messages (buf, ts4access.v1)
  bridge/             Go: pipe to the game, Connect API, pairing, event detection
  web/                SvelteKit phone app (fr/en/es), embedded in the bridge
  speaker/            TS4AccessSay.swift (macOS voices)
mac/                  build, install, sync, docs
  ui-package.json     pinned upstream release for TS4_Access_UI.package
docs/mac/             player guides (phone app, keyboard, fixed address, AI helper)
docs/plus/            the original's player guides (TS4 Access Plus)
flash/                the original's patched game interface source (reference)
```

## Development

- Phone app without the game: `go run ./cmd/ts4access-bridge --demo` in
  `companion/bridge` (fake household on port 47800, link printed), then
  `pnpm dev` in `companion/web` (proxies the API to it).
- While the game runs: `mac/dev.sh web` (phone app with live reload),
  `mac/dev.sh bridge`, `mac/dev.sh speaker` (rebuilt and swapped in; the mod
  restarts the bridge if it exits). Python changes need a game restart.
- After changing `companion/proto`: `buf generate` in `companion/bridge` and
  `pnpm generate` in `companion/web`.
- Checks: `go vet ./... && go test ./...` (bridge), `pnpm verify` (web).
- Game-side rules on macOS: the game's Python has no `ctypes`, and a
  background thread importing a not-yet-loaded module crashed the game. Run
  game work from the game's tick (see `tyjokr_a11y_bridge.py`), never import
  inside a thread, and keep macOS changes as small `sys.platform == 'darwin'`
  branches so upstream merges stay clean. The original's Windows developer
  material (its build tooling, baselines and docs) is not kept in this fork:
  mac/windows-paths.txt lists it and mac/sync-upstream.sh removes it again
  after each merge.

## Licence and credits

MIT, as upstream (see `LICENSE`). TS4 Access by **tyjokr**; lead development
**LeonianUniverse**; TS4 Access Plus by **Michael Maslo**; Build cursor and
scanner additions by **DaRainbowGirlXX**. macOS port by the40n8.
Please report Mac-specific problems here, not upstream.
