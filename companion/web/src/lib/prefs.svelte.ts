// Per-phone preferences: which events VoiceOver announces automatically.
import { EventKind, type Event } from './gen/ts4access/v1/game_pb.js';

export type Group = 'messages' | 'mood' | 'moodlet' | 'need' | 'action' | 'game';

export const GROUPS: readonly Group[] = ['messages', 'mood', 'moodlet', 'need', 'action', 'game'];

// How much of the Sims' social life to follow.
export type SocialLevel = 'off' | 'labels' | 'normal' | 'all';
export const SOCIAL_LEVELS: readonly SocialLevel[] = ['off', 'labels', 'normal', 'all'];

/** Whether a social event passes the chosen level (other events always do). */
export function socialShown(e: Event, level: SocialLevel = prefs.social): boolean {
	if (e.kind === EventKind.RELATIONSHIP) return level !== 'off';
	if (e.kind !== EventKind.SOCIAL) return true;
	if (level === 'all') return true;
	if (level !== 'normal') return false;
	// Small drifts (time passing) are left out; real interactions move 2+.
	return Math.abs(e.friendshipDelta) >= 2 || Math.abs(e.romanceDelta) >= 2;
}

export function groupOf(kind: EventKind): Group | 'social' | 'spoken' | null {
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
		case EventKind.AWAY_STARTED:
		case EventKind.AWAY_ENDED:
		case EventKind.SPEED:
		case EventKind.FUNDS:
			return 'game';
		case EventKind.NOTIFICATION:
		case EventKind.DIALOG:
		case EventKind.PROBLEM:
		case EventKind.VISITOR:
		case EventKind.BILLS:
		case EventKind.ROUTE_FAIL:
			return 'messages';
		case EventKind.SOCIAL:
		case EventKind.RELATIONSHIP:
			return 'social';
		case EventKind.SPOKEN:
			return 'spoken';
		default:
			return null;
	}
}

type Stored = {
	announce: Record<Group, boolean>;
	showSpoken: boolean;
	autoPause: boolean;
	social: SocialLevel;
};

const KEY = 'ts4access.prefs';
const DEFAULTS: Stored = {
	// Actions start often; off by default so announcements stay meaningful.
	announce: { messages: true, mood: true, moodlet: true, need: true, action: false, game: true },
	showSpoken: false,
	// Pause while choosing an action, resume once it is queued (off: a choice).
	autoPause: false,
	social: 'normal'
};

function load(): Stored {
	try {
		const raw = localStorage.getItem(KEY);
		if (raw) {
			const saved = JSON.parse(raw) as Partial<Stored>;
			return {
				announce: { ...DEFAULTS.announce, ...saved.announce },
				showSpoken: saved.showSpoken ?? DEFAULTS.showSpoken,
				autoPause: saved.autoPause ?? DEFAULTS.autoPause,
				social: saved.social ?? DEFAULTS.social
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
