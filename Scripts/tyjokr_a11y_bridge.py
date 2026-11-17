"""Companion bridge (macOS fork): game <-> ts4access-bridge <-> phone.

The bridge is a child process started here, on the game's loading thread.
- Speech: the speech module writes its usual "I..."/"Q..." lines to the
  bridge, which forwards them to the macOS voice (see SpeechPipe).
- State: a JSON snapshot of the household (needs, mood, moodlets, actions,
  clock, funds), sent when it changes, at most once per second.
- Commands from the phone arrive on the bridge's stdout and run here.

Everything that touches the game runs from the game's own tick on the
simulation thread. No threads: the Mac game crashes when a background
thread imports a module from a zip while the simulation runs Python.
Message format: companion/proto/ts4access/v1/pipe.proto (protobuf JSON).
"""
import json
import os
import re
import sys
import threading
import time

try:
    import subprocess
except ImportError:
    subprocess = None

import tyjokr_a11y_speech as speech

IS_MAC = sys.platform == 'darwin'
BRIDGE = 'ts4access-bridge'
SPEAKER = 'ts4access-say'
STATE_INTERVAL = 1.0      # seconds between state snapshots (sent only if changed)
POLL_INTERVAL = 0.1       # seconds between command polls
LOC_PER_TICK = 20         # string-table keys requested from the UI per poll

_proc = None
_voice = None             # last voice settings, re-sent to a restarted bridge
_last_spawn = 0.0
RESPAWN_DELAY = 2.0       # seconds before restarting a bridge that exited
_write_lock = threading.Lock()
_read_buf = b''
_last_state_json = None
_last_state_time = 0.0
_last_poll = 0.0
_seq = 0


def log(message):
    speech.log_line('bridge: ' + message)


# ---------------------------------------------------------------- process

def _find(name):
    for folder in (speech.MODS_FOLDER, os.path.join(speech._mods_root(), 'TS4_Access')):
        path = os.path.join(folder, name)
        if os.path.isfile(path):
            return path
    return None


def start():
    """Start the bridge if needed. Returns True when it is running."""
    global _proc, _last_spawn, _read_buf, _last_state_json
    if _proc is not None:
        if _proc.poll() is None:
            return True
        # It exited (crash, or replaced during development): start it again,
        # at most every RESPAWN_DELAY seconds.
        if time.monotonic() - _last_spawn < RESPAWN_DELAY:
            return False
        log('restarting (previous one exited with {})'.format(_proc.returncode))
        _proc = None
        _read_buf = b''
        _last_state_json = None   # the new bridge needs a full state
    _last_spawn = time.monotonic()
    path = _find(BRIDGE)
    if not IS_MAC or subprocess is None or path is None:
        return False
    args = [path]
    speaker = _find(SPEAKER)
    if speaker:
        args += ['--speaker', speaker]
    try:
        _proc = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                 stderr=subprocess.DEVNULL, close_fds=True)
        os.set_blocking(_proc.stdout.fileno(), False)
    except Exception as ex:
        log('start failed: {!r}'.format(ex))
        _proc = None
        return False
    log('started (pid {})'.format(_proc.pid))
    send({'hello': {'modVersion': 'mac', 'gameVersion': _game_version()}})
    if _voice is not None:
        send({'voice': _voice})
    return True


def alive():
    return _proc is not None and _proc.poll() is None


def write_line(data):
    """One complete line to the bridge. Safe from any thread (speech timers)."""
    if not alive():
        return False
    with _write_lock:
        try:
            _proc.stdin.write(data + b'\n')
            _proc.stdin.flush()
            return True
        except Exception as ex:
            log('write failed: {!r}'.format(ex))
            return False


def send(message):
    return write_line(json.dumps(message, separators=(',', ':'), ensure_ascii=False).encode('utf-8'))


def set_voice(rate, volume, name, lang):
    global _voice
    _voice = {'rate': int(rate), 'volume': int(volume), 'name': name or '', 'lang': lang or ''}
    send({'voice': _voice})


class SpeechPipe(object):
    """Stands in for the fallback-voice process the speech module expects:
    `.stdin.write()/.flush()` send lines to the bridge, `.kill()` (used when
    voice settings change) leaves the bridge running; the speech module then
    calls set_voice again with the new settings."""

    def __init__(self):
        self.stdin = self
        self._buf = b''

    def write(self, data):
        self._buf += data
        while b'\n' in self._buf:
            line, self._buf = self._buf.split(b'\n', 1)
            write_line(line)

    def flush(self):
        pass

    def poll(self):
        return None if alive() else 1

    def kill(self):
        pass


