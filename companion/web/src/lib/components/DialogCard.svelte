<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { reasonText } from '$lib/text.js';
	import type { Dialog, DialogChoice } from '$lib/gen/ts4access/v1/game_pb.js';
	import { announce } from './Announcer.svelte';

	// One window the game is waiting on, answered from the phone like its own
	// buttons: a question, a list to pick from (recipes, Sims...), a text box.
	let { dialog }: { dialog: Dialog } = $props();

	const OK = 10001;
	const CANCEL = 10002;

	let picked = $state<string[]>([]);
	let text = $state('');
	let busy = $state(false);
	let lastId = '';
	$effect(() => {
		// A new window: start from the game's own selection and default text.
		if (dialog.id !== lastId) {
			lastId = dialog.id;
			picked = dialog.rows.filter((r) => r.selected).map((r) => r.id);
			text = dialog.textDefault;
		}
	});

	const single = $derived(dialog.maxSelectable <= 1);
	const heading = $derived(dialog.title || dialog.text || m.dialog_loading());

	function label(c: DialogChoice): string {
		if (c.text) return c.text;
		if (c.id === OK) return m.dialog_ok();
		if (c.id === CANCEL) return m.dialog_cancel();
		return m.dialog_button({ n: dialog.choices.indexOf(c) + 1 });
	}

	function toggle(id: string, on: boolean) {
		if (single) picked = on ? [id] : [];
		else if (on) picked = picked.length < dialog.maxSelectable ? [...picked, id] : picked;
		else picked = picked.filter((x) => x !== id);
	}

	// Confirming a picker needs enough rows chosen; Cancel always works.
	const blocked = (c: DialogChoice) =>
		c.id !== CANCEL && dialog.rows.length > 0 && picked.length < Math.max(1, dialog.minSelectable);

	async function answer(c: DialogChoice) {
		busy = true;
		const out = await run(() =>
			client.respondDialog({
				dialogId: dialog.id,
				choiceId: c.id,
				rowIds: c.id === CANCEL ? [] : picked,
				text: dialog.textInput && c.id !== CANCEL ? text : ''
			})
		);
		busy = false;
		announce(out.ok ? m.done() : m.failed({ reason: reasonText(out.reason) }));
	}
</script>

<section class="card stack ask" aria-labelledby={`dialog-${dialog.id}`}>
	<h2 id={`dialog-${dialog.id}`}>{m.dialog_title()}: {heading}</h2>
	{#if dialog.title && dialog.text}<p>{dialog.text}</p>{/if}

	{#if dialog.rows.length}
		<fieldset class="stack">
			<legend>
				{single ? m.dialog_pick_one() : m.dialog_pick_many({ max: dialog.maxSelectable })}
			</legend>
			{#each dialog.rows as row (row.id)}
				<label class="pick" class:off={!row.enabled}>
					<input
						type={single ? 'radio' : 'checkbox'}
						name={`dialog-${dialog.id}`}
						checked={picked.includes(row.id)}
						disabled={!row.enabled || busy}
						onchange={(e) => toggle(row.id, e.currentTarget.checked)}
					/>
					<span>
						{#if row.enabled}
							{row.name}
						{:else if row.reason}
							{m.unavailable_because({ name: row.name, reason: row.reason })}
						{:else}
							{m.unavailable({ name: row.name })}
						{/if}
						{#if row.description}<span class="muted"> · {row.description}</span>{/if}
					</span>
				</label>
			{/each}
		</fieldset>
	{/if}

	{#if dialog.textInput}
		<label class="stack">
			<span>{m.dialog_answer()}</span>
			<input type="text" bind:value={text} disabled={busy} autocomplete="off" />
		</label>
	{/if}

	<div class="buttons">
		{#each dialog.choices as c (c.id)}
			<button type="button" disabled={busy || blocked(c)} onclick={() => answer(c)}>
				{label(c)}{#if c.reason}<span class="muted"> · {c.reason}</span>{/if}
			</button>
		{/each}
	</div>
</section>

<style>
	.ask {
		border: 3px solid var(--accent);
	}
	fieldset {
		border: none;
		margin: 0;
		padding: 0;
	}
	legend {
		padding: 0;
		margin-bottom: 0.5rem;
	}
	.pick {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		min-height: var(--tap);
	}
	.pick input {
		width: 1.5rem;
		height: 1.5rem;
		flex: none;
		accent-color: var(--accent);
	}
	.off {
		opacity: 0.7;
	}
	input[type='text'] {
		font: inherit;
		min-height: var(--tap);
		border-radius: var(--r);
		border: 2px solid var(--line);
		background: var(--surface-2);
		color: var(--text);
		padding: 0 0.75rem;
	}
	.buttons {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
		gap: 0.5rem;
	}
</style>
