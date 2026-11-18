// Turning game data into sentences in the phone's language. Fixed sets
// (needs, levels, moods, speeds) use our translations; names that only the
// game knows (moodlets, actions) come in the game's language.
import { roomName } from './rooms.js';
import { m } from './i18n/paraglide/messages.js';
import {
	ActionState,
	EventKind,
	NeedLevel,
	Speed,
	type Event,
	type Mood,
	type Sim
} from './gen/ts4access/v1/game_pb.js';

const NEEDS: Record<string, () => string> = {
	hunger: m.need_hunger,
	bladder: m.need_bladder,
	energy: m.need_energy,
	hygiene: m.need_hygiene,
	fun: m.need_fun,
	social: m.need_social
};

const MOODS: Record<string, () => string> = {
	fine: m.mood_fine,
	happy: m.mood_happy,
	energized: m.mood_energized,
	playful: m.mood_playful,
	flirty: m.mood_flirty,
	focused: m.mood_focused,
	inspired: m.mood_inspired,
	confident: m.mood_confident,
	sad: m.mood_sad,
	angry: m.mood_angry,
	bored: m.mood_bored,
	embarrassed: m.mood_embarrassed,
	tense: m.mood_tense,
	uncomfortable: m.mood_uncomfortable,
	dazed: m.mood_dazed,
	scared: m.mood_scared,
	sleeping: m.mood_sleeping
};

export function needName(key: string): string {
	return NEEDS[key]?.() ?? key;
}

export function levelName(level: NeedLevel): string {
	switch (level) {
		case NeedLevel.CRITICAL:
			return m.level_critical();
		case NeedLevel.LOW:
			return m.level_low();
		case NeedLevel.OKAY:
			return m.level_okay();
		default:
			return m.level_good();
	}
}

/** A mood in the phone's language, else the game's own name. */
export function moodName(mood: Pick<Mood, 'key' | 'name'> | undefined): string {
	if (!mood) return '';
	return MOODS[mood.key]?.() ?? (mood.name || mood.key);
}

export function speedName(speed: Speed): string {
	switch (speed) {
		case Speed.PAUSED:
			return m.speed_paused();
		case Speed.FAST:
			return m.speed_fast();
		case Speed.ULTRA:
			return m.speed_ultra();
		case Speed.SUPER:
			return m.speed_super();
		default:
			return m.speed_normal();
	}
}

export function money(amount: bigint | number): string {
	return new Intl.NumberFormat(undefined).format(amount);
}

/** "2 h 15 min" style, in the phone's language. */
export function duration(minutes: number): string {
	const h = Math.floor(minutes / 60);
	const min = minutes % 60;
	const fmt = (value: number, unit: 'hour' | 'minute') =>
		new Intl.NumberFormat(undefined, { style: 'unit', unit, unitDisplay: 'short' }).format(value);
	return h > 0 ? (min > 0 ? `${fmt(h, 'hour')} ${fmt(min, 'minute')}` : fmt(h, 'hour')) : fmt(min, 'minute');
}

export function simName(sim: Sim): string {
	return sim.firstName || sim.lastName;
}

/** Where a Sim who is not on the lot is: work, school, or elsewhere. */
export function awaySentence(sim: Sim): string {
	const name = simName(sim);
	if (sim.awayAt === 'work')
		return sim.until ? m.summary_at_work({ name, time: sim.until }) : m.summary_at_work_now({ name });
	if (sim.awayAt === 'school')
		return sim.until ? m.summary_at_school({ name, time: sim.until }) : m.summary_at_school_now({ name });
	return m.summary_away({ name });
}

/** One sentence per Sim for the top of the Live screen. */
export function simSummary(sim: Sim): string {
	const name = simName(sim);
	if (!sim.here) return awaySentence(sim);
	const mood = moodName(sim.mood);
	const running = sim.actions.find((a) => a.state === ActionState.RUNNING) ?? sim.actions[0];
	let text: string = running
		? m.summary_doing({ name, mood, action: running.name })
		: m.summary_idle({ name, mood });
	if (sim.room) text += ' ' + m.summary_room({ room: roomName(sim.room) });
	const lowest = [...sim.needs].sort((a, b) => a.percent - b.percent)[0];
	if (lowest && lowest.percent < 60) {
		text += ' ' + m.summary_lowest({ need: needName(lowest.key), percent: lowest.percent });
	}
	return text;
}

/** Enum value from a proto enum name in an event ("NEED_LEVEL_LOW" -> NeedLevel.LOW). */
function enumValue<T extends Record<string, string | number>>(
	e: T,
	name: string,
	prefix: string
): T[keyof T] | undefined {
	const value = e[name.replace(prefix, '') as keyof T];
	return typeof value === 'number' ? value : undefined;
}

export function eventText(e: Event): string {
	const sim = e.simName;
	switch (e.kind) {
		case EventKind.MOOD:
			return m.ev_mood({ sim, mood: moodName({ key: e.subject, name: e.detail }) });
		case EventKind.MOODLET_ADDED:
			return m.ev_moodlet_added({ sim, moodlet: e.subject });
		case EventKind.MOODLET_REMOVED:
			return m.ev_moodlet_removed({ sim, moodlet: e.subject });
		case EventKind.NEED: {
			const level = enumValue(NeedLevel, e.detail, 'NEED_LEVEL_') ?? NeedLevel.LOW;
			return m.ev_need({ sim, need: needName(e.subject), level: levelName(level) });
		}
		case EventKind.ACTION_STARTED:
			return m.ev_action_started({ sim, action: e.subject });
		case EventKind.ACTION_QUEUED:
			return m.ev_action_queued({ sim, action: e.subject });
		case EventKind.ACTIVE_SIM:
			return m.ev_active_sim({ sim });
		case EventKind.SPEED:
			return m.ev_speed({ speed: speedName(enumValue(Speed, e.subject, 'SPEED_') ?? Speed.NORMAL) });
		case EventKind.FUNDS:
			return m.ev_funds({ amount: e.subject });
		case EventKind.SPOKEN:
			return m.ev_spoken({ text: e.subject });
		case EventKind.AWAY_STARTED:
			return e.subject === 'school' ? m.ev_left_school({ sim }) : m.ev_left_work({ sim });
		case EventKind.AWAY_ENDED:
			return e.subject === 'school' ? m.ev_back_school({ sim }) : m.ev_back_work({ sim });
		case EventKind.NOTIFICATION:
			return e.detail ? m.ev_notification({ title: e.subject, text: e.detail }) : e.subject;
		case EventKind.VISITOR:
			return m.ev_visitor({ name: e.subject });
		case EventKind.BILLS:
			return m.ev_bills({ amount: Number(e.subject).toLocaleString() });
		case EventKind.PROBLEM:
			return e.subject === 'FOOD' ? m.pb_food({ name: e.detail }) : m.ev_problem({ name: e.detail });
		case EventKind.DIALOG:
			return m.ev_dialog({ text: e.subject });
		default:
			return e.subject;
	}
}

export function reasonText(reason: string): string {
	const map: Record<string, () => string> = {
		not_connected: m.reason_not_connected,
		timeout: m.reason_timeout,
		sim_unavailable: m.reason_sim_unavailable,
		not_cancelable: m.reason_not_cancelable,
		action_finished: m.reason_action_finished,
		not_in_game: m.reason_not_in_game,
		menu_expired: m.reason_menu_expired,
		refused: m.reason_refused,
		target_unavailable: m.reason_target_unavailable,
		dialog_closed: m.reason_dialog_closed
	};
	return (map[reason] ?? m.reason_other)();
}
