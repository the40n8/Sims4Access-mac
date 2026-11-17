import { Code, ConnectError, createClient, type Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CompanionService, type CommandOutcome } from './gen/ts4access/v1/companion_pb.js';
import { loadToken } from './session.js';

export const token = loadToken();

const auth: Interceptor = (next) => (req) => {
	if (token) req.header.set('Authorization', `Bearer ${token}`);
	return next(req);
};

export const client = createClient(
	CompanionService,
	createConnectTransport({ baseUrl: location.origin, interceptors: [auth] })
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
