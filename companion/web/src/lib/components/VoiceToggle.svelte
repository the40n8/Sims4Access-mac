<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { client, run } from '$lib/client.js';
	import { computer, live } from '$lib/live.svelte.js';
	import { reasonText } from '$lib/text.js';
	import { announce } from './Announcer.svelte';

	// One tap to silence the Mac (or bring it back), like Option+Command+S there.
	const narration = $derived(live.state?.narration);
	const on = $derived(narration?.macVoice ?? false);
	let busy = $state(false);

	async function toggle() {
		if (busy || !live.gameConnected) return;
		busy = true;
		const out = await run(() =>
			client.setNarration({ narration: { macVoice: !on, simActions: narration?.simActions ?? true } })
		);
		busy = false;
		if (!out.ok) announce(m.failed({ reason: reasonText(out.reason) }));
	}
</script>

<div class="voice">
	<!-- A switch: VoiceOver says "Mac voice, switch, on" ("PC voice" on Windows). -->
	<button
		type="button"
		role="switch"
		aria-checked={on}
		aria-disabled={busy || !live.gameConnected}
		onclick={toggle}
	>
		<span>{m.narration_title(computer())}</span>
		<span class="state" aria-hidden="true">{on ? m.voice_on() : m.voice_off()}</span>
	</button>
	{#if computer().computer === 'Mac'}<p class="muted">{m.voice_shortcut()}</p>{/if}
</div>

<style>
	.voice {
		display: grid;
		gap: 0.25rem;
	}
	.voice button {
		width: 100%;
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}
	.voice button[aria-checked='true'] .state {
		font-weight: 800;
	}
	.voice p {
		margin: 0;
		font-size: 0.9em;
	}
</style>
