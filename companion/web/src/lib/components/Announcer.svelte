<script lang="ts" module>
	// One polite live region for the whole app: VoiceOver reads each message
	// after what it is currently saying, without moving focus.
	let queue: string[] = [];
	let timer: ReturnType<typeof setTimeout> | null = null;
	let setText: ((t: string) => void) | null = null;

	function flush() {
		timer = null;
		const next = queue.shift();
		if (next === undefined || !setText) return;
		setText('');
		// Clearing first makes a repeated message count as a change.
		setTimeout(() => setText?.(next), 50);
		if (queue.length) timer = setTimeout(flush, 1200);
	}

	/** Queue a message for VoiceOver. Bursts are spaced out, and capped. */
	export function announce(text: string) {
		queue.push(text);
		if (queue.length > 5) queue = queue.slice(-5);
		if (!timer) timer = setTimeout(flush, 100);
	}
</script>

<script lang="ts">
	import { onDestroy } from 'svelte';
	let text = $state('');
	setText = (t) => (text = t);
	onDestroy(() => (setText = null));
</script>

<div class="visually-hidden" aria-live="polite" aria-atomic="true">{text}</div>
