<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, simName, speedName } from '$lib/text.js';
	import { Speed, type Narration } from '$lib/gen/ts4access/v1/game_pb.js';
	import type { CommandOutcome } from '$lib/gen/ts4access/v1/companion_pb.js';
	import { announce } from './Announcer.svelte';

	const game = $derived(live.state);
	const speeds = [Speed.PAUSED, Speed.NORMAL, Speed.FAST, Speed.ULTRA];
	let busy = $state(false);

	async function act(call: () => Promise<CommandOutcome>) {
		busy = true;
		const out = await call();
		busy = false;
		announce(out.ok ? m.done() : m.failed({ reason: reasonText(out.reason) }));
	}

	const setSpeed = (speed: Speed) => act(() => run(() => client.setSpeed({ speed })));
	const selectSim = (simId: string) => act(() => run(() => client.selectSim({ simId })));
	function setNarration(change: Partial<Narration>) {
		const current = game?.narration ?? { macVoice: true, simActions: true };
		const narration = { macVoice: current.macVoice, simActions: current.simActions, ...change };
		return act(() => run(() => client.setNarration({ narration })));
	}
</script>

{#if game?.inGame}
	<section class="card stack" aria-labelledby="speed">
		<h2 id="speed">{m.speed_title()}</h2>
		<div class="row" role="group" aria-labelledby="speed">
			{#each speeds as speed (speed)}
				<button
					type="button"
					aria-pressed={game.clock?.speed === speed}
					disabled={busy}
					onclick={() => setSpeed(speed)}>{speedName(speed)}</button
				>
			{/each}
		</div>
	</section>

	<section class="card stack" aria-labelledby="control">
		<h2 id="control">{m.control_sim_title()}</h2>
		<div class="row">
			{#each game.sims.filter((s) => s.here) as sim (sim.id)}
				<button type="button" aria-pressed={sim.active} disabled={busy} onclick={() => selectSim(sim.id)}>
					{m.control_sim({ name: simName(sim) })}
				</button>
			{/each}
		</div>
	</section>
{/if}

<section class="card stack" aria-labelledby="narration">
	<h2 id="narration">{m.narration_title()}</h2>
	<button
		type="button"
		aria-pressed={game?.narration?.macVoice ?? false}
		disabled={busy || !live.gameConnected}
		onclick={() => setNarration({ macVoice: !game?.narration?.macVoice })}>{m.mac_voice()}</button
	>
	<button
		type="button"
		aria-pressed={game?.narration?.simActions ?? false}
		disabled={busy || !live.gameConnected}
		onclick={() => setNarration({ simActions: !game?.narration?.simActions })}>{m.sim_actions()}</button
	>
</section>

<style>
	.row {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
		gap: 0.5rem;
	}
</style>
