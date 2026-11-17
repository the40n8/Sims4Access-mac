<script lang="ts">
	import { m } from '$lib/i18n/paraglide/messages.js';
	import { LANGUAGES, changeLocale, i18n, isLocale } from '$lib/i18n/locale.svelte.js';
	import { GROUPS, prefs, savePrefs, type Group } from '$lib/prefs.svelte.js';
	import { forgetToken } from '$lib/session.js';

	const labels: Record<Group, () => string> = {
		mood: m.kind_mood,
		moodlet: m.kind_moodlet,
		need: m.kind_need,
		action: m.kind_action,
		game: m.kind_game
	};

	function unpair() {
		forgetToken();
		location.reload();
	}
</script>

<section class="card stack" aria-labelledby="lang">
	<h2 id="lang">{m.language()}</h2>
	<select
		aria-labelledby="lang"
		value={i18n.locale}
		onchange={(e) => isLocale(e.currentTarget.value) && changeLocale(e.currentTarget.value)}
	>
		{#each LANGUAGES as l (l.locale)}
			<option value={l.locale} lang={l.locale}>{l.name}</option>
		{/each}
	</select>
</section>

<fieldset class="card stack">
	<legend><h2>{m.announce_title()}</h2></legend>
	<p class="muted">{m.announce_help()}</p>
	{#each GROUPS as group (group)}
		<label class="check">
			<input type="checkbox" bind:checked={prefs.announce[group]} onchange={savePrefs} />
			{labels[group]()}
		</label>
	{/each}
	<label class="check">
		<input type="checkbox" bind:checked={prefs.showSpoken} onchange={savePrefs} />
		{m.show_spoken()}
	</label>
</fieldset>

<section class="card">
	<button type="button" onclick={unpair}>{m.unpair()}</button>
</section>

<style>
	select {
		font: inherit;
		min-height: var(--tap);
		border-radius: var(--r);
		padding: 0 0.75rem;
		background: var(--surface-2);
		color: var(--text);
		border: 2px solid var(--line);
	}
	fieldset {
		margin: 0;
	}
	legend {
		padding: 0;
		margin-bottom: 0.5rem;
	}
	.check {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		min-height: var(--tap);
	}
	.check input {
		width: 1.5rem;
		height: 1.5rem;
		accent-color: var(--accent);
	}
</style>
