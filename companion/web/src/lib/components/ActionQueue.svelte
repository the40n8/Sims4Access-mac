<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { reasonText, simName } from '$lib/text.js';
	import { ActionState, type Action, type Sim } from '$lib/gen/simsight/v1/game_pb.js';
	import { announce } from './Announcer.svelte';

	// What a Sim is doing and will do next, each with Cancel: on Live and Play.
	let { sim }: { sim: Sim } = $props();

	let busy = $state(false);
	let showAll = $state(false);
	const listId = $derived(`queue-${sim.id}`);

	async function cancel(action: Action) {
		if (busy) return;
		busy = true;
		const out = await run(() => client.cancelAction({ simId: sim.id, actionId: action.id }));
		busy = false;
		announce(out.ok ? m.done() : m.failed({ reason: reasonText(out.reason) }));
	}
</script>

<section class="card stack" aria-labelledby={`${listId}-title`}>
	<h2 id={`${listId}-title`}>{m.queue_title({ name: simName(sim) })}</h2>
	{#if sim.actions.length === 0}
		<p class="muted">{m.actions_none()}</p>
	{:else}
		<!-- The running action and the next few; the rest on demand, so what
		     follows stays one swipe away. -->
		<ul class="stack" id={listId}>
			{#each showAll ? sim.actions : sim.actions.slice(0, 4) as action, n (action.id)}
				<li class="queued">
					<span id={`action-${action.id}`}>
						{n + 1}. {action.name}
						<span class="muted">
							· {action.state === ActionState.RUNNING ? m.action_running() : m.action_queued()}
						</span>
					</span>
					{#if action.cancelable}
						<button
							type="button"
							aria-describedby={`action-${action.id}`}
							aria-disabled={busy}
							onclick={() => cancel(action)}
						>
							{m.cancel_action({ action: action.name })}
						</button>
					{/if}
				</li>
			{/each}
		</ul>
		{#if sim.actions.length > 4}
			<button
				type="button"
				aria-expanded={showAll}
				aria-controls={listId}
				onclick={() => (showAll = !showAll)}
			>
				{showAll ? m.show_fewer() : m.show_all_actions({ count: sim.actions.length })}
			</button>
		{/if}
	{/if}
</section>

<style>
	.queued {
		display: grid;
		gap: 0.4rem;
	}
</style>
