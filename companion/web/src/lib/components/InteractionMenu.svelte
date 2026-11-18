<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { reasonText } from '$lib/text.js';
	import type { Interaction, Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import { menuDone, menuOpened } from '$lib/speed.svelte.js';
	import { announce } from './Announcer.svelte';

	// The game's pie menu for one Sim or object.
	let { target, onBack }: { target: Target; onBack: () => void } = $props();

	let menu = $state<Interaction[] | null>(null);
	let busy = $state(false);
	let heading: HTMLHeadingElement | undefined = $state();

	$effect(() => {
		void load(target.id);
	});

	// Optional: pause while choosing, resume once done (Settings).
	void menuOpened().then((paused) => paused && announce(m.paused_to_choose()));
	onDestroy(() => void menuDone());

	async function load(targetId: string) {
		menu = null;
		await tick();
		heading?.focus();
		// Names come from the game's string tables: ask again while some are pending.
		for (let attempt = 0; attempt < 4; attempt++) {
			try {
				const res = await client.getInteractions({ targetId });
				if (targetId !== target.id) return;
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

	const groups = $derived.by(() => {
		const out: { label: string; items: Interaction[] }[] = [];
		for (const i of menu ?? []) {
			let g = out.find((x) => x.label === i.category);
			if (!g) out.push((g = { label: i.category, items: [] }));
			g.items.push(i);
		}
		// Top-level rows first, as in the game's pie menu.
		return out.sort((a, b) => (a.label === '' ? -1 : b.label === '' ? 1 : 0));
	});

	async function choose(i: Interaction) {
		if (busy) return;
		if (!i.enabled) {
			announce(
				i.reason ? m.unavailable_because({ name: i.name, reason: i.reason }) : m.unavailable({ name: i.name })
			);
			return;
		}
		busy = true;
		const out = await run(() => client.runInteraction({ targetId: target.id, interactionId: i.id }));
		busy = false;
		announce(out.ok ? m.queued({ name: i.name }) : m.failed({ reason: reasonText(out.reason) }));
		if (out.ok && (await menuDone())) announce(m.resumed());
	}
</script>

<section class="card stack" aria-labelledby="menu-title">
	<button type="button" onclick={onBack}>{m.back_to_list()}</button>
	<h2 id="menu-title" tabindex="-1" bind:this={heading}>
		{m.interactions_of({ name: target.group === 'PHONE' ? m.group_PHONE() : target.name })}
	</h2>
	{#if menu === null}
		<p class="muted" role="status">{m.loading()}</p>
	{:else if menu.length === 0}
		<p class="muted">{m.no_interactions()}</p>
	{:else}
		{#each groups as g (g.label)}
			{#if g.label}<h3>{g.label}</h3>{/if}
			<ul class="stack">
				{#each g.items as i (i.id)}
					<li>
						<button type="button" class="row" aria-disabled={!i.enabled} onclick={() => choose(i)}>
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

<style>
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
</style>
