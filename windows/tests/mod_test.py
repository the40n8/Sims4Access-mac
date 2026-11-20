"""The installed SimSight.ts4script's companion code on Windows, outside the
game: stand-ins for the game's sims4 modules only. Starts the real
simsight-bridge.exe through the mod, then opens the pairing page: the code
must come back through the pipe and be spoken by NVDA."""
import os, sys, time, types, traceback, urllib.request
out = open(r'C:\SimSightTest\mod-test.txt', 'w')
def log(m):
    out.write(m + '\n'); out.flush()
def module(name, **kw):
    m = types.ModuleType(name); m.__dict__.update(kw); sys.modules[name] = m; return m
try:
    docs = os.path.join(os.environ['USERPROFILE'], 'Documents')
    mods = os.path.join(docs, 'Electronic Arts', 'The Sims 4', 'Mods', 'SimSight')
    sys.path.insert(0, os.path.join(mods, 'SimSight.ts4script'))
    module('sims4')
    module('sims4.commands', Command=lambda *a, **k: (lambda f: f), CommandType=types.SimpleNamespace(Live=1, Cheat=2))
    module('sims4.core_services', on_tick=lambda *a, **k: None)
    import simsight_bridge as bridge
    log('IS_WINDOWS=%s alive=%s' % (bridge.IS_WINDOWS, bridge.alive()))
    log('bridge exe: %s' % bridge._find(bridge.BRIDGE))
    for _ in range(30):
        if bridge.alive():
            break
        time.sleep(0.2); bridge.start()
    log('alive after start: %s' % bridge.alive())
    import simsight_speech as speech
    log('speech backend: %s' % speech.init())
    time.sleep(2)
    bridge.mirror('Mirror test line', False)
    page, t0 = '', time.time()
    while 'class="pin"' not in page and time.time() - t0 < 60:   # the tunnel opens first
        bridge._read_commands(); time.sleep(1)
        page = urllib.request.urlopen('http://127.0.0.1:47800/pair').read().decode('utf-8', 'replace')
    log('pair page has code after %d s: %s' % (time.time() - t0, 'class="pin"' in page))
    t0 = time.time()
    while time.time() - t0 < 6:
        bridge._read_commands()      # what the game tick does, every 0.1 s
        time.sleep(0.1)
    log('read loop never blocked (6 s of polling done)')
    time.sleep(3)
except Exception:
    log('FAILED: ' + traceback.format_exc())
log('done')
