<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { live } from '$lib/live.svelte.js';
	import { groupOf, prefs, socialShown } from '$lib/prefs.svelte.js';
	import { eventText } from '$lib/text.js';

	/** Show only the newest events (0 = all). */
	let { limit = 0 }: { limit?: number } = $props();

	const feed = $derived.by(() => {
		const list = live.feed.filter(
			(e) => (prefs.showSpoken || groupOf(e.kind) !== 'spoken') && socialShown(e)
		);
		return limit > 0 ? list.slice(0, limit) : list;
	});
	const time = (at: bigint) =>
		new Date(Number(at)).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
</script>

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

<style>
	.feed li {
		padding: 0.6rem 0;
		border-bottom: 1px solid var(--line);
	}
	.feed .muted {
		margin-inline-end: 0.5rem;
	}
</style>
