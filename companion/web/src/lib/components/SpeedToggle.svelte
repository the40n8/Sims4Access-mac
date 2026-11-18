<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { live } from '$lib/live.svelte.js';
	import { reasonText, speedName } from '$lib/text.js';
	import { isPaused, togglePause } from '$lib/speed.svelte.js';
	import { announce } from './Announcer.svelte';

	// Always first on screen: one button to pause or resume, then the clock.
	// VoiceOver: "Pause the game, button. Normal speed, Friday 9:12 PM".
	const clock = $derived(live.state?.clock);
	const paused = $derived(isPaused());
	let busy = $state(false);

	async function toggle() {
		// Stays enabled while busy: disabling a focused button loses VoiceOver's place.
		if (busy || !live.gameConnected) return;
		busy = true;
		const wasPaused = paused;
		const out = await togglePause();
		busy = false;
		announce(
			out.ok ? (wasPaused ? m.resumed() : m.paused_now()) : m.failed({ reason: reasonText(out.reason) })
		);
	}
</script>

{#if live.state?.inGame && clock}
	<div class="bar">
		<button
			type="button"
			class:paused
			aria-describedby="clock-status"
			aria-disabled={busy || !live.gameConnected}
			onclick={toggle}
		>
			<!-- Drawn icons (not emoji), hidden from VoiceOver: the words say it. -->
			<svg class="icon" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
				{#if paused}
					<path d="M4 2.5v11l9-5.5z" />
				{:else}
					<rect x="3.5" y="2.5" width="3" height="11" rx="0.5" /><rect
						x="9.5"
						y="2.5"
						width="3"
						height="11"
						rx="0.5"
					/>
				{/if}
			</svg>
			{paused ? m.resume_game() : m.pause_game()}
		</button>
		<span id="clock-status" class="muted">{speedName(clock.speed)}, {clock.day} {clock.time}</span>
	</div>
{/if}

<style>
	.bar {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.icon {
		width: 1em;
		height: 1em;
		fill: currentColor;
		vertical-align: -0.125em;
		margin-inline-end: 0.35em;
	}
	.bar button {
		flex: 1;
		font-weight: 700;
	}
	.bar button.paused {
		background: var(--accent);
		color: var(--accent-ink);
	}
	.bar > span {
		flex: none;
		max-width: 45%;
		text-align: end;
		font-size: 0.9em;
	}
</style>
