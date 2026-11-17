// Per-phone preferences: which events VoiceOver announces automatically.
import { EventKind } from './gen/ts4access/v1/game_pb.js';

export type Group = 'mood' | 'moodlet' | 'need' | 'action' | 'game';

export const GROUPS: readonly Group[] = ['mood', 'moodlet', 'need', 'action', 'game'];

export function groupOf(kind: EventKind): Group | 'spoken' | null {
	switch (kind) {
		case EventKind.MOOD:
			return 'mood';
		case EventKind.MOODLET_ADDED:
		case EventKind.MOODLET_REMOVED:
			return 'moodlet';
		case EventKind.NEED:
			return 'need';
		case EventKind.ACTION_STARTED:
		case EventKind.ACTION_QUEUED:
			return 'action';
		case EventKind.ACTIVE_SIM:
		case EventKind.SPEED:
		case EventKind.FUNDS:
			return 'game';
		case EventKind.SPOKEN:
			return 'spoken';
		default:
			return null;
	}
}

type Stored = { announce: Record<Group, boolean>; showSpoken: boolean };

const KEY = 'ts4access.prefs';
const DEFAULTS: Stored = {
	// Actions start often; off by default so announcements stay meaningful.
	announce: { mood: true, moodlet: true, need: true, action: false, game: true },
	showSpoken: false
};

function load(): Stored {
	try {
		const raw = localStorage.getItem(KEY);
		if (raw) {
			const saved = JSON.parse(raw) as Partial<Stored>;
			return {
				announce: { ...DEFAULTS.announce, ...saved.announce },
				showSpoken: saved.showSpoken ?? DEFAULTS.showSpoken
			};
		}
	} catch {
		// ignore unreadable preferences
	}
	return structuredClone(DEFAULTS);
}

export const prefs = $state(load());

export function savePrefs(): void {
	try {
		localStorage.setItem(KEY, JSON.stringify(prefs));
	} catch {
		// private browsing: preferences last for this page only
	}
}
