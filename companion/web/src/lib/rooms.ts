// Room, floor and direction words shared by the House tab and the rest.
import { m } from './i18n/paraglide/messages.js';

function pick(key: string, fallback: string): string {
	const fn = m[key as keyof typeof m] as (() => string) | undefined;
	return typeof fn === 'function' ? fn() : fallback;
}

export const dir = (d: string) => pick(`dir_${d}`, d);
export const part = (p: string) => (p === 'C' ? m.part_C() : m.part_dir({ dir: dir(p) }));
export const side = (p: string) => (p === 'C' ? m.side_C() : m.side_dir({ dir: dir(p) }));
export const kindName = (kind: string) => pick(`room_${kind}`, m.room_ROOM());

type Named = { kind: string; number: number; customName: string; kind2?: string };

/** "Kitchen", "Bedroom 2", "Kitchen and Dining room", or the player's own name. */
export function roomName(r: Named): string {
	if (r.customName) return r.customName;
	let base = kindName(r.kind);
	if (r.kind2) base = m.room_and({ a: base, b: kindName(r.kind2) });
	return r.number ? `${base} ${r.number}` : base;
}

export function floorName(f: number): string {
	if (f === 0) return m.floor_ground();
	if (f === -1) return m.floor_basement();
	if (f < 0) return m.floor_basement_n({ n: -f });
	return m.floor_up({ n: f });
}
