<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { live } from '$lib/live.svelte.js';

	const message = $derived.by(() => {
		switch (live.status) {
			case 'connecting':
				return m.status_connecting();
			case 'retrying':
				return m.status_retrying();
			case 'live':
				if (!live.gameConnected) return m.status_no_game();
				if (!live.state?.inGame) return m.status_menu();
				return '';
			default:
				return '';
		}
	});
</script>

<p class="banner" role="status" hidden={!message}>{message}</p>

<style>
	.banner {
		background: var(--surface-2);
		border: 2px solid var(--warn);
		border-radius: var(--r);
		padding: 0.75rem 1rem;
	}
</style>
