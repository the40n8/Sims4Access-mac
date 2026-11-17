<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, simName } from '$lib/text.js';
	import { Speed, type Interaction, type Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import type { CommandOutcome } from '$lib/gen/ts4access/v1/companion_pb.js';
	import { announce } from './Announcer.svelte';

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
	let menu = $state<Interaction[] | null>(null);
	let menuHeading: HTMLHeadingElement | undefined = $state();
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

	function groupLabel(t: Target): string {
		if (t.isSelf) return m.group_SELF();
		const key = `group_${t.group}` as keyof typeof m;
		const fn = m[key] as (() => string) | undefined;
		return typeof fn === 'function' ? fn() : t.group;
	}

	// Things a player uses most first; lights, decorations and doors last.
	const GROUP_ORDER = [
		'SELF',
		'SIMS',
		'APPLIANCES',
		'PLUMBING',
		'COMFORT',
		'SURFACE',
		'ELECTRONICS',
		'ENTERTAINMENT',
		'KIDS',
		'STORAGE',
		'PETS',
		'VEHICLE',
		'MISC',
		'DECOR',
		'LIGHTING',
		'BUILD'
	];
	const rank = (t: Target) => {
		const i = GROUP_ORDER.indexOf(t.isSelf ? 'SELF' : t.group);
		return i < 0 ? GROUP_ORDER.length : i;
	};

	function matches(t: Target, words: string[]): boolean {
		const hay = `${t.name} ${t.kind} ${t.keywords} ${groupLabel(t)}`.toLocaleLowerCase();
		return words.every((w) => hay.includes(w));
	}

	const groups = $derived.by(() => {
		const words = query.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean);
		const list = (targets ?? []).filter((t) => matches(t, words)).sort((a, b) => rank(a) - rank(b));
		const out: { label: string; items: Target[] }[] = [];
		for (const t of list) {
			const label = groupLabel(t);
			let g = out.find((x) => x.label === label);
			if (!g) out.push((g = { label, items: [] }));
			g.items.push(t);
		}
		return out;
	});

	async function open(t: Target) {
		selected = t;
		lastTargetId = t.id;
		menu = null;
		await tick();
		menuHeading?.focus();
		// Names come from the game's string tables: ask again while some are pending.
		for (let attempt = 0; attempt < 4; attempt++) {
			try {
				const res = await client.getInteractions({ targetId: t.id });
				if (!res.outcome?.ok) {
					menu = [];
					announce(m.failed({ reason: reasonText(res.outcome?.reason ?? '') }));
					return;
				}
				menu = res.interactions;
				if (!res.namesPending) return;
			} catch {
				menu = [];
				return;
			}
			await new Promise((r) => setTimeout(r, 700));
		}
	}

	async function back() {
		selected = null;
		menu = null;
		await tick();
		document.getElementById(`target-${lastTargetId}`)?.focus();
	}

	const menuGroups = $derived.by(() => {
		const out: { label: string; items: Interaction[] }[] = [];
		for (const i of menu ?? []) {
			const label = i.category;
			let g = out.find((x) => x.label === label);
			if (!g) out.push((g = { label, items: [] }));
			g.items.push(i);
		}
		// Top-level rows first, as in the game's pie menu.
		return out.sort((a, b) => (a.label === '' ? -1 : b.label === '' ? 1 : 0));
	});

	async function choose(i: Interaction) {
		if (!selected) return;
		if (!i.enabled) {
			announce(
				i.reason ? m.unavailable_because({ name: i.name, reason: i.reason }) : m.unavailable({ name: i.name })
			);
			return;
		}
		const target = selected;
		await act(
			() => run(() => client.runInteraction({ targetId: target.id, interactionId: i.id })),
			() => m.queued({ name: i.name })
		);
	}

	function targetLabel(t: Target): string {
		if (t.isSelf) return m.target_self({ name: t.name });
		const name = t.kind ? `${t.kind}: ${t.name}` : t.name;
		return t.distance > 0 ? m.target_distance({ name, meters: t.distance }) : name;
	}
</script>

{#if game?.inGame}
	{#if selected}
		<section class="card stack" aria-labelledby="menu-title">
			<button type="button" onclick={back}>{m.back_to_list()}</button>
			<h2 id="menu-title" tabindex="-1" bind:this={menuHeading}>
				{m.interactions_of({ name: selected.name })}
			</h2>
			{#if menu === null}
				<p class="muted" role="status">{m.loading()}</p>
			{:else if menu.length === 0}
				<p class="muted">{m.no_interactions()}</p>
			{:else}
				{#each menuGroups as g (g.label)}
					{#if g.label}<h3>{g.label}</h3>{/if}
					<ul class="stack">
						{#each g.items as i (i.id)}
							<li>
								<button
									type="button"
									class="row"
									aria-disabled={!i.enabled}
									disabled={busy}
									onclick={() => choose(i)}
								>
									{#if i.enabled}
										{i.name}
									{:else if i.reason}
										{m.unavailable_because({ name: i.name, reason: i.reason })}
									{:else}
										{m.unavailable({ name: i.name })}
									{/if}
								</button>
							</li>
						{/each}
					</ul>
				{/each}
				<p class="muted">{m.dialog_on_mac()}</p>
			{/if}
		</section>
	{:else}
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
	button[aria-disabled='true'] {
		opacity: 0.7;
		border-style: dashed;
	}
	h2:focus {
		outline: none;
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
