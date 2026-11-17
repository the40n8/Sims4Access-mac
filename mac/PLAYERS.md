# TS4 Access for macOS (unofficial port)

https://github.com/the40n8/Sims4Access-mac

TS4 Access by tyjokr, LeonianUniverse and contributors (MIT licence),
source: https://github.com/LeonianUniverse478/Sims4Access (branch `dev`,
commit adb868c = release build-2026-11-12_220652, for game 1.129.39).

This port changes the speech output: instead of Windows screen readers
(Prism) or the Windows SAPI voice, the mod starts `ts4access-say`, a tiny
macOS companion that speaks with the system voices (AVSpeechSynthesizer).
The UI package is the official, unmodified release file.

## Controls that matter on Mac

Cheat console: Cmd+Shift+C (Ctrl+Shift+C on some keyboards), then type:

- `a11y.speechmode off` / `a11y.speechmode hover` — silence / resume narration
- `a11y.simactions on|off` — announce what the active Sim starts doing
- `a11y.reload` — re-read `TS4_Access.cfg` after editing it
- `a11y.log` — diagnostic log, `Mods/TS4_Access/TS4_Access.log`

Voice: edit `Mods/TS4_Access/TS4_Access.cfg`, section `[voice]`:
`name = Thomas` (or Amelie, Audrey… for French; Samantha, Ava… for English),
`rate = 3` (-10..10), `volume = 70`. List voices:
`Mods/TS4_Access/ts4access-say --list-voices`.
Better voices: System Settings > Accessibility > Spoken Content > System voice > Manage Voices.

## Mac defaults (shared play, sighted player on the mouse)

Keyboard navigation, simplified menus, scanner, travel list and build
controls start OFF (the Windows keyboard hooks don't exist on macOS, and
simplified menus would change menus for the mouse player). Narration is in
`hover` mode, verbosity `brief`, Sim actions ON, logging ON (prototype).

## Known limitations

- Keyboard shortcuts that rely on Windows hooks (Ctrl+Alt+…, scanner,
  build cursor) do not work. Shortcuts handled inside the game UI may.
- Main menu / Create a Sim narration is off on macOS (its startup channel
  needs a Windows API, and its listener thread crashed the Mac game).
- No braille, no VoiceOver integration (VoiceOver can stay on, but the mod
  speaks with its own voice).
- Not tested by the TS4 Access team on macOS. Report problems to this port,
  not upstream, unless they also happen on Windows.