def _game_version():
    try:
        import sims4.common
        return str(sims4.common.get_version() if hasattr(sims4.common, 'get_version') else '')
    except Exception:
        return ''


# ---------------------------------------------------------------- tick

def _tick():
    """Runs on the game's tick (simulation thread), every frame."""
    global _last_poll, _last_state_time
    if not alive() and not start():
        return
    now = time.monotonic()
    if now - _last_poll >= POLL_INTERVAL:
        _last_poll = now
        _read_commands()
        _request_localization()
    if now - _last_state_time >= STATE_INTERVAL:
        _last_state_time = now
        _publish_state()


def _install_tick():
    import sims4.core_services as core
    original = core.on_tick
    if getattr(original, '_ts4access_bridge', False):
        return

    def on_tick(*args, **kwargs):
        result = original(*args, **kwargs)
        try:
            _tick()
        except Exception as ex:
            log('tick failed: {!r}'.format(ex))
        return result

    on_tick._ts4access_bridge = True
    core.on_tick = on_tick


def _read_commands():
    global _read_buf
    try:
        chunk = os.read(_proc.stdout.fileno(), 65536)
    except BlockingIOError:
        return
    except Exception as ex:
        log('read failed: {!r}'.format(ex))
        return
    if not chunk:
        return
    _read_buf += chunk
    while b'\n' in _read_buf:
        line, _read_buf = _read_buf.split(b'\n', 1)
        if line.strip():
            _handle(line)


def _handle(line):
    try:
        cmd = json.loads(line.decode('utf-8'))
    except Exception as ex:
        log('bad command: {!r}'.format(ex))
        return
    cmd_id = cmd.get('id', '')
    try:
        ok, reason = _run(cmd)
    except Exception as ex:
        log('command failed: {!r}'.format(ex))
        ok, reason = False, 'error'
    outcome = {'ok': True} if ok else {'reason': reason}
    send({'result': {'id': cmd_id, 'outcome': outcome}})
    # Let the phone see the effect right away.
    global _last_state_time
    _last_state_time = 0.0


def _in_game():
    import services
    zone = services.current_zone()
    return (zone is not None and getattr(zone, 'is_zone_running', False)
            and services.active_household() is not None)


_SPEEDS = {'SPEED_PAUSED': 'paused', 'SPEED_NORMAL': 'one', 'SPEED_FAST': 'two', 'SPEED_ULTRA': 'three'}


def _run(cmd):
    if 'setNarration' in cmd:
        narration = cmd['setNarration'].get('narration', {})
        speech.set_mode('hover' if narration.get('macVoice') else 'off')
        import tyjokr_a11y_settings as settings
        settings.set('sim_actions', bool(narration.get('simActions')))
        return True, ''
    if not _in_game():
        return False, 'not_in_game'
    if 'setSpeed' in cmd:
        name = _SPEEDS.get(cmd['setSpeed'].get('speed', ''))
        if name is None:
            return False, 'error'
        return _set_speed(name), 'error'
    if 'selectSim' in cmd:
        import tyjokr_a11y_hub as hub
        return hub.select_sim(int(cmd['selectSim'].get('simId', '0'))), 'sim_unavailable'
    if 'cancelAction' in cmd:
        import services
        import tyjokr_a11y_hub as hub
        c = cmd['cancelAction']
        sim = services.get_active_sim()
        if sim is None or str(sim.id) != c.get('simId'):
            return False, 'sim_unavailable'
        return hub.cancel_actions(sim.id, (int(c.get('actionId', '0')),)), 'not_cancelable'
    return False, 'unknown_command'


def _set_speed(name):
    """Through the game's own clock.setspeed command, like the speed buttons."""
    try:
        from server_commands import clock_commands
        clock_commands.set_speed(name)
        return True
    except Exception as ex:
        log('clock.setspeed failed, using the clock directly: {!r}'.format(ex))
    import services
    from clock import ClockSpeedMode, GameSpeedChangeSource
    mode = {'paused': ClockSpeedMode.PAUSED, 'one': ClockSpeedMode.NORMAL,
            'two': ClockSpeedMode.SPEED2, 'three': ClockSpeedMode.SPEED3}[name]
    services.game_clock_service().set_clock_speed(mode, source=GameSpeedChangeSource.GAMEPLAY)
    return True


# ---------------------------------------------------------------- localization
# Python only sees string-table keys; the patched UI resolves them in the
# player's language through the scanner's loc channel (A11y|loc / a11y.loc_enc).

_loc_wanted = set()
_loc_requested = set()


def _loc_cache():
    scanner = sys.modules.get('tyjokr_a11y_scanner')
    return getattr(scanner, '_loc_cache', {}) if scanner else {}


