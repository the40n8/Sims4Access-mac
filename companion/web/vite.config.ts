import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		// `pnpm dev`: proxy the API to a bridge on this Mac
		// (game running, or `simsight-bridge --demo`).
		host: true,
		// The phone reaches the Mac by its Bonjour name (mac/dev.sh web).
		allowedHosts: ['.local'],
		proxy: { '/simsight.v1.CompanionService': 'http://127.0.0.1:47800' }
	}
});
