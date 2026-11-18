<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, simName } from '$lib/text.js';
	import { announce } from './Announcer.svelte';

	// Which Sim you play: one button per Sim on the lot (on Live and Play).
	const sims = $derived((live.state?.sims ?? []).filter((s) => s.here));
	let busy = $state(false);

	async function select(id: string) {
		if (busy) return;
		busy = true;
		const out = await run(() => client.selectSim({ simId: id }));
		busy = false;
		const sim = sims.find((s) => s.id === id);
		announce(
			out.ok && sim ? m.control_now({ name: simName(sim) }) : m.failed({ reason: reasonText(out.reason) })
		);
	}
</script>

{#if sims.length > 1}
	<section class="card stack" aria-labelledby="control">
		<h2 id="control">{m.control_sim_title()}</h2>
		<div class="pair">
			{#each sims as sim (sim.id)}
				<button type="button" aria-pressed={sim.active} aria-disabled={busy} onclick={() => select(sim.id)}>
					{simName(sim)}
				</button>
			{/each}
		</div>
	</section>
{/if}

<style>
	.pair {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
		gap: 0.5rem;
	}
</style>