def _request_localization():
    scanner = sys.modules.get('tyjokr_a11y_scanner')
    if scanner is None or not _loc_wanted:
        return
    cache = scanner._loc_cache
    for key in list(_loc_wanted)[:LOC_PER_TICK]:
        _loc_wanted.discard(key)
        if key in cache or key in _loc_requested:
            continue
        _loc_requested.add(key)
        scanner._send_ui('A11y|loc|{}|{}'.format(key, key))


def _loc_value(value):
    """A LocalizedString proto from a tuned value (proto or factory)."""
    if value is None:
        return None
    if callable(value) and not hasattr(value, 'hash'):
        try:
            value = value()
        except Exception:
            return None
    return value if hasattr(value, 'hash') else None


_TOKEN = re.compile(r'\{([MF])(\d+)\.([^}]*)\}')
_FIELD = re.compile(r'\{(\d+)\.(\w+)\}')


def _fill_tokens(template, tokens):
    """Fill the common Sims placeholders: {0.SimFirstName}, {M0.he}{F0.she}..."""
    def token(i):
        return tokens[i] if 0 <= i < len(tokens) else None

    def gender(match):
        t = token(int(match.group(2)))
        female = bool(getattr(t, 'is_female', False)) if t is not None else False
        return match.group(3) if (match.group(1) == 'F') == female else ''

    def field(match):
        t = token(int(match.group(1)))
        if t is None:
            return ''
        name = match.group(2)
        first, last = getattr(t, 'first_name', ''), getattr(t, 'last_name', '')
        if name == 'SimFirstName':
            return first
        if name == 'SimLastName':
            return last
        if name in ('SimName', 'SimFullName'):
            return '{} {}'.format(first, last).strip()
        if name == 'Number':
            return str(getattr(t, 'number', ''))
        if name in ('String', 'RawText'):
            return getattr(t, 'raw_text', '') or ''
        return ''

    text = _TOKEN.sub(gender, template)
    text = _FIELD.sub(field, text)
    return re.sub(r'\{[^}]*\}', '', text).replace('  ', ' ').strip()


def text_of(value, fallback=''):
    """Game text for a localized value, or fallback until the UI resolves it."""
    loc = _loc_value(value)
    if loc is None:
        return fallback
    try:
        key = int(loc.hash)
    except Exception:
        return fallback
    if not key:
        return fallback
    cached = _loc_cache().get(key)
    if cached is None:
        _loc_wanted.add(key)
        return fallback
    tokens = list(getattr(loc, 'tokens', []) or [])
    return _fill_tokens(cached, tokens) or fallback


def _pretty(tuning_name, prefixes=()):
    """'Buff_Energized_Coffee' -> 'Energized Coffee' when no game text yet."""
    name = tuning_name or ''
    for p in prefixes:
        if name.lower().startswith(p):
            name = name[len(p):]
    name = re.sub(r'([a-z])([A-Z])', r'\1 \2', name.replace('_', ' ').replace('-', ' '))
    return ' '.join(w.capitalize() if w.islower() else w for w in name.split())


# ---------------------------------------------------------------- state

_NEEDS = (('hunger', 'motive_hunger'), ('bladder', 'motive_bladder'), ('energy', 'motive_energy'),
          ('hygiene', 'motive_hygiene'), ('fun', 'motive_fun'), ('social', 'motive_social'))


def _need_level(p):
    if p < 15:
        return 'NEED_LEVEL_CRITICAL'
    if p < 30:
        return 'NEED_LEVEL_LOW'
    if p < 60:
        return 'NEED_LEVEL_OKAY'
    return 'NEED_LEVEL_GOOD'


def _needs(info):
    import tyjokr_a11y_plus_needs as needs
    out = []
    for key, stat in _NEEDS:
        p = needs.percent(info, stat)
        if p is not None:
            out.append({'key': key, 'percent': p, 'level': _need_level(p)})
    return out


def _mood_key(mood):
    name = getattr(mood, '__name__', '') or ''
    return re.sub(r'^mood_', '', name, flags=re.I).lower()


def _mood(info):
    buffs = getattr(info, 'Buffs', None)
    if buffs is None:
        return None
    mood = buffs.get_mood()
    if mood is None:
        return None
    try:
        intensity = int(buffs.get_mood_intensity())
    except Exception:
        intensity = 0
    name = ''
    names = getattr(mood, 'mood_names', None)
    if names:
        try:
            name = text_of(names[min(intensity, len(names) - 1)])
        except Exception:
            name = ''
    return {'key': _mood_key(mood), 'name': name, 'intensity': intensity}


