<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { linkInAddressBar } from '$lib/session.js';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { live } from '$lib/live.svelte.js';
	import { groupOf, prefs } from '$lib/prefs.svelte.js';
	import { eventText } from '$lib/text.js';
	import Announcer, { announce } from '$lib/components/Announcer.svelte';
	import StatusBanner from '$lib/components/StatusBanner.svelte';
	import LiveView from '$lib/components/LiveView.svelte';
	import SimsView from '$lib/components/SimsView.svelte';
	import PlayView from '$lib/components/PlayView.svelte';
	import SettingsView from '$lib/components/SettingsView.svelte';
	import Feed from '$lib/components/Feed.svelte';

	type Tab = 'live' | 'play' | 'sims' | 'history' | 'settings';
	const TABS: { id: Tab; label: () => string }[] = [
		{ id: 'live', label: m.tab_live },
		{ id: 'play', label: m.tab_play },
		{ id: 'sims', label: m.tab_sims },
		{ id: 'history', label: m.tab_history },
		{ id: 'settings', label: m.tab_settings }
	];
	const TAB_KEY = 'ts4access.tab';

	function savedTab(): Tab {
		try {
			const t = localStorage.getItem(TAB_KEY);
			if (TABS.some((x) => x.id === t)) return t as Tab;
		} catch {
			// default tab
		}
		return 'live';
	}

	let tab = $state<Tab>(savedTab());
	let heading: HTMLHeadingElement | undefined = $state();
	const title = $derived(TABS.find((t) => t.id === tab)!.label());

	async function show(next: Tab) {
		tab = next;
		try {
			localStorage.setItem(TAB_KEY, next);
		} catch {
			// not remembered
		}
		await tick();
		// Move VoiceOver to the new screen's title.
		heading?.focus();
	}

	onMount(() => {
		// Drop the pairing token from the address bar once the app has started.
		// SvelteKit's own state object is kept, so its router is unaffected.
		setTimeout(() => {
			if (linkInAddressBar()) history.replaceState(history.state, '', location.pathname + location.search);
		});
		live.onEvent = (e) => {
			const group = groupOf(e.kind);
			if (group && group !== 'spoken' && prefs.announce[group]) announce(eventText(e));
		};
		live.start();
		return () => live.stop();
	});
</script>

<svelte:head><title>{m.page_title()}</title></svelte:head>

<Announcer />

{#if live.status === 'unpaired' || live.status === 'rejected'}
	<main class="page stack">
		<h1>{m.pair_title()}</h1>
		<p>{live.status === 'rejected' ? m.pair_rejected() : m.pair_help()}</p>
	</main>
{:else}
	<!-- First in reading order so VoiceOver reaches the tabs at once; drawn at the bottom. -->
	<nav aria-label={m.nav_label()}>
		{#each TABS as t (t.id)}
			<button type="button" aria-current={tab === t.id ? 'page' : undefined} onclick={() => show(t.id)}>
				{t.label()}
			</button>
		{/each}
	</nav>

	<main class="page stack" aria-labelledby="title">
		<h1 id="title" tabindex="-1" bind:this={heading}>{title}</h1>
		<StatusBanner />
		{#if tab === 'live'}
			<LiveView onSeeAll={() => show('history')} />
		{:else if tab === 'play'}
			<PlayView />
		{:else if tab === 'sims'}
			<SimsView />
		{:else if tab === 'history'}
			<Feed />
		{:else}
			<SettingsView />
		{/if}
	</main>
{/if}

<style>
	.page {
		max-width: 40rem;
		margin: 0 auto;
		padding: calc(env(safe-area-inset-top) + 1rem) var(--gap) calc(env(safe-area-inset-bottom) + 6rem);
	}
	h1:focus {
		outline: none;
	}
	nav {
		position: fixed;
		inset: auto 0 0 0;
		display: grid;
		grid-template-columns: repeat(5, 1fr);
		gap: 0.25rem;
		padding: 0.5rem 0.5rem calc(env(safe-area-inset-bottom) + 0.5rem);
		background: var(--surface);
		border-top: 1px solid var(--line);
	}
	nav button {
		border: none;
		background: transparent;
		font-weight: 600;
		padding: 0.5rem 0.25rem;
	}
	nav button[aria-current='page'] {
		background: var(--accent);
		color: var(--accent-ink);
	}
</style>
