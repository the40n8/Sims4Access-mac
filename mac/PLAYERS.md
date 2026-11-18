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

## Phone companion

While the game runs, a phone can follow and control it (any network, through
a Cloudflare quick tunnel): household status, a live feed read by VoiceOver,
game speed, controlled Sim, cancelling actions, and the game's own interaction
menus for Sims and objects (Play). House describes the home room by room: each
floor, the rooms on it and where they sit, which doors join them, and what is
in each room (and the garden and pool), with its interactions one tap away.
When the game asks something (which recipe, who to invite, a name to type),
the question appears at the top of the phone with its choices and buttons,
and the corner notifications are read out too. Play also lists the Sim's
own phone (calls, invitations, travel, services). Sims shows skills,
relationships and work performance. When the game starts, the Mac voice says "Phone
link ready". Open the cheat console (Cmd+Shift+C) and type `a11y.companion`
(or open the accessibility menu with `a11y.hub` > phone companion): a page
with a QR code opens on the Mac. Scan it with the phone, then in Safari:
Share > Add to Home Screen. The address changes each game session. Language follows the phone
(French, English, Spanish) and can be changed in Settings.

## Shortcuts and menu bar icon

While the game runs (no need to leave it):

- **Option+Command+S**: accessibility speech off or back on ("Speech off." /
  "Speech on.").
- **Option+Command+A**: the TS4 Access menu (same as `a11y.hub`).

The keys follow the letters printed on the keyboard (AZERTY too). An
accessibility icon in the macOS menu bar offers the same two actions plus
"Phone companion…", which opens the page with the phone QR code. In full
screen, move the pointer to the top of the screen to show the menu bar. The
phone has the speech switch at the top of Live and Play. While the game runs,
these two shortcuts belong to TS4 Access in every app.

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
