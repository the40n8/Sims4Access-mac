<script lang="ts">
	import { tick } from 'svelte';
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { i18n } from '$lib/i18n/locale.svelte.js';
	import { AI_PROMPT, HELP } from '$lib/help.js';
	import { announce } from './Announcer.svelte';

	// Plain sections with headings: VoiceOver's rotor (Headings) jumps between them.
	const sections = $derived(HELP[i18n.locale] ?? HELP.en);

	const prompt = $derived(AI_PROMPT[i18n.locale] ?? AI_PROMPT.en);

	async function copyPrompt() {
		let ok: boolean;
		try {
			await navigator.clipboard.writeText(prompt);
			ok = true;
		} catch {
			// Plain http (local Wi-Fi link): the old way still works.
			const area = document.createElement('textarea');
			area.value = prompt;
			area.setAttribute('readonly', '');
			area.style.position = 'fixed';
			area.style.opacity = '0';
			document.body.append(area);
			area.select();
			ok = document.execCommand('copy');
			area.remove();
		}
		announce(ok ? m.prompt_copied() : m.prompt_copy_failed());
	}

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
		{#if s.id === 'ai'}
			<p class="prompt">{prompt}</p>
			<div class="ai">
				<button type="button" onclick={copyPrompt}>{m.copy_prompt()}</button>
				<a
					class="button"
					href={`https://chatgpt.com/?q=${encodeURIComponent(prompt)}`}
					target="_blank"
					rel="noopener">{m.open_chatgpt()}</a
				>
			</div>
		{/if}
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
	.prompt {
		padding: 0.75rem;
		border-radius: var(--r);
		background: var(--surface-2);
	}
	.ai {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 0.5rem;
	}
	.ai .button {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: var(--tap);
		border-radius: var(--r);
		border: 2px solid var(--line);
		background: var(--surface-2);
		color: var(--text);
		font-weight: 600;
		text-decoration: none;
		text-align: center;
	}
</style>
