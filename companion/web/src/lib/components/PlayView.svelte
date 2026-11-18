<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, simName } from '$lib/text.js';
	import { Speed, type Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import type { CommandOutcome } from '$lib/gen/ts4access/v1/companion_pb.js';
	import { groupLabel, rank, targetLabel } from '$lib/targets.js';
	import { matchesWords } from '$lib/search.js';
	import { announce } from './Announcer.svelte';
	import InteractionMenu from './InteractionMenu.svelte';
	import VoiceToggle from './VoiceToggle.svelte';

	const game = $derived(live.state);
	// Buttons say what they do ("Play"); the status line says the state ("Normal speed").
	const speeds: { speed: Speed; label: () => string }[] = [
		{ speed: Speed.PAUSED, label: m.btn_pause },
		{ speed: Speed.NORMAL, label: m.btn_play },
		{ speed: Speed.FAST, label: m.btn_fast },
		{ speed: Speed.ULTRA, label: m.btn_ultra }
	];

	let busy = $state(false);
	let targets = $state<Target[] | null>(null);
	let query = $state('');
	let selected = $state<Target | null>(null);
	let lastTargetId = '';

	async function act(call: () => Promise<CommandOutcome>, success = m.done) {
		busy = true;
		const out = await call();
		busy = false;
		announce(out.ok ? success() : m.failed({ reason: reasonText(out.reason) }));
		return out.ok;
	}
	const setSpeed = (speed: Speed) => act(() => run(() => client.setSpeed({ speed })));
	const selectSim = (simId: string) => act(() => run(() => client.selectSim({ simId })));

	async function loadTargets() {
		targets = null;
		try {
			const res = await client.listTargets({});
			targets = res.outcome?.ok ? res.targets : [];
		} catch {
			targets = [];
		}
	}

	// Load the list when this screen opens and when the controlled Sim changes.
	let loadedFor = '';
	$effect(() => {
		const active = game?.sims.find((s) => s.active)?.id ?? '';
		if (game?.inGame && active !== loadedFor) {
			loadedFor = active;
			void loadTargets();
		}
	});

	const matches = (t: Target) => matchesWords(`${t.name} ${t.kind} ${t.keywords} ${groupLabel(t)}`, query);

	const groups = $derived.by(() => {
		const list = (targets ?? []).filter(matches).sort((a, b) => rank(a) - rank(b));
		const out: { label: string; items: Target[] }[] = [];
		for (const t of list) {
			const label = groupLabel(t);
			let g = out.find((x) => x.label === label);
			if (!g) out.push((g = { label, items: [] }));
			g.items.push(t);
		}
		return out;
	});

	function open(t: Target) {
		selected = t;
		lastTargetId = t.id;
	}

	async function back() {
		selected = null;
		await tick();
		document.getElementById(`target-${lastTargetId}`)?.focus();
	}
</script>

{#if game?.inGame}
	{#if selected}
		<InteractionMenu target={selected} onBack={back} />
	{:else}
		<section class="card"><VoiceToggle /></section>

		<section class="card stack" aria-labelledby="speed">
			<h2 id="speed">{m.speed_title()}</h2>
			<div class="grid" role="group" aria-labelledby="speed">
				{#each speeds as { speed, label } (speed)}
					<button
						type="button"
						aria-pressed={game.clock?.speed === speed}
						disabled={busy}
						onclick={() => setSpeed(speed)}>{label()}</button
					>
				{/each}
			</div>
		</section>

		{#if game.sims.filter((s) => s.here).length > 1}
			<section class="card stack" aria-labelledby="control">
				<h2 id="control">{m.control_sim_title()}</h2>
				<div class="grid">
					{#each game.sims.filter((s) => s.here) as sim (sim.id)}
						<button type="button" aria-pressed={sim.active} disabled={busy} onclick={() => selectSim(sim.id)}>
							{m.control_sim({ name: simName(sim) })}
						</button>
					{/each}
				</div>
			</section>
		{/if}

		<section class="card stack" aria-labelledby="interact">
			<h2 id="interact">{m.interact_title()}</h2>
			<label class="search">
				<span class="visually-hidden">{m.search()}</span>
				<input type="search" placeholder={m.search()} bind:value={query} autocomplete="off" />
			</label>
			<button type="button" onclick={loadTargets}>{m.refresh()}</button>
			{#if targets === null}
				<p class="muted" role="status">{m.loading()}</p>
			{:else if groups.length === 0}
				<p class="muted">{m.nothing_found()}</p>
			{:else}
				{#each groups as g (g.label)}
					<h3>{g.label}</h3>
					<ul class="stack">
						{#each g.items as t (t.id)}
							<li>
								<button type="button" class="row" id={`target-${t.id}`} onclick={() => open(t)}>
									{targetLabel(t)}
								</button>
							</li>
						{/each}
					</ul>
				{/each}
			{/if}
		</section>
	{/if}
{/if}

<style>
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
		gap: 0.5rem;
	}
	.row {
		width: 100%;
		text-align: start;
	}
	.search input {
		font: inherit;
		width: 100%;
		min-height: var(--tap);
		border-radius: var(--r);
		border: 2px solid var(--line);
		background: var(--surface-2);
		color: var(--text);
		padding: 0 0.75rem;
	}
</style>
