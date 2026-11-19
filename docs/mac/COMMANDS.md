# SimSight console commands (Mac)

You do not need these for normal play: **Option+Command+A** opens the
accessibility menu, and the phone app does most of the rest. The keyboard
is in [KEYBOARD.md](KEYBOARD.md).

Open the game's console with **Command+Shift+C** (Control+Shift+C on some
keyboards), type a command, press Return. Close the console with the same
keys before using menus.

## Help and settings

- `simsight.hub`: the accessibility menu.
- `simsight.companion`: the phone link page (QR code and pairing code).
- `simsight.menu window`: the settings window (simplified menus, Build controls).
- `simsight.settings`: read the current settings.
- `simsight.reload`: apply changes made by hand to `SimSight.cfg`.
- `simsight.scannermode on`: scanner browsing with the arrow keys.
- `simsight.keys on`: keyboard navigation of menus and panels.
- `simsight.simactions on`: announce what the active Sim starts doing.
- `simsight.objectnames on`: read the names of objects in the world.
- `simsight.captions on`: show the spoken text on screen.

Use `off` instead of `on` to turn an option off.

## Speech

- `simsight.speechmode hover`: read what the pointer is on, automatically.
- `simsight.speechmode manual`: read only when asked.
- `simsight.speechmode off`: silence (Option+Command+S does the same).
- `simsight.tooltipdelay 1.5`: wait 1.5 seconds before longer descriptions.
- `simsight.voice rate 3` (-10 to 10), `simsight.voice volume 70` (0 to 100),
  `simsight.voice name Thomas` (part of a macOS voice name).
  Better voices: System Settings > Accessibility > Spoken Content > System
  voice > Manage Voices.

## Quick access

- `simsight.phonelist`: the Sim's phone actions.
- `simsight.phone`: answer a waiting call or text.
- `simsight.household`: choose the active Sim.
- `simsight.contacts`: known contacts.
- `simsight.funds`: household funds.
- `simsight.scan rescan`: refresh the scanner.
- `simsight.log`: diagnostic log on or off (`Mods/SimSight/SimSight.log`).

Settings commands are saved; updating the mod keeps `SimSight.cfg`.
