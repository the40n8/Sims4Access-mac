// The app at a meeting point (companion/meeting-point): one fixed address
// for the phone, which finds each paired Mac wherever its tunnel is today.
// Served by the Mac itself instead, none of this is used.
import { refreshManifest } from './session.js';

/** True when this app is served by a meeting point (it marks the page). */
export const hosted =
	typeof document !== 'undefined' && !!document.querySelector('meta[name="simsight-meeting-point"]');

export type Game = { id: string; token: string; name: string };
export type Found = { game: Game; url: string };

const GAMES = 'simsight.games';
const KEY = /^[A-Za-z0-9_-]{20,64}$/;

export function games(): Game[] {
	try {
		const list = JSON.parse(localStorage.getItem(GAMES) ?? '[]') as Game[];
		return Array.isArray(list) ? list.filter((g) => KEY.test(g.id) && KEY.test(g.token)) : [];
	} catch {
		return [];
	}
}

function store(list: Game[]): void {
	try {
		localStorage.setItem(GAMES, JSON.stringify(list));
	} catch {
		// private browsing: paired for this page only
	}
	// For "Add to Home Screen" (its own storage on iOS): the meeting point
	// puts these in the install's start address (see its manifest).
	try {
		const value = list.map((g) => `${g.id}.${g.token}`).join('~');
		const secure = location.protocol === 'https:' ? '; Secure' : '';
		document.cookie = value
			? `ssg=${value}; Path=/; Max-Age=31536000; SameSite=Strict${secure}`
			: 'ssg=; Path=/; Max-Age=0';
		refreshManifest();
	} catch {
		// cookies off
	}
}

export function saveGame(game: Game): void {
	const list = games().filter((g) => g.id !== game.id);
	const old = games().find((g) => g.id === game.id);
	store([{ ...game, name: game.name || old?.name || '' }, ...list]);
}

export function forgetGame(id: string): void {
	store(games().filter((g) => g.id !== id));
}

/** Takes the Mac and key from a link (#m=…&k=…, or #g=id.key~… from the Home Screen install). */
export function takeLink(text: string): boolean {
	const m = /[#&?]m=([A-Za-z0-9_-]{20,64})&k=([A-Za-z0-9_-]{20,64})/.exec(text);
	if (m) {
		saveGame({ id: m[1], token: m[2], name: '' });
		return true;
	}
	const g = /[#&]g=([A-Za-z0-9_.~-]+)/.exec(text);
	if (!g) return false;
	let any = false;
	for (const part of g[1].split('~')) {
		const [id, token] = part.split('.');
		if (KEY.test(id ?? '') && KEY.test(token ?? '')) {
			if (!games().some((x) => x.id === id)) saveGame({ id, token, name: '' });
			any = true;
		}
	}
	return any;
}

/** True while a pairing link is still in the address bar. */
export function linkInAddressBar(): boolean {
	return /(?:^#|&)(m|g)=/.test(location.hash);
}

type Where = { id: string; name: string; url: string; online: boolean; seen: number };

async function where(id: string): Promise<Where | null> {
	try {
		const res = await fetch(`/mp/where?id=${encodeURIComponent(id)}`, { cache: 'no-store' });
		return res.ok ? ((await res.json()) as Where) : null;
	} catch {
		return null;
	}
}

/** The paired Mac whose game is running (the most recently seen when several are). */
export async function findGame(): Promise<Found | null> {
	const list = games();
	const answers = await Promise.all(list.map((g) => where(g.id)));
	let best: { game: Game; w: Where } | null = null;
	answers.forEach((w, i) => {
		if (!w) return;
		const game = list[i];
		if (w.name && w.name !== game.name) saveGame({ ...game, name: w.name });
		if (w.online && w.url && (!best || w.seen > best.w.seen)) best = { game: { ...game, name: w.name }, w };
	});
	const found = best as { game: Game; w: Where } | null;
	return found ? { game: found.game, url: found.w.url } : null;
}

/** Pairs with the code the Mac says: the meeting point knows which Mac shows it. */
export async function pairWithCode(code: string): Promise<'ok' | 'wrong' | 'locked' | 'offline'> {
	try {
		const res = await fetch('/mp/code', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ code })
		});
		if (res.status === 429) return 'locked';
		if (res.status === 404) return 'wrong';
		if (!res.ok) return 'offline';
		const mac = (await res.json()) as Where;
		if (!mac.url) return 'offline';
		const pair = await fetch(`${mac.url}/api/pair-code`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ code })
		});
		if (pair.status === 429) return 'locked';
		if (!pair.ok) return 'wrong';
		const { token } = (await pair.json()) as { token?: string };
		if (!token || !KEY.test(token)) return 'wrong';
		saveGame({ id: mac.id, token, name: mac.name });
		return 'ok';
	} catch {
		return 'offline';
	}
}

// A pairing link opened here (from the Mac's QR code, or the Home Screen install).
if (hosted) takeLink(location.hash);