def _moodlets(info):
    import services
    buffs = getattr(info, 'Buffs', None)
    if buffs is None:
        return []
    now = services.time_service().sim_now
    out = []
    for buff in buffs:
        if not getattr(buff, 'visible', False):
            continue
        # Buff instances share the Buff class; the tuning (name, mood, text)
        # is on buff_type.
        btype = getattr(buff, 'buff_type', None) or type(buff)
        tuning = getattr(btype, '__name__', '') or 'Buff'

        def tuned(attr):
            value = getattr(btype, attr, None)
            return value if value is not None else getattr(buff, attr, None)

        mood_type = tuned('mood_type')
        moodlet = {
            'key': tuning,
            'name': text_of(tuned('buff_name'), _pretty(tuning, ('buff_',))),
            'description': text_of(tuned('buff_description')),
            'moodKey': _mood_key(mood_type) if mood_type else '',
            'weight': int(tuned('mood_weight') or 0),
            'minutesLeft': -1,
        }
        try:
            timeout = buff.get_timeout_time()
            if timeout:
                moodlet['minutesLeft'] = max(0, int((timeout - now).in_minutes()))
        except Exception:
            pass
        out.append(moodlet)
    return out


def _actions(sim):
    import tyjokr_a11y_hub as hub
    out = []
    for info, interaction, state in hub._queue_entries(sim):
        fallback = ''
        if interaction is not None and hasattr(interaction, 'affordance'):
            fallback = _pretty(getattr(interaction.affordance, '__name__', ''), ('sim-', 'si_', 'mixer_'))
        name = text_of(getattr(info, 'display_name', None), fallback)
        if not name:
            continue
        out.append({'id': str(info.interaction_id), 'name': name,
                    'state': 'ACTION_STATE_RUNNING' if state == 'running' else 'ACTION_STATE_QUEUED',
                    'cancelable': bool(getattr(info, 'user_cancelable', False))})
    return out


def _speed():
    import services
    from clock import ClockSpeedMode
    speed = services.game_clock_service().clock_speed
    return {ClockSpeedMode.PAUSED: 'SPEED_PAUSED', ClockSpeedMode.SPEED2: 'SPEED_FAST',
            ClockSpeedMode.SPEED3: 'SPEED_ULTRA', ClockSpeedMode.SUPER_SPEED3: 'SPEED_SUPER'}.get(speed, 'SPEED_NORMAL')


def _clock():
    import services
    clock = {'speed': _speed()}
    try:
        import tyjokr_a11y_plus_time as t
        now = services.time_service().sim_now
        clock['day'] = t.day_text(now)
        clock['time'] = t.clock_text(now)
    except Exception:
        pass
    return clock


def _sim(info, client):
    sim = info.get_sim_instance()
    entry = {'id': str(info.sim_id), 'firstName': info.first_name or '', 'lastName': info.last_name or '',
             'active': client is not None and client.active_sim_info is info, 'here': sim is not None}
    for key, fn in (('mood', lambda: _mood(info)), ('needs', lambda: _needs(info)),
                    ('moodlets', lambda: _moodlets(info)), ('actions', lambda: _actions(sim) if sim else [])):
        if key != 'mood' and sim is None:
            continue
        try:
            value = fn()
            if value:
                entry[key] = value
        except Exception as ex:
            log('{} of {} unavailable: {!r}'.format(key, info.first_name, ex))
    return entry


def _narration():
    try:
        import tyjokr_a11y_settings as settings
        return {'macVoice': speech.speech_mode == 'hover', 'simActions': bool(settings.get('sim_actions'))}
    except Exception:
        return {}


def build_state():
    import services
    if not _in_game():
        return {'inGame': False, 'narration': _narration()}
    household = services.active_household()
    client = services.client_manager().get_first_client()
    return {
        'inGame': True,
        'clock': _clock(),
        'funds': str(int(household.funds.money)),
        'householdName': household.name or '',
        'sims': [_sim(info, client) for info in household.sim_info_gen()],
        'narration': _narration(),
    }


def _publish_state():
    global _last_state_json, _seq
    try:
        state = build_state()
    except Exception as ex:
        log('state failed: {!r}'.format(ex))
        return
    body = json.dumps(state, sort_keys=True, ensure_ascii=False)
    if body == _last_state_json:
        return
    _last_state_json = body
    _seq += 1
    state['seq'] = str(_seq)
    send({'state': state})


# ---------------------------------------------------------------- startup

if IS_MAC:
    try:
        _install_tick()
        start()
    except Exception as _ex:
        log('install failed: {!r}'.format(_ex))
