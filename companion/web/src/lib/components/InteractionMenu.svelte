<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { reasonText } from '$lib/text.js';
	import type { Interaction, Target } from '$lib/gen/ts4access/v1/game_pb.js';
	import { menuDone, menuOpened } from '$lib/speed.svelte.js';
	import { announce } from './Announcer.svelte';

	// The game's pie menu for one Sim or object, as the game shows it: the
	// top level first, sub-menus (Cook…, Serve Lunch…) opened one at a time,
	// with Back to return.
	let { target, onBack }: { target: Target; onBack: () => void } = $props();

	let menu = $state<Interaction[] | null>(null);
	let busy = $state(false);
	let heading: HTMLHeadingElement | undefined = $state();
	let subHeading: HTMLHeadingElement | undefined = $state();
	let category = $state<string | null>(null);

	// Reload only when the target itself changes (its name may be refreshed).
	const targetId = $derived(target.id);
	$effect(() => {
		void load(targetId);
	});

	// Optional: pause while choosing, resume once done (Settings).
	void menuOpened().then((paused) => paused && announce(m.paused_to_choose()));
	onDestroy(() => void menuDone());

	async function load(id: string) {
		menu = null;
		category = null;
		await tick();
		heading?.focus();
		// Names come from the game's string tables: ask again while some are pending.
		for (let attempt = 0; attempt < 5; attempt++) {
			try {
				const res = await client.getInteractions({ targetId: id });
				if (id !== target.id) return;
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
			await new Promise((r) => setTimeout(r, 800));
		}
	}

	// Top level in the game's order: direct actions, and one entry per sub-menu.
	type Entry = { kind: 'action'; item: Interaction } | { kind: 'category'; label: string; count: number };
	const entries = $derived.by(() => {
		const out: Entry[] = [];
		for (const i of menu ?? []) {
			if (!i.category) {
				out.push({ kind: 'action', item: i });
				continue;
			}
			const existing = out.find((e) => e.kind === 'category' && e.label === i.category);
			if (existing && existing.kind === 'category') existing.count++;
			else out.push({ kind: 'category', label: i.category, count: 1 });
		}
		return out;
	});
	const inCategory = $derived((menu ?? []).filter((i) => i.category === category));

	async function openCategory(label: string) {
		category = label;
		await tick();
		subHeading?.focus();
	}
	async function closeCategory() {
		const label = category;
		category = null;
		await tick();
		document.getElementById(categoryId(label ?? ''))?.focus();
	}
	const categoryId = (label: string) => `menu-cat-${label.replace(/[^\p{L}\p{N}]+/gu, '-')}`;

	function label(i: Interaction): string {
		if (i.enabled) return i.name;
		return i.reason
			? m.unavailable_because({ name: i.name, reason: i.reason })
			: m.unavailable({ name: i.name });
	}

	async function choose(i: Interaction) {
		if (busy) return;
		if (!i.enabled) {
			announce(label(i));
			return;
		}
		busy = true;
		const out = await run(() => client.runInteraction({ targetId: target.id, interactionId: i.id }));
		busy = false;
		announce(out.ok ? m.queued({ name: i.name }) : m.failed({ reason: reasonText(out.reason) }));
		if (out.ok && (await menuDone())) announce(m.resumed());
	}

	const title = $derived(target.group === 'PHONE' ? m.group_PHONE() : target.name);
</script>

<section class="card stack" aria-labelledby="menu-title">
	{#if category === null}
		<button type="button" onclick={onBack}>{m.back_to_list()}</button>
		<h2 id="menu-title" tabindex="-1" bind:this={heading}>{m.interactions_of({ name: title })}</h2>
		{#if menu === null}
			<p class="muted" role="status">{m.loading()}</p>
		{:else if menu.length === 0}
			<p class="muted">{m.no_interactions()}</p>
		{:else}
			<ul class="stack">
				{#each entries as e (e.kind === 'action' ? e.item.id : 'cat:' + e.label)}
					<li>
						{#if e.kind === 'action'}
							<button
								type="button"
								class="row"
								aria-disabled={!e.item.enabled}
								onclick={() => choose(e.item)}
							>
								{label(e.item)}
							</button>
						{:else}
							<button
								type="button"
								class="row submenu"
								id={categoryId(e.label)}
								onclick={() => openCategory(e.label)}
							>
								<span
									>{e.count === 1
										? m.submenu_label_one({ name: e.label })
										: m.submenu_label({ name: e.label, count: e.count })}</span
								>
								<svg class="chevron" viewBox="0 0 16 16" aria-hidden="true" focusable="false"
									><path d="M6 3l5 5-5 5z" /></svg
								>
							</button>
						{/if}
					</li>
				{/each}
			</ul>
			<p class="muted">{m.dialog_on_mac()}</p>
		{/if}
	{:else}
		<button type="button" onclick={closeCategory}>{m.back_to_menu({ name: title })}</button>
		<h2 id="menu-title" tabindex="-1" bind:this={subHeading}>{category}</h2>
		<ul class="stack">
			{#each inCategory as i (i.id)}
				<li>
					<button type="button" class="row" aria-disabled={!i.enabled} onclick={() => choose(i)}
						>{label(i)}</button
					>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	.row {
		width: 100%;
		text-align: start;
	}
	.submenu {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 0.75rem;
	}
	.chevron {
		width: 1rem;
		height: 1rem;
		flex: none;
		fill: currentColor;
	}
	button[aria-disabled='true'] {
		opacity: 0.7;
		border-style: dashed;
	}
	h2:focus {
		outline: none;
	}
</style>
