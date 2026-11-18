// Quick actions for the "Needs attention" card: each situation (a low need,
// a puddle, a broken object...) maps to the objects that help and to the
// game's own interactions to run, found by their names in the game's menu.
// Names come from the game in its language (English here).
import { m } from './i18n/paraglide/messages.js';
import { client } from './client.js';
import type { Interaction, Target } from './gen/ts4access/v1/game_pb.js';

export type Quick = {
	label: () => string;
	// Run on: the situation's own object, the nearest matching object, or the Sim's phone.
	on: 'object' | 'nearest' | 'phone';
	action: RegExp;
};

export type Help = {
	// Objects that help (for "nearest" and "Other ways").
	objects?: RegExp;
	quick: Quick[];
	// "Other ways" also offers the phone (order food, call someone...).
	phone?: boolean;
};

export const NEED_HELP: Record<string, Help> = {
	hunger: {
		objects: /fridge|refrigerator|stove|oven|microwave|grill|barbecue/i,
		quick: [
			{ label: m.qa_eat, on: 'nearest', action: /quick meal|^grab|^have |^eat/i },
			{ label: m.qa_cook, on: 'nearest', action: /^cook/i }
		],
		phone: true
	},
	bladder: { objects: /toilet/i, quick: [{ label: m.qa_toilet, on: 'nearest', action: /^use|pee/i }] },
	energy: {
		objects: /\bbeds?\b|mattress|sofa|couch/i,
		quick: [
			{ label: m.qa_sleep, on: 'nearest', action: /^sleep/i },
			{ label: m.qa_nap, on: 'nearest', action: /^nap/i }
		]
	},
	hygiene: {
		objects: /shower|bathtub|\btub\b|sink/i,
		quick: [
			{ label: m.qa_shower, on: 'nearest', action: /shower/i },
			{ label: m.qa_bath, on: 'nearest', action: /bath/i }
		]
	},
	fun: {
		objects: /\btv\b|television|computer|stereo|chess|easel|guitar|piano/i,
		quick: [{ label: m.qa_fun, on: 'nearest', action: /^watch|^play|^browse|^listen/i }]
	},
	social: {
		quick: [{ label: m.qa_call, on: 'phone', action: /call to chat|^call|^text/i }],
		phone: true
	}
};

export const PROBLEM_HELP: Record<string, Help> = {
	FIRE: {
		quick: [
			{ label: m.qa_extinguish, on: 'object', action: /extinguish|put out/i },
			{ label: m.qa_call_fire, on: 'phone', action: /fire|emergency/i }
		]
	},
	BROKEN: {
		quick: [
			{ label: m.qa_repair, on: 'object', action: /repair|fix/i },
			{ label: m.qa_hire_repair, on: 'phone', action: /repair|technician/i }
		]
	},
	PUDDLE: {
		quick: [
			{ label: m.qa_mop, on: 'object', action: /mop|clean/i },
			{ label: m.qa_hire_maid, on: 'phone', action: /maid|housekeeper/i }
		]
	},
	SPOILED: {
		quick: [{ label: m.qa_throw, on: 'object', action: /throw away|clean up|dispose|discard|put away/i }]
	},
	FOOD: {
		quick: [
			{
				label: m.qa_eat_this,
				on: 'object',
				action: /grab a plate|grab a serving|^eat|^have |^serve|^drink/i
			},
			{ label: m.qa_put_away, on: 'object', action: /put away/i }
		]
	}
};

export const VISITOR_HELP: Help = {
	quick: [
		{ label: m.qa_greet, on: 'object', action: /greet|introduce|wave|^chat/i },
		{ label: m.qa_welcome, on: 'object', action: /welcome|invite in/i }
	]
};

export const BILLS_HELP: Help = {
	objects: /mailbox/i,
	quick: [{ label: m.qa_pay, on: 'nearest', action: /pay bills|^pay/i }]
};

const hay = (t: Target) => `${t.kind} ${t.keywords} ${t.name}`;

/** Nearest objects that help, closest first. */
export function helpers(targets: Target[], objects: RegExp | undefined, limit = 5): Target[] {
	if (!objects) return [];
	return targets
		.filter((t) => !t.isSim && t.group !== 'PHONE' && objects.test(hay(t)))
		.sort((a, b) => a.distance - b.distance)
		.slice(0, limit);
}

/** The game's menu for a target, waiting a moment for names still on their way. */
async function menuOf(targetId: string): Promise<Interaction[] | null> {
	for (let attempt = 0; attempt < 4; attempt++) {
		const res = await client.getInteractions({ targetId });
		if (!res.outcome?.ok) return null;
		if (!res.namesPending || attempt === 3) return res.interactions;
		await new Promise((r) => setTimeout(r, 600));
	}
	return null;
}

export type QuickResult =
	| { kind: 'queued'; name: string }
	| { kind: 'unavailable'; name: string; reason: string }
	| { kind: 'missing' }
	| { kind: 'failed'; reason: string };

/** Find the interaction in the target's menu and queue it. */
export async function runQuick(targetId: string, action: RegExp): Promise<QuickResult> {
	try {
		const menu = await menuOf(targetId);
		if (menu === null) return { kind: 'failed', reason: 'target_unavailable' };
		const rows = menu.filter((i) => action.test(i.name));
		const row = rows.find((i) => i.enabled);
		if (!row) {
			return rows.length
				? { kind: 'unavailable', name: rows[0].name, reason: rows[0].reason }
				: { kind: 'missing' };
		}
		const res = await client.runInteraction({ targetId, interactionId: row.id });
		return res.outcome?.ok
			? { kind: 'queued', name: row.name }
			: { kind: 'failed', reason: res.outcome?.reason ?? '' };
	} catch {
		return { kind: 'failed', reason: 'not_connected' };
	}
}
