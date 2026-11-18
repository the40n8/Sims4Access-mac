<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { i18n } from '$lib/i18n/locale.svelte.js';
	import { HELP } from '$lib/help.js';

	// Plain sections with headings: VoiceOver's rotor (Headings) jumps between them.
	const sections = $derived(HELP[i18n.locale] ?? HELP.en);

	async function jump(id: string) {
		await tick();
		document.getElementById(`help-${id}`)?.focus();
	}
</script>

<nav class="card stack" aria-labelledby="help-contents">
	<h2 id="help-contents">{m.help_contents()}</h2>
	<ul class="stack">
		{#each sections as s (s.id)}
			<li><button type="button" class="row" onclick={() => jump(s.id)}>{s.title}</button></li>
		{/each}
	</ul>
</nav>

{#each sections as s (s.id)}
	<section class="card stack" aria-labelledby={`help-${s.id}`}>
		<h2 id={`help-${s.id}`} tabindex="-1">{s.title}</h2>
		{#each s.body as paragraph, i (i)}
			<p>{paragraph}</p>
		{/each}
	</section>
{/each}

<style>
	.row {
		width: 100%;
		text-align: start;
	}
	h2:focus {
		outline: none;
	}
</style>
