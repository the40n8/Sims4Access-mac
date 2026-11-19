<script lang="ts">
	import '@fontsource-variable/atkinson-hyperlegible-next';
	import '../app.css';
	import { onMount } from 'svelte';
	import { applyDocumentLocale, i18n } from '$lib/i18n/locale.svelte.js';
	let { children } = $props();
	onMount(applyDocumentLocale);
</script>

{#key i18n.locale}
	{@render children()}
{/key}
