import { Code, ConnectError, createClient, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanionService, type CommandOutcome } from './gen/simsight/v1/companion_pb.js';
import { loadToken } from './session.js';
import { hosted } from './meeting.js';

/** The pairing key; at a meeting point, the key of the Mac in use (useGame). */
export let token: string | null = hosted ? null : loadToken();
let base = location.origin;

/** At a meeting point: talk to this Mac, through its tunnel of the day. */
export function useGame(url: string, key: string): void {
	base = url.replace(/\/+$/, '');
	token = key;
}

// Requests are made to the page's own address, then sent to the Mac in use.
const toMac: typeof fetch = (input, init) => {
	if (base !== location.origin && typeof input === 'string' && input.startsWith(location.origin)) {
		return fetch(base + input.slice(location.origin.length), init);
	}
	return fetch(input, init);
};

const auth: Interceptor = (next) => (req) => {
	if (token) req.header.set('Authorization', `Bearer ${token}`);
	return next(req);
};

export const client = createClient(
	CompanionService,
	createConnectTransport({ baseUrl: location.origin, interceptors: [auth], fetch: toMac })
);

export function isUnauthenticated(err: unknown): boolean {
	return ConnectError.from(err).code === Code.Unauthenticated;
}

/** Runs a command and returns its outcome; transport errors become a failed outcome. */
export async function run(call: () => Promise<{ outcome?: CommandOutcome }>): Promise<CommandOutcome> {
	try {
		const res = await call();
		return res.outcome ?? ({ ok: false, reason: 'other' } as CommandOutcome);
	} catch {
		return { ok: false, reason: 'not_connected' } as CommandOutcome;
	}
}
