# Windows checks (in a test VM)

Run inside Windows with the official 64-bit embeddable Python 3.7.9 (the
game's Python) after installing a release with `install.ps1`, in the
signed-in desktop session (not over SSH: NVDA and Prism talk per session).
Results go to `C:\SimSightTest\*.txt`; NVDA started with
`--log-level=12 --log-file=C:\SimSightTest\nvda.log` logs every line it says.
Start NVDA through the shell (`start "" nvda.exe`): it needs uiAccess, which
a direct process start (a scheduled task) refuses.

- `prism_test.py`: SimSight's Prism binding finds the screen reader and
  speaks through it (NVDA log: "SimSight test. Hello from The Sims 4.").
- `mod_test.py`: the installed `SimSight.ts4script`, with stand-ins for the
  game's sims4 modules only, starts `simsight-bridge.exe` (no window), waits
  for the tunnel, opens the pairing page, and reads the code that comes
  back through the pipe without blocking; NVDA says "Pairing code: …".

Not covered here: anything inside the game (needs a VM with 3D, or a PC).

Results 2026-11-20, Windows 11 ARM64 26100 in QEMU, NVDA 2026.3.1: both
pass; the PowerShell one-liner installs, updates, refuses a tampered
download and removes; the companion listens on 127.0.0.1 only, opens the
tunnel, pairs by code over the internet, and exits with the game process.
