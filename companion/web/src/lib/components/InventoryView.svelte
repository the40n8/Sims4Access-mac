<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { reasonText } from '$lib/text.js';
	import type { InventoryItem, Target } from '$lib/gen/simsight/v1/game_pb.js';
	import { objectName } from '$lib/targets.js';
	import { announce } from './Announcer.svelte';

	// What the controlled Sim carries, each item with the places that take it
	// (the game's drag and drop to a fridge or a bookcase, in one tap).
	let { owner, onBack }: { owner: string; onBack: () => void } = $props();

	let items = $state<InventoryItem[] | null>(null);
	let busy = $state(false);
	let heading: HTMLHeadingElement | undefined = $state();

	async function load() {
		try {
			const res = await client.listInventory({});
			items = res.outcome?.ok ? res.items : [];
			if (!res.outcome?.ok) announce(m.failed({ reason: reasonText(res.outcome?.reason ?? '') }));
		} catch {
			items = [];
		}
	}

	onMount(async () => {
		await tick();
		heading?.focus();
		await load();
	});

	async function move(item: InventoryItem, to: Target) {
		if (busy || !item.item) return;
		busy = true;
		const out = await run(() => client.moveItem({ itemId: item.item!.id, destinationId: to.id }));
		busy = false;
		announce(
			out.ok
				? m.moved_to({ name: item.item.name, place: to.name })
				: m.failed({ reason: reasonText(out.reason) })
		);
		if (out.ok) await load();
	}
</script>

<section class="card stack" aria-labelledby="inventory-title">
	<button type="button" onclick={onBack}>{m.back_to_list()}</button>
	<h2 id="inventory-title" tabindex="-1" bind:this={heading}>{m.inventory_of({ name: owner })}</h2>
	{#if items === null}
		<p class="muted" role="status">{m.loading()}</p>
	{:else if items.length === 0}
		<p class="muted">{m.inventory_empty()}</p>
	{:else}
		<ul class="stack">
			{#each items as it, n (it.item?.id ?? n)}
				<li class="stack item">
					<p id={`inv-${it.item?.id}`}>
						{it.item ? objectName(it.item) : ''}{#if it.count > 1}{` ×${it.count}`}{/if}
					</p>
					{#if it.destinations.length === 0}
						<p class="muted">{m.inventory_nowhere()}</p>
					{:else}
						<div class="buttons">
							{#each it.destinations as d (d.id)}
								<button
									type="button"
									aria-describedby={`inv-${it.item?.id}`}
									aria-disabled={busy}
									onclick={() => move(it, d)}
								>
									{m.put_in({ place: d.name, meters: d.distance })}
								</button>
							{/each}
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	.item {
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--line);
	}
	.item:last-child {
		border-bottom: none;
	}
	.item p {
		margin: 0;
	}
	.item > p:first-child {
		font-weight: 600;
	}
	.buttons {
		display: grid;
		gap: 0.5rem;
	}
	h2:focus {
		outline: none;
	}
</style>
