<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { live } from '$lib/live.svelte.js';
	import { groupOf, prefs } from '$lib/prefs.svelte.js';
	import { eventText, money, simSummary, speedName } from '$lib/text.js';
	import { Speed } from '$lib/gen/ts4access/v1/game_pb.js';

	const game = $derived(live.state);
	const feed = $derived(live.feed.filter((e) => prefs.showSpoken || groupOf(e.kind) !== 'spoken'));
	const time = (at: bigint) =>
		new Date(Number(at)).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
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
	<h2 id="feed">{m.feed_title()}</h2>
	{#if feed.length === 0}
		<p class="muted">{m.feed_empty()}</p>
	{:else}
		<ol class="feed">
			{#each feed as e (e.seq)}
				<li>
					<span class="muted num">{time(e.at)}</span>
					{eventText(e)}
				</li>
			{/each}
		</ol>
	{/if}
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
	.feed li {
		padding: 0.6rem 0;
		border-bottom: 1px solid var(--line);
	}
	.feed .muted {
		margin-inline-end: 0.5rem;
	}
</style>
