<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, simName } from '$lib/text.js';
	import { Speed, type Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import { CameraMode, type CommandOutcome } from '$lib/gen/ts4access/v1/companion_pb.js';
	import { groupLabel, rank, targetLabel } from '$lib/targets.js';
	import { matchesWords } from '$lib/search.js';
	import { announce } from './Announcer.svelte';
	import InteractionMenu from './InteractionMenu.svelte';
	import ActionQueue from './ActionQueue.svelte';
	import ControlSim from './ControlSim.svelte';
	import InventoryView from './InventoryView.svelte';
	import { setSpeed } from '$lib/speed.svelte.js';

	const game = $derived(live.state);
	const me = $derived(game?.sims.find((x) => x.active));
	// Buttons say what they do ("Play"); the status line says the state ("Normal speed").
	const speeds: { speed: Speed; label: () => string }[] = [
		{ speed: Speed.PAUSED, label: m.btn_pause },
		{ speed: Speed.NORMAL, label: m.btn_play },
		{ speed: Speed.FAST, label: m.btn_fast },
		{ speed: Speed.ULTRA, label: m.btn_ultra }
	];

	let busy = $state(false);
	let showInventory = $state(false);
	async function closeInventory() {
		showInventory = false;
		await tick();
		document.getElementById('open-inventory')?.focus();
	}
	let targets = $state<Target[] | null>(null);
	let query = $state('');
	let selected = $state<Target | null>(null);
	let lastTargetId = '';

	async function act(call: () => Promise<CommandOutcome>, success = m.done) {
		if (busy) return false;
		busy = true;
		const out = await call();
		busy = false;
		announce(out.ok ? success() : m.failed({ reason: reasonText(out.reason) }));
		return out.ok;
	}
	const changeSpeed = (speed: Speed) => act(() => setSpeed(speed));
	const camera = (mode: CameraMode) => me && act(() => run(() => client.setCamera({ simId: me.id, mode })));

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

	// The Sim itself and its phone are shortcuts at the top; other Sims and
	// objects are searched. Objects are grouped by catalog category.
	const self = $derived((targets ?? []).find((t) => t.isSelf));
	const phone = $derived((targets ?? []).find((t) => t.group === 'PHONE'));
	const sims = $derived((targets ?? []).filter((t) => t.isSim && !t.isSelf && matches(t)));
	const groups = $derived.by(() => {
		const list = (targets ?? [])
			.filter((t) => !t.isSim && t.group !== 'PHONE' && matches(t))
			.sort((a, b) => rank(a) - rank(b));
		const out: { key: string; label: string; items: Target[] }[] = [];
		for (const t of list) {
			let g = out.find((x) => x.key === t.group);
			if (!g) out.push((g = { key: t.group, label: groupLabel(t), items: [] }));
			g.items.push(t);
		}
		return out;
	});
	const found = $derived(sims.length + groups.reduce((n, g) => n + g.items.length, 0));
	const countText = $derived(targets === null ? '' : m.found_count({ count: found }));

	// Categories start folded; a search unfolds everything it finds.
	let opened = $state<string[]>([]);
	const isOpen = (key: string) => query.trim() !== '' || opened.includes(key);
	function toggleGroup(key: string) {
		opened = opened.includes(key) ? opened.filter((k) => k !== key) : [...opened, key];
	}

	// Enter on the keyboard: close the keyboard and go to the results.
	let searchBox: HTMLInputElement | undefined = $state();
	let resultsHeading: HTMLHeadingElement | undefined = $state();
	async function submitSearch(e: SubmitEvent) {
		e.preventDefault();
		searchBox?.blur();
		announce(countText);
		await tick();
		resultsHeading?.focus();
	}
	// While typing, say how many things match (after a short pause).
	let countTimer: ReturnType<typeof setTimeout> | undefined;
	$effect(() => {
		const q = query;
		const text = countText;
		clearTimeout(countTimer);
		if (q.trim()) countTimer = setTimeout(() => announce(text), 900);
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
	{:else if showInventory && me}
		<InventoryView owner={simName(me)} onBack={closeInventory} />
	{:else}
		<ControlSim />

		{#if me && me.here}
			<section class="card stack" aria-labelledby="camera">
				<h2 id="camera">{m.camera_title()}</h2>
				<p class="muted">{m.camera_help()}</p>
				<div class="pair">
					<button
						type="button"
						aria-pressed={game.cameraFollows === me.id}
						aria-disabled={busy}
						onclick={() => camera(CameraMode.FOLLOW)}>{m.camera_follow({ name: simName(me) })}</button
					>
					<button type="button" aria-disabled={busy} onclick={() => camera(CameraMode.CLOSE)}
						>{m.camera_close({ name: simName(me) })}</button
					>
				</div>
				{#if game.cameraFollows}
					<button type="button" aria-disabled={busy} onclick={() => camera(CameraMode.FREE)}
						>{m.camera_free()}</button
					>
				{/if}
			</section>
		{/if}

		{#if me}<ActionQueue sim={me} />{/if}

		{#if self || phone}
			<section class="card stack" aria-labelledby="you">
				<h2 id="you">{m.you_title()}</h2>
				<div class="pair">
					{#if self}
						<button type="button" id={`target-${self.id}`} onclick={() => open(self)}>
							{m.target_self({ name: self.name })}
						</button>
					{/if}
					{#if phone}
						<button type="button" id={`target-${phone.id}`} onclick={() => open(phone)}>
							{targetLabel(phone, me ? simName(me) : '')}
						</button>
					{/if}
					{#if me}
						<button type="button" id="open-inventory" onclick={() => (showInventory = true)}>
							{m.inventory_of({ name: simName(me) })}
						</button>
					{/if}
				</div>
			</section>
		{/if}

		<section class="card stack" aria-labelledby="interact">
			<h2 id="interact">{m.interact_title()}</h2>
			<form role="search" class="stack" onsubmit={submitSearch}>
				<label class="search stack">
					<span>{m.search()}</span>
					<input
						type="search"
						bind:this={searchBox}
						bind:value={query}
						enterkeyhint="search"
						autocomplete="off"
						autocorrect="off"
						aria-describedby="search-count"
					/>
				</label>
			</form>
			<p id="search-count" class="muted">{countText}</p>
			{#if targets === null}
				<p class="muted" role="status">{m.loading()}</p>
			{:else if found === 0}
				<p class="muted">{m.nothing_found()}</p>
			{:else}
				<h3 class="visually-hidden" tabindex="-1" bind:this={resultsHeading}>{m.results_title()}</h3>
				{#if sims.length}
					<h3>{m.sims_nearby()}</h3>
					<ul class="stack items">
						{#each sims as t (t.id)}
							<li>
								<button type="button" class="row" id={`target-${t.id}`} onclick={() => open(t)}>
									{targetLabel(t)}
								</button>
							</li>
						{/each}
					</ul>
				{/if}
				{#each groups as g (g.key)}
					{@const expanded = isOpen(g.key)}
					<h3 class="group">
						<button
							type="button"
							class="group-toggle"
							aria-expanded={expanded}
							aria-controls={`group-${g.key}`}
							onclick={() => toggleGroup(g.key)}
						>
							<svg
								class="chevron"
								class:open={expanded}
								viewBox="0 0 16 16"
								aria-hidden="true"
								focusable="false"><path d="M6 3l5 5-5 5z" /></svg
							>
							<span>{m.group_count({ group: g.label, count: g.items.length })}</span>
						</button>
					</h3>
					{#if expanded}
						<ul class="stack items" id={`group-${g.key}`}>
							{#each g.items as t (t.id)}
								<li>
									<button type="button" class="row" id={`target-${t.id}`} onclick={() => open(t)}>
										{targetLabel(t)}
									</button>
								</li>
							{/each}
						</ul>
					{/if}
				{/each}
			{/if}
			<button type="button" onclick={loadTargets}>{m.refresh()}</button>
		</section>

		<section class="card stack" aria-labelledby="speed">
			<h2 id="speed">{m.speed_title()}</h2>
			<div class="speeds" role="group" aria-labelledby="speed">
				{#each speeds as { speed, label } (speed)}
					<button
						type="button"
						aria-pressed={game.clock?.speed === speed}
						aria-disabled={busy}
						onclick={() => changeSpeed(speed)}>{label()}</button
					>
				{/each}
			</div>
		</section>
	{/if}
{/if}

<style>
	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;
	}
	.speeds {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 0.35rem;
	}
	.speeds button {
		padding-inline: 0.25rem;
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
	h3:focus {
		outline: none;
	}
	/* Categories read as folders: bold title with an arrow, items indented below. */
	.group {
		margin: 0.25rem 0 0;
		border-bottom: 1px solid var(--line);
	}
	.group-toggle {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		width: 100%;
		min-height: var(--tap);
		padding: 0.25rem 0;
		border: none;
		background: transparent;
		font: inherit;
		font-weight: 700;
		color: var(--text);
		text-align: start;
	}
	.chevron {
		width: 1rem;
		height: 1rem;
		flex: none;
		fill: currentColor;
		transition: transform 0.15s;
	}
	.chevron.open {
		transform: rotate(90deg);
	}
	.items {
		padding-inline-start: 1rem;
	}
</style>
