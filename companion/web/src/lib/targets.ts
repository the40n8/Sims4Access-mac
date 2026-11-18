// Shared helpers for things a Sim can interact with (Play and House screens).
import { m } from './i18n/paraglide/messages.js';
import type { Target } from './gen/ts4access/v1/game_pb.js';

export function groupLabel(t: Target): string {
	if (t.isSelf) return m.group_SELF();
	const key = `group_${t.group}` as keyof typeof m;
	const fn = m[key] as (() => string) | undefined;
	return typeof fn === 'function' ? fn() : t.group;
}

// Things a player uses most first; lights, decorations and doors last.
export const GROUP_ORDER = [
	'SELF',
	'PHONE',
	'SIMS',
	'APPLIANCES',
	'PLUMBING',
	'COMFORT',
	'SURFACE',
	'ELECTRONICS',
	'ENTERTAINMENT',
	'KIDS',
	'STORAGE',
	'PETS',
	'VEHICLE',
	'MISC',
	'DECOR',
	'LIGHTING',
	'BUILD'
];

export function rank(t: Target): number {
	const i = GROUP_ORDER.indexOf(t.isSelf ? 'SELF' : t.group);
	return i < 0 ? GROUP_ORDER.length : i;
}

// Background things: listed apart so the useful ones come first.
export const isMinor = (t: Target) => ['DECOR', 'LIGHTING', 'BUILD'].includes(t.group);

/** "Refrigerators: Crisponix Ultragreat" */
export function objectName(t: Target): string {
	return t.kind ? `${t.kind}: ${t.name}` : t.name;
}

export function targetLabel(t: Target, owner = ''): string {
	if (t.isSelf) return m.target_self({ name: t.name });
	if (t.group === 'PHONE') return owner ? m.target_phone({ name: owner }) : m.group_PHONE();
	const name = objectName(t);
	return t.distance > 0 ? m.target_distance({ name, meters: t.distance }) : name;
}
