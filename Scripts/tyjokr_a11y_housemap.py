"""House map for the phone companion (macOS fork).

The house room by room: what each room is, how big, where it sits on the
floor, which rooms it opens onto, and what is in it. Built from the game's
own room ids (the same grid tyjokr_a11y_plus_rooms uses for Alt+R).
Runs on the game tick only (called from tyjokr_a11y_bridge).
"""
import math
import re
import sys

import tyjokr_a11y_plus_rooms as rooms

# Same convention as tyjokr_a11y_plus_rooms.bearing: north is +z.
COMPASS = ('N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW')

# Room kinds from tyjokr_a11y_plus_rooms.KINDS, as keys the phone translates.
_KIND_KEYS = {'Bathroom': 'BATHROOM', 'Kitchen': 'KITCHEN', 'Nursery': 'NURSERY', 'Bedroom': 'BEDROOM',
              'Dining room': 'DINING', 'Living room': 'LIVING', 'Study': 'STUDY'}
_DOOR = re.compile(r'(?<!Out)door', re.IGNORECASE)


def compass(dx, dz):
    angle = math.degrees(math.atan2(dx, dz)) % 360.0
    return COMPASS[int((angle + 22.5) % 360.0 // 45) % 8]


class _Room(object):
    """Cells of one room on one floor (also what plus_rooms.given_name expects)."""

    def __init__(self, rid, level):
        self.rid = rid
        self.level = level
        self.cells = []      # (i, j) grid indices
        self.tiles = []      # (x, z) world tile centres
        self.objects = []
        self.sims = []
        self.links = {}      # other room key or None (outside) -> {'door': bool, 'sides': {side: count}}
        self.kind = 'ROOM'
        self.number = 0
        self.custom = ''

    @property
    def key(self):
        return '{}:{}'.format(self.rid, self.level)

    @property
    def center(self):
        n = float(len(self.tiles)) or 1.0
        return (sum(t[0] for t in self.tiles) / n, sum(t[1] for t in self.tiles) / n)

    def link(self, other, side, door=False):
        entry = self.links.setdefault(other, {'door': False, 'sides': {}, 'doors': {}})
        if door:
            entry['door'] = True
            entry['doors'][side] = entry['doors'].get(side, 0) + 1
        else:
            entry['sides'][side] = entry['sides'].get(side, 0) + 1


def _is_door(obj):
    names = rooms.tag_names(obj)
    if any(_DOOR.search(n) for n in names):
        return True
    return bool(_DOOR.search(str(getattr(obj.definition, 'name', '') or '')))


def build():
    """The house map as a dict in the pipe's JSON (HouseMap), or None off a lot."""
    import services
    bridge = sys.modules['tyjokr_a11y_bridge']
    lot = services.active_lot()
    sim = services.get_active_sim()
    if lot is None:
        return None
    corners = list(lot.corners)
    c0, c1, c3 = corners[0], corners[1], corners[3]
    ux, uz = c1.x - c0.x, c1.z - c0.z
    vx, vz = c3.x - c0.x, c3.z - c0.z
    nu = max(1, int(round(math.hypot(ux, uz))))
    nv = max(1, int(round(math.hypot(vx, vz))))
    try:
        levels = range(int(lot.min_level), int(lot.max_level) + 1)
    except Exception:
        levels = range(0, 3)

    def world(i, j):
        fu, fv = (i + 0.5) / nu, (j + 0.5) / nv
        return (c0.x + ux * fu + vx * fv, c0.z + uz * fu + vz * fv)

    grid = {}
    found = {}
    for level in levels:
        for i in range(nu):
            for j in range(nv):
                x, z = world(i, j)
                rid = rooms.room_id_at(x, z, level)
                if not rid:
                    continue
                room = found.get((rid, level))
                if room is None:
                    room = found[(rid, level)] = _Room(rid, level)
                room.cells.append((i, j))
                room.tiles.append((x, z))
                grid[(level, i, j)] = room
    # A "room" that reaches the edge of the lot and covers much of it is the
    # yard, not a room.
    for key, room in list(found.items()):
        edge = any(i in (0, nu - 1) or j in (0, nv - 1) for (i, j) in room.cells)
        if edge and len(room.cells) > 0.25 * nu * nv:
            del found[key]
    if not found:
        return {'rooms': [], 'floors': []}

    def room_at(level, i, j):
        room = grid.get((level, i, j))
        return room if room is not None and (room.rid, room.level) in found else None

    def room_at_world(x, z, level):
        return found.get((rooms.room_id_at(x, z, level), level))

    # Walls and openings: neighbouring cells that belong to another room or outside.
    steps = ((1, 0, ux / nu, uz / nu), (-1, 0, -ux / nu, -uz / nu),
             (0, 1, vx / nv, vz / nv), (0, -1, -vx / nv, -vz / nv))
    for room in found.values():
        for (i, j) in room.cells:
            for (di, dj, dx, dz) in steps:
                other = room_at(room.level, i + di, j + dj)
                if other is not room:
                    room.link(other, compass(dx, dz))

    for obj in services.object_manager().get_valid_objects_gen():
        try:
            if getattr(obj, 'is_sim', False) or getattr(obj, 'is_terrain', False) or obj.is_hidden():
                continue
            on_lot = getattr(obj, 'is_on_active_lot', None)
            if on_lot is not None and not on_lot():
                continue
            in_inventory = getattr(obj, 'is_in_inventory', None)
            if in_inventory is not None and in_inventory():
                continue
            pos, level = obj.position, int(obj.level)
            if _is_door(obj):
                fwd = obj.forward
                a = room_at_world(pos.x + fwd.x * 0.6, pos.z + fwd.z * 0.6, level)
                b = room_at_world(pos.x - fwd.x * 0.6, pos.z - fwd.z * 0.6, level)
                for (here, there) in ((a, b), (b, a)):
                    if here is not None and here is not there:
                        cx, cz = here.center
                        here.link(there, compass(pos.x - cx, pos.z - cz), door=True)
                continue
            if not list(obj.super_affordances()):
                continue
            room = room_at_world(pos.x, pos.z, level)
            if room is not None:
                room.objects.append(obj)
        except Exception:
            continue
    for other in services.sim_info_manager().instanced_sims_gen():
        try:
            room = room_at_world(other.position.x, other.position.z, int(other.level))
            if room is not None:
                room.sims.append(other)
        except Exception:
            continue

    _name(list(found.values()))

    # Where each room sits on its floor, from the middle of the floor.
    middles = {}
    for room in found.values():
        middles.setdefault(room.level, []).extend(room.tiles)
    out = []
    for room in sorted(found.values(), key=lambda r: (r.level, r.kind, r.number, -len(r.cells))):
        cx, cz = room.center
        tiles = middles[room.level]
        mx = sum(t[0] for t in tiles) / len(tiles)
        mz = sum(t[1] for t in tiles) / len(tiles)
        span = max(max(abs(t[0] - mx) for t in tiles), max(abs(t[1] - mz) for t in tiles), 1.0)
        part = 'C' if math.hypot(cx - mx, cz - mz) < span * 0.3 else compass(cx - mx, cz - mz)
        iu = [c[0] for c in room.cells]
        jv = [c[1] for c in room.cells]
        entry = {'id': room.key, 'kind': room.kind, 'number': room.number, 'customName': room.custom,
                 'floor': room.level, 'width': max(iu) - min(iu) + 1, 'depth': max(jv) - min(jv) + 1,
                 'area': len(room.cells), 'part': part,
                 'sims': [s.sim_info.first_name for s in room.sims]}
        if sim is not None:
            entry['here'] = sim in room.sims
            if not entry['here']:
                dx, dz = cx - sim.position.x, cz - sim.position.z
                entry['distance'] = int(round(math.hypot(dx, dz)))
                entry['bearing'] = compass(dx, dz)
        links = []
        for other, info in room.links.items():
            sides = info['doors'] or info['sides']
            side = max(sides.items(), key=lambda kv: kv[1])[0]
            links.append({'roomId': other.key if other is not None else '', 'door': info['door'], 'side': side})
        links.sort(key=lambda l: (not l['door'], l['roomId'] == '', COMPASS.index(l['side'])))
        entry['links'] = links
        half = max(entry['width'], entry['depth']) / 2.0
        items = []
        for obj in room.objects:
            try:
                group, kind, keywords = bridge._classify(obj)
                dx, dz = obj.position.x - cx, obj.position.z - cz
                side = 'C' if math.hypot(dx, dz) < max(1.0, half * 0.4) else compass(dx, dz)
                target = {'id': str(obj.id), 'name': bridge._object_name(obj), 'group': group,
                          'kind': kind, 'keywords': keywords}
                if sim is not None:
                    target['distance'] = int(round(math.hypot(obj.position.x - sim.position.x,
                                                              obj.position.z - sim.position.z)))
                items.append({'target': target, 'side': side})
            except Exception:
                continue
        entry['items'] = items
        out.append(entry)
    return {'rooms': out, 'floors': sorted(middles)}


def _name(found):
    counts = {}
    for room in found:
        try:
            room.custom = rooms.given_name(room) or ''
        except Exception:
            room.custom = ''
        score = {}
        for obj in room.objects:
            names = rooms.tag_names(obj)
            for (kind, tags, weight) in rooms.KINDS:
                if names.intersection(tags):
                    score[kind] = score.get(kind, 0) + weight
        if score:
            room.kind = _KIND_KEYS.get(max(score.items(), key=lambda kv: kv[1])[0], 'ROOM')
        else:
            room.kind = 'ROOM' if room.objects else 'EMPTY'
        counts[room.kind] = counts.get(room.kind, 0) + 1
    numbered = {}
    for room in sorted(found, key=lambda r: (r.level, -len(r.cells))):
        if counts[room.kind] > 1:
            numbered[room.kind] = numbered.get(room.kind, 0) + 1
            room.number = numbered[room.kind]
