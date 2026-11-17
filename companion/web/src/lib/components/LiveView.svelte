<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { live } from '$lib/live.svelte.js';
	import { money, simSummary, speedName } from '$lib/text.js';
	import { Speed } from '$lib/gen/ts4access/v1/game_pb.js';
	import Feed from './Feed.svelte';

	let { onSeeAll }: { onSeeAll: () => void } = $props();
	const game = $derived(live.state);
</script>

{#if game?.inGame}
	<section class="card stack" aria-labelledby="household">
		<h2 id="household">{m.household_title({ name: game.householdName })}</h2>
		<p>
			{m.clock_line({
				day: game.clock?.day ?? '',
				time: game.clock?.time ?? '',
				speed: speedName(game.clock?.speed ?? Speed.NORMAL)
			})}
			<span class="num">{m.funds({ amount: money(game.funds) })}</span>
		</p>
		<ul class="stack">
			{#each game.sims as sim (sim.id)}
				<li class:active={sim.active}>
					{simSummary(sim)}
					{#if sim.active}<span class="badge">{m.active_badge()}</span>{/if}
				</li>
			{/each}
		</ul>
	</section>
{/if}

<section class="stack" aria-labelledby="feed">
	<h2 id="feed">{m.recent_title()}</h2>
	<Feed limit={5} />
	<button type="button" onclick={onSeeAll}>{m.see_all()}</button>
</section>

<style>
	.active {
		font-weight: 700;
	}
	.badge {
		display: inline-block;
		margin-inline-start: 0.5rem;
		padding: 0 0.5rem;
		border-radius: 999px;
		background: var(--accent);
		color: var(--accent-ink);
		font-size: 0.85em;
	}
</style>
