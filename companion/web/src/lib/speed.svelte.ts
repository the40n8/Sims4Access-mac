// Game speed from anywhere in the app: the top bar's Pause/Resume toggle and
// the optional pause while choosing an action.
import { client, run } from './client.js';
import { live } from './live.svelte.js';
import { prefs } from './prefs.svelte.js';
import { Speed } from './gen/ts4access/v1/game_pb.js';
import type { CommandOutcome } from './gen/ts4access/v1/companion_pb.js';

let resumeTo = Speed.NORMAL;

export const currentSpeed = () => live.state?.clock?.speed ?? Speed.UNSPECIFIED;
export const isPaused = () => currentSpeed() === Speed.PAUSED;

export function setSpeed(speed: Speed): Promise<CommandOutcome> {
	const now = currentSpeed();
	if (speed === Speed.PAUSED && now !== Speed.PAUSED && now !== Speed.UNSPECIFIED) resumeTo = now;
	return run(() => client.setSpeed({ speed }));
}

/** Pause, or resume at the speed the game had before the pause. */
export function togglePause(): Promise<CommandOutcome> {
	return isPaused() ? setSpeed(resumeTo) : setSpeed(Speed.PAUSED);
}

// Auto-pause while a menu is open: only resumes what it paused itself.
let pausedForMenu = false;

export async function menuOpened(): Promise<boolean> {
	if (!prefs.autoPause || isPaused() || currentSpeed() === Speed.UNSPECIFIED) return false;
	const out = await setSpeed(Speed.PAUSED);
	pausedForMenu = out.ok;
	return out.ok;
}

export async function menuDone(): Promise<boolean> {
	if (!pausedForMenu) return false;
	pausedForMenu = false;
	if (!isPaused()) return false; // someone already resumed
	const out = await setSpeed(resumeTo);
	return out.ok;
}
